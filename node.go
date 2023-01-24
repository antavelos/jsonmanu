package jsonmanu

import (
	"errors"
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

// Full array JSONPath pattern.
// Example: `books[*]`
const ARRAY_TOKEN_PATTERN = `^(?P<node>\w+)\[\*\]$`

// Indexed array JSONPath pattern.
// Examples:
// - `books[2]`
// - `books[1,2]`
const INDEXED_ARRAY_TOKEN_PATTERN = `^(?P<node>\w+)\[(?P<indices>( *\d+,? *)+)\]$`

// Sliced array JSONPath pattern.
// Examples:
// - `books[:2]`
// - `books[3:]`
// - `books[1:2]`
const SLICED_ARRAY_TOKEN_PATTERN = `^(?P<node>\w+)\[(?P<start>\-?\d*):(?P<end>\-?\d*)\]$`

// Filtered array JSONPath pattern.
// Examples:
// - `books[?(@.isbn)]`
// - `books[?(@.price<10)]`
const FILTERED_ARRAY_TOKEN_PATTERN = `^(?P<node>\w+)\[\?\(@\.(?P<key>\w+)((?P<op>(\<|\>|!|=|(<=)|(>=))?)(?P<value>[\w\d]*))?\)\]$`

// Simple JSON node pattern.
const SIMPLE_NODE_TOKEN_PATTERN = `^(?P<node>(\w*|\*))$`

// Interface to be implemented by all Node like structs for name retrieval.
type NamedNode interface {
	GetName() string
}

// Interface to be implemented by all Node like structs for retrieving and updating JSON node data.
type NodeDataAccessor interface {
	NamedNode

	// Retrieves a value out of a given map according to the rules of the node called.
	Get(any) any

	// Updates a given map according to the rules of the node called.
	Put(any, any) error
}

type NodePutError string

func (err NodePutError) Error() string {
	return fmt.Sprintf("NodePutError error: %s", string(err))
}

// Represents a simple JSON object node or leaf.
type Node struct {
	Name string
}

// Parent object for array like JSON nodes.
type ArrayNode struct {
	Node
}

// Represents an indexed array node i.e. `books[2]`.
type ArrayIndexedNode struct {
	ArrayNode

	// Holds the indices
	Indices []int
}

// Represents an sliced array node i.e. `books[2:4]`.
type ArraySlicedNode struct {
	ArrayNode

	// The start index
	Start int

	// The end index
	End int
}

// Represents a filtered array node i.e. `books[?(@.isbn)]`.
type ArrayFilteredNode struct {
	ArrayNode

	// The property to filter with.
	Key string

	// The comparison oparator. Can be one of '=', '!', '<', '=<', '=>', '>'.
	Op string

	// The value to compare with.
	Value any
}

// ----
// Node
// ----

// validateSourceData ensures that the provided data can be used by the node for retrieval or update.
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

// Get returns the value of the provided map data with key same as the name of the node.
func (node Node) Get(sourceData any) any {
	_, err := node.validateSourceData(sourceData)
	if err != nil {
		return nil
	}

	return sourceData.(map[string]any)[node.Name]
}

// Put updates the value of the provided map data with key same as the name of the node.
func (node Node) Put(sourceData any, value any) error {
	if _, err := node.validateSourceData(sourceData); err != nil {
		return err
	}

	sourceData.(map[string]any)[node.Name] = value

	return nil
}

// GetName returns the name of the node.
func (node Node) GetName() string { return node.Name }

// ---------
// ArrayNode
// ---------

// validateSourceData ensures that the provided data can be used by the array node for retrieval or update
func (node ArrayNode) validateSourceData(sourceData any) (any, error) {
	data, err := node.Node.validateSourceData(sourceData)
	if err != nil {
		return nil, err
	}

	if !isSlice(data) {
		return nil, NodePutError(fmt.Sprintf("Value of key '%v' is not an array: %#v", node.Node.Name, sourceData))
	}

	return data, nil
}

// ----------------
// ArrayIndexedNode
// ----------------

// Get returns the value of the provided map data with key same as the name of the node.
// The underlying value must be a slice and the returned value will be the slice
// containing only the values of the `sourceData` defined by the indices of the node.
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

// Put updates the value of the provided map data with key same as the name of the node.
// The underlying value must be a slice and the new value will apply on the slice
// defined by the indices of the node.
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

// GetName returns the name of the node.
func (node ArrayIndexedNode) GetName() string { return node.Node.Name }

// ----------------
// ArraySlicedNode
// ----------------

// Get returns the value of the provided map data with key same as the name of the node.
// The underlying value must be a slice and the returned value will be the subslice
// defined by the Start and End values of the node.
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

// Put updates the value of the provided map data with key same as the name of the node.
// The underlying value must be a slice and the new value will apply on the slice
// defined by the indices of the node.
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

// GetName returns the name of the node.
func (node ArraySlicedNode) GetName() string { return node.Node.Name }

// -----------------
// ArrayFilteredNode
// -----------------

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

// isString returns whether the valus is of type string or not
func isString(value any) bool {
	switch value.(type) {
	case string:
		return true
	}
	return false
}

// assertCondition asserts the condition defined by the values and the operator.
// The operator can be one of `=`, `!“, `<`, `>`, `<=`, `>=`
// First a comparison will be attempted between floats (if applicable) and then between strings (if applicable)
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

// Get returns the value of the provided map data with key same as the name of the node.
// The underlying value must be a slice and the returned value will be the subslice
// that satisfies the condition defived by the key, value and operator of the node.
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

// Put updates the value of the provided map data with key same as the name of the node.
// The underlying value must be a slice and the returned value will be the subslice
// that satisfies the condition defived by the key, value and operator of the node.
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

// GetName returns the name of the node.
func (node ArrayFilteredNode) GetName() string { return node.Node.Name }

// ----------
// node utils
// ----------

type MatchDictionary map[string]string

// getMatchDictionary returns a map of placeholders and their values found in a string given a pattern with placeholders in it.
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

// nodeFromToken checks one by one the existing token patterns and returns an appropriate node data accessor.
func nodeFromToken(token string) NodeDataAccessor {
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

// isArrayNode returns whether the node is of array type or not.
func isArrayNode(node NodeDataAccessor) bool {
	switch node.(type) {
	case ArrayIndexedNode, ArraySlicedNode, ArrayFilteredNode:
		return true
	}
	return false
}

// walkNodes iterates through a slice of nodes and at the same time descends in the given `data` map object replacing it with the new value.
// The value held in data at the end of the itaration will be returned.
func walkNodes(data any, nodes []NodeDataAccessor) any {
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
