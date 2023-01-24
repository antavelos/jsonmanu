package jsonmanu

import (
	"fmt"
	"strings"
)

type JsonPathError string

func (jpe JsonPathError) Error() string {
	return fmt.Sprintf("JSONPath error: %s", string(jpe))
}

type JsonData map[string]any

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
func parse(path string) ([]NodeDataAccessor, error) {
	if !strings.HasPrefix(path, "$.") {
		return nil, JsonPathError("JsonPath should start with '$.'")
	}

	if strings.HasSuffix(path, ".") {
		return nil, JsonPathError("JsonPath should not end with '.'")
	}

	tokens := splitJsonPath(path)

	var nodes []NodeDataAccessor
	for i, token := range tokens[1:] {
		node := nodeFromToken(token)
		if node == nil {
			return nil, JsonPathError(fmt.Sprintf("Couldn't parse token %v: '%v'", i, token))
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

	data = walkNodes(data, nodes)

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
	if nodes[nodesCount-2].GetName() == "" && !isArrayNode(nodes[nodesCount-1]) {
		return mapPutDeep(data, nodes[nodesCount-1].GetName(), value)
	}

	allButLastNodes, lastNode := nodes[:nodesCount-1], nodes[nodesCount-1]

	data = walkNodes(data, allButLastNodes)

	if isSlice(data) {
		for _, item := range data.([]any) {
			if err := lastNode.Put(item, value); err != nil {
				return err
			}
		}
		return nil
	}

	return lastNode.Put(data, value)
}
