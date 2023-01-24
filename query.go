package jsonmanu

import (
	"fmt"
	"strings"
)

type JsonPathError string

func (jpe JsonPathError) Error() string {
	return fmt.Sprintf("JSONPath error: %s", string(jpe))
}

// splitJsonPath splits a string based on a `.` separator. However, the string is supposed to be a JSONPath so
// the case of `@.` shall be specially handled.
func splitJsonPath(path string) []string {
	tempPath := strings.Replace(path, "@.", "@:", -1)

	tokens := strings.Split(tempPath, ".")

	for i := 0; i < len(tokens); i++ {
		if strings.Contains(tokens[i], "@:") {
			tokens[i] = strings.Replace(tokens[i], "@:", "@.", -1)
		}
	}

	return tokens
}

// parse translates a provided JSONPath to an array of node data accessors that can be used to retrieve values from or update a given map.
func parse(path string) ([]nodeDataAccessor, error) {
	if !strings.HasPrefix(path, "$.") {
		return nil, JsonPathError("JSONPath should start with '$.'")
	}

	if strings.HasSuffix(path, ".") {
		return nil, JsonPathError("JSONPath should not end with '.'")
	}

	jsonPathSubNodes := splitJsonPath(path)

	var nodes []nodeDataAccessor
	for i, jsonPathSubNode := range jsonPathSubNodes[1:] {
		node := nodeFromJsonPathSubNode(jsonPathSubNode)
		if node == nil {
			return nil, JsonPathError(fmt.Sprintf("Couldn't parse JSONPath substring %v: '%v'", i, jsonPathSubNode))
		}

		nodes = append(nodes, node)
	}

	return nodes, nil
}

// Get retrieves a value out of the given map as it is described in the provided JSONPath.
func Get(data any, path string) (any, error) {
	nodes, err := parse(path)
	if err != nil {
		return nil, err
	}

	data, err = walkNodes(data, nodes)
	if err != nil {
		return nil, err
	}

	return data, nil
}

// Put updates a given map as it is described in the provided JSONPath.
func Put(data any, path string, value any) error {
	nodes, err := parse(path)
	if err != nil {
		return err
	}

	nodesCount := len(nodes)

	// handle reccursive descent in case the last node is a leaf node and its previous one
	// is not known, i.e. "$..price"
	if nodes[nodesCount-2].getName() == "" && !isArrayNode(nodes[nodesCount-1]) {
		return mapPutDeep(data, nodes[nodesCount-1].getName(), value)
	}

	allButLastNodes, lastNode := nodes[:nodesCount-1], nodes[nodesCount-1]

	data, err = walkNodes(data, allButLastNodes)
	if err != nil {
		return err
	}

	if isSlice(data) {
		for _, item := range data.([]any) {
			if err := lastNode.put(item, value); err != nil {
				return err
			}
		}
		return nil
	}

	err = lastNode.put(data, value)

	return err
}
