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

func parse(path string) ([]NodeDataManager, error) {
	if !strings.HasPrefix(path, "$.") {
		return nil, JsonPathError("JsonPath should start with '$.'")
	}

	if strings.HasSuffix(path, ".") {
		return nil, JsonPathError("JsonPath should not end with '.'")
	}

	tokens := splitJsonPath(path)

	var nodes []NodeDataManager
	for i, token := range tokens[1:] {
		node := nodeFromToken(token)
		if node == nil {
			return nil, JsonPathError(fmt.Sprintf("Couldn't parse token %v: '%v'", i, token))
		}

		nodes = append(nodes, node)
	}

	return nodes, nil
}

func get(data any, nodes []NodeDataManager) any {
	withReccursiveDescent := false
	for _, node := range nodes {
		if node.GetName() == "*" {
			continue
		}

		if node.GetName() == "" {
			withReccursiveDescent = true
			continue
		}

		if isSlice(data) {
			var idata []any
			for _, item := range data.([]any) {
				idata = append(idata, node.Get(item))
			}
			data = idata
			continue
		}

		if withReccursiveDescent {
			data = mapGetDeepFlattened(data, node.GetName())
			if isArrayNode(node) {
				dataWithKey := map[string]any{node.GetName(): data}
				data = node.Get(dataWithKey)
			}
			continue
		}

		data = node.Get(data)
	}

	return data
}

func Get(data any, path string) (any, error) {
	nodes, err := parse(path)
	if err != nil {
		return nil, err
	}

	data = get(data, nodes)

	return data, nil
}

func Put(data any, path string, value any) error {
	nodes, err := parse(path)
	if err != nil {
		return err
	}

	nodesCount := len(nodes)
	// handle reccursive descent Put() in case the previous of the last node is not known
	// i.e. $..price
	if nodesCount == 2 &&
		nodes[0].GetName() == "" &&
		!isArrayNode(nodes[1]) {

		return mapPutDeep(data, nodes[1].GetName(), value)
	}

	allButLastNodes, lastNode := nodes[:nodesCount-1], nodes[nodesCount-1]

	data = get(data, allButLastNodes)

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
