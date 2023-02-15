package jsonmanu

import (
	"fmt"
	"strings"
)

func jsonPathHasReccursiveDescent(path string) bool {
	return strings.Contains(path, "..")
}

// splitJsonPath splits a string based on a `.` separator. However, the string is supposed to be a JSONPath so
// the case of `@.` shall be specially handled.
func splitJsonPath(jsonPath string) []string {
	tempJsonPath := strings.Replace(jsonPath, "@.", "@:", -1)

	tokens := strings.Split(tempJsonPath, ".")

	for i := 0; i < len(tokens); i++ {
		if strings.Contains(tokens[i], "@:") {
			tokens[i] = strings.Replace(tokens[i], "@:", "@.", -1)
		}
	}

	return tokens
}

// parseJsonPath translates a provided JSONPath to an array of node data accessors that can be used to retrieve values from or update a given map.
func parseJsonPath(jsonPath string) ([]nodeDataAccessor, error) {
	if !strings.HasPrefix(jsonPath, "$.") {
		return nil, fmt.Errorf("JSONPath should start with '$.'")
	}

	if strings.HasSuffix(jsonPath, ".") {
		return nil, fmt.Errorf("JSONPath should not end with '.'")
	}

	jsonPathSubNodes := splitJsonPath(jsonPath)

	var nodes []nodeDataAccessor
	for i, jsonPathSubNode := range jsonPathSubNodes[1:] {
		node := nodeFromJsonPathSubNode(jsonPathSubNode)
		if node == nil {
			return nil, fmt.Errorf("Couldn't parse JSONPath substring %v: '%v'", i, jsonPathSubNode)
		}

		nodes = append(nodes, node)
	}

	return nodes, nil
}

// Get retrieves a value out of a given map or a slice of maps as it is described in the provided JSONPath.
//
// The `data` must not be nil.
//
// It returns the retrieved data if everything goes well. Otherwise nil along with the relevant error.
func Get(data map[string]any, jsonPath string) (any, error) {
	nodes, err := parseJsonPath(jsonPath)
	if err != nil {
		return nil, err
	}

	result, err := walkNodes(data, nodes)
	if err != nil {
		return nil, err
	}

	return result, nil
}

// ensureDataStrunctureFromNodes creates the map tree structure in case in is not present so it can be safely used later by Put.
// The data argument is any because the function runs reccursively and besides a map it can be of any type.
func ensureDataStrunctureFromNodes(data any, nodes []nodeDataAccessor) {

	if len(nodes) == 0 {
		return
	}

	if isSlice(data) {
		for item := range iterAny(data, nil) {
			ensureDataStrunctureFromNodes(item, nodes[1:])
		}
	} else if isMap(data) {
		firstNodeName := nodes[0].getName()

		val, ok := data.(map[string]any)[firstNodeName]
		if !ok || val == nil {
			data.(map[string]any)[firstNodeName] = make(map[string]any)
			val, _ = data.(map[string]any)[firstNodeName]
		}

		ensureDataStrunctureFromNodes(val, nodes[1:])
	}
}

// Put updates the branch(es) of a map or a slice of maps as it is described in the provided JSONPath with a new value.
//
// The `data` must not be nil. The changes will apply in place.
//
// If the path described in the `jsonPath` does not exist then it will be created on the fly. Attibutes referred within an array condition will be ignored.
//
// An error will be returned should anything goes wrong.
func Put(data map[string]any, jsonPath string, value any) error {
	nodes, err := parseJsonPath(jsonPath)
	if err != nil {
		return err
	}

	if !jsonPathHasReccursiveDescent(jsonPath) && data != nil {
		ensureDataStrunctureFromNodes(data, nodes)
	}

	nodesCount := len(nodes)

	if nodesCount >= 2 && nodes[nodesCount-2].getName() == "" && !isArrayNode(nodes[nodesCount-1]) {
		return mapPutDeep(data, nodes[nodesCount-1].getName(), value)
	}

	allButLastNodes, lastNode := nodes[:nodesCount-1], nodes[nodesCount-1]

	walkedData, err := walkNodes(data, allButLastNodes)
	if err != nil {
		switch err.(dataValidationError).errorType {
		case dataValidationErrorNotMap, dataValidationErrorValueNotArray:
			return err
		case dataValidationErrorKeyNotFound:
			walkedData = data
		}
	}

	if isSlice(walkedData) {
		for _, item := range walkedData.([]any) {
			if err := lastNode.put(item.(map[string]any), value); err != nil {
				return err
			}
		}
		return nil
	}

	err = lastNode.put(walkedData.(map[string]any), value)

	return err
}
