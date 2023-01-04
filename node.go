package jsonmanu

import (
	"errors"
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

// full array i.e. books[*]
const ARRAY_TOKEN_PATTERN = `^(?P<node>\w+)\[\*\]$`

// indexed arrays i.e. books[1,2] books[2]
const INDEXED_ARRAY_TOKEN_PATTERN = `^(?P<node>\w+)\[(?P<indices>( *\d+,? *)+)\]$`

// sliced array i.e. books[:2], books[3:], books[1:2]
const SLICED_ARRAY_TOKEN_PATTERN = `^(?P<node>\w+)\[(?P<start>\-?\d*):(?P<end>\-?\d*)\]$`

// filtered array i.e. books[?(@.isbn)], books[?(@.price<10)]
const FILTERED_ARRAY_TOKEN_PATTERN = `^(?P<node>\w+)\[\?\(@\.(?P<key>\w+)((?P<op>(\<|\>|!|=|(<=)|(>=))?)(?P<value>[\w\d]*))?\)\]$`

// empty or alphanumeric named nodes (object nodes)
const SIMPLE_NODE_TOKEN_PATTERN = `^(?P<node>(\w*|\*))$`

type NodeDataManager interface {
	Get(any) any
	Put(any, any) error
	GetName() string
}

type NodePutError string

func (err NodePutError) Error() string {
	return fmt.Sprintf("NodePutError error: %s", string(err))
}

type Node struct {
	Name string
}

type ArrayNode struct {
	Node
}

type ArrayIndexedNode struct {
	ArrayNode
	Indices []int
}

type ArraySlicedNode struct {
	ArrayNode
	Start int
	End   int
}

type ArrayFilteredNode struct {
	ArrayNode
	Key   string
	Op    string
	Value any
}

// To be used to validate the provided unstructured map data in Get and Put methods
func (node Node) validateSourceData(sourceData any) (any, error) {
	if !isMap(sourceData) {
		return nil, NodePutError(fmt.Sprintf("Source data is not a map: %#v", sourceData))
	}

	data, ok := sourceData.(map[string]any)[node.Name]
	if !ok {
		return nil, NodePutError(fmt.Sprintf("Key '%v' not found", node.Name))
	}

	return data, nil
}

func (node Node) Get(sourceData any) any {
	data, err := node.validateSourceData(sourceData)
	if err != nil {
		return nil
	}

	return data
}

func (node Node) Put(sourceData any, value any) error {
	if _, err := node.validateSourceData(sourceData); err != nil {
		return err
	}

	sourceData.(map[string]any)[node.Name] = value

	return nil
}

func (node Node) GetName() string { return node.Name }

func (node ArrayNode) validateSourceData(sourceData any) (any, error) {
	data, err := node.Node.validateSourceData(sourceData)
	if err != nil {
		return nil, err
	}

	if !isSlice(data) {
		return nil, NodePutError(fmt.Sprintf("Value of key '%v' is not a slice: %#v", node.Node.Name, sourceData))
	}

	return data, nil
}

func (node ArrayIndexedNode) Get(sourceData any) any {
	data, err := node.validateSourceData(sourceData)
	if err != nil {
		return nil
	}

	if len(node.Indices) == 0 {
		return data
	}

	var result []any
	for _, i := range node.Indices {
		if i < 0 || i >= len(data.([]any)) {
			continue
		}
		result = append(result, data.([]any)[i])
	}

	return result
}

func (node ArrayIndexedNode) Put(sourceData any, value any) error {
	data, err := node.validateSourceData(sourceData)
	if err != nil {
		return err
	}

	for _, i := range node.Indices {
		if i < 0 || i >= len(data.([]any)) {
			continue
		}
		data.([]any)[i] = value
	}

	return nil
}

func (node ArrayIndexedNode) GetName() string { return node.Node.Name }

func (node ArraySlicedNode) Get(sourceData any) any {
	data, err := node.validateSourceData(sourceData)
	if err != nil {
		return nil
	}

	if node.Start != 0 && node.End != 0 {
		return data.([]any)[node.Start:node.End]
	}

	if node.Start != 0 && node.End == 0 {
		return data.([]any)[node.Start:]
	}

	if node.Start == 0 && node.End != 0 {
		return data.([]any)[:node.End]
	}

	return sourceData
}

func (node ArraySlicedNode) Put(sourceData any, value any) error {
	data, err := node.validateSourceData(sourceData)
	if err != nil {
		return err
	}

	if node.Start != 0 && node.End != 0 {
		data = data.([]any)[node.Start:node.End]
	} else if node.Start != 0 && node.End == 0 {
		data = data.([]any)[node.Start:]
	} else if node.Start == 0 && node.End != 0 {
		data = data.([]any)[:node.End]
	} else {
		return nil
	}

	for i, _ := range data.([]any) {
		data.([]any)[i] = value
	}

	return nil
}

func (node ArraySlicedNode) GetName() string { return node.Node.Name }

func toFloat64(value any) (float64, error) {
	switch v := value.(type) {
	case int:
		return float64(v), nil
	case int8:
		return float64(v), nil
	case int16:
		return float64(v), nil
	case int32:
		return float64(v), nil
	case int64:
		return float64(v), nil
	case uint8:
		return float64(v), nil
	case uint16:
		return float64(v), nil
	case uint32:
		return float64(v), nil
	case uint64:
		return float64(v), nil
	case float32:
		return float64(v), nil
	case float64:
		return float64(v), nil
	case string:
		fv, err := strconv.ParseFloat(v, 1)
		if err == nil {
			return fv, nil
		}
	}
	return 0, errors.New("Can't convert to float64")
}

func isString(value any) bool {
	switch value.(type) {
	case string:
		return true
	}
	return false
}
func assertCondition(val1 any, val2 any, op string) bool {
	fval1, err1 := toFloat64(val1)
	fval2, err2 := toFloat64(val2)
	areFloats := err1 == nil && err2 == nil

	switch op {
	case "<":
		if areFloats {
			return fval1 < fval2
		}
		if isString(val1) && isString(val2) {
			return val1.(string) < val2.(string)
		}
	case ">":
		if areFloats {
			return fval1 > fval2
		}
		if isString(val1) && isString(val2) {
			return val1.(string) < val2.(string)
		}
	case "<=":
		if areFloats {
			return fval1 <= fval2
		}
		if isString(val1) && isString(val2) {
			return val1.(string) <= val2.(string)
		}
	case ">=":
		if areFloats {
			return fval1 >= fval2
		}
		if isString(val1) && isString(val2) {
			return val1.(string) >= val2.(string)
		}
	case "=":
		if areFloats {
			return fval1 == fval2
		}
		if isString(val1) && isString(val2) {
			return val1.(string) == val2.(string)
		}
	case "!":
		if areFloats {
			return fval1 != fval2
		}
		if isString(val1) && isString(val2) {
			return val1.(string) != val2.(string)
		}
	}

	return false
}

func (node ArrayFilteredNode) Get(sourceData any) any {
	data, err := node.validateSourceData(sourceData)
	if err != nil {
		return nil
	}

	var filtered []any
	for _, item := range data.([]any) {
		value, ok := item.(map[string]any)[node.Key]
		if !ok {
			continue
		}

		if len(node.Op) == 0 || node.Value == nil || assertCondition(value, node.Value, node.Op) {
			filtered = append(filtered, item)
		}
	}

	return filtered
}

func (node ArrayFilteredNode) Put(sourceData any, value any) error {
	data, err := node.validateSourceData(sourceData)
	if err != nil {
		return err
	}

	for _, item := range data.([]any) {
		currValue, ok := item.(map[string]any)[node.Key]
		if !ok {
			continue
		}

		if len(node.Op) == 0 || node.Value == nil || assertCondition(currValue, node.Value, node.Op) {
			item.(map[string]any)[node.Key] = value
		}
	}

	return nil
}

func (node ArrayFilteredNode) GetName() string { return node.Node.Name }

type MatchDictionary map[string]string

func getMatchDictionary(patt string, s string) (dict MatchDictionary) {
	defer func() {
		if r := recover(); r != nil {
			fmt.Println("Recovered from: ", r)
		}
	}()

	dict = map[string]string{}

	re := regexp.MustCompile(patt)

	subexpNames := re.SubexpNames()
	if len(subexpNames) == 0 {
		return
	}

	submatches := re.FindStringSubmatch(s)
	if len(submatches) == 0 {
		return
	}
	for _, subexpName := range subexpNames {
		if subexpName == "" {
			continue
		}
		dict[subexpName] = submatches[re.SubexpIndex(subexpName)]
	}

	return
}

func nodeFromToken(token string) NodeDataManager {
	var dict map[string]string

	dict = getMatchDictionary(ARRAY_TOKEN_PATTERN, token)
	if len(dict) > 0 {
		return ArrayIndexedNode{
			ArrayNode: ArrayNode{
				Node: Node{
					Name: dict["node"],
				},
			},
		}
	}

	dict = getMatchDictionary(INDEXED_ARRAY_TOKEN_PATTERN, token)
	if len(dict) > 0 {
		node := ArrayIndexedNode{
			ArrayNode: ArrayNode{
				Node: Node{
					Name: dict["node"],
				},
			},
		}
		indices := strings.Split(dict["indices"], ",")
		for _, index := range indices {
			indexInt, _ := strconv.Atoi(strings.TrimSpace(index))
			node.Indices = append(node.Indices, indexInt)
		}

		return node
	}

	dict = getMatchDictionary(SLICED_ARRAY_TOKEN_PATTERN, token)
	if len(dict) > 0 {
		node := ArraySlicedNode{
			ArrayNode: ArrayNode{
				Node: Node{
					Name: dict["node"],
				},
			},
		}
		node.Start, _ = strconv.Atoi(dict["start"])
		node.End, _ = strconv.Atoi(dict["end"])

		return node
	}

	dict = getMatchDictionary(FILTERED_ARRAY_TOKEN_PATTERN, token)
	if len(dict) > 0 {
		return ArrayFilteredNode{
			ArrayNode: ArrayNode{
				Node: Node{
					Name: dict["node"],
				},
			},
			Key:   dict["key"],
			Op:    dict["op"],
			Value: dict["value"],
		}
	}

	dict = getMatchDictionary(SIMPLE_NODE_TOKEN_PATTERN, token)
	if len(dict) > 0 {
		return Node{
			Name: dict["node"],
		}
	}

	return nil
}

func isArrayNode(node NodeDataManager) bool {
	switch node.(type) {
	case ArrayIndexedNode, ArraySlicedNode, ArrayFilteredNode:
		return true
	}
	return false
}
