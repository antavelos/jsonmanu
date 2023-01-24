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
const JSON_PATH_ARRAY_NODE_PATTERN = `^(?P<node>\w+)\[\*\]$`

// Indexed array JSONPath pattern.
// Examples:
// - `books[2]`
// - `books[1,2]`
const JSON_PATH_INDEXED_ARRAY_NODE_PATTERN = `^(?P<node>\w+)\[(?P<indices>( *\d+,? *)+)\]$`

// Sliced array JSONPath pattern.
// Examples:
// - `books[:2]`
// - `books[3:]`
// - `books[1:2]`
const JSON_PATH_SLICED_ARRAY_NODE_PATTERN = `^(?P<node>\w+)\[(?P<start>\-?\d*):(?P<end>\-?\d*)\]$`

// Filtered array JSONPath pattern.
// Examples:
// - `books[?(@.isbn)]`
// - `books[?(@.price<10)]`
const JSON_PATH_FILTERED_ARRAY_NODE_PATTERN = `^(?P<node>\w+)\[\?\(@\.(?P<key>\w+)\s*((?P<op>(\<|\>|(!=)|={2}|(<=)|(>=))?)\s*(?P<value>[\w\d]*))?\)\]$`

// Simple JSON node pattern.
const JSON_PATH_SIMPLE_NODE_PATTERN = `^(?P<node>(\w*|\*))$`

// Interface to be implemented by all node like structs for name retrieval.
type namedNode interface {
	getName() string
}

// Interface to be implemented by all node like structs for retrieving and updating JSON node data.
type nodeDataAccessor interface {
	namedNode

	// Retrieves a value out of a given map according to the rules of the node called.
	get(any) (any, error)

	// Updates a given map according to the rules of the node called.
	put(any, any) error
}

type SourceValidationError string

func (err SourceValidationError) Error() string {
	return fmt.Sprintf("SourceValidationError error: %s", string(err))
}

// Represents a simple JSON object node or leaf.
type node struct {
	name string
}

// Parent object for array like JSON nodes.
type arrayNode struct {
	node
}

// Represents an indexed array node i.e. `books[2]`.
type arrayIndexedNode struct {
	arrayNode

	// Holds the indices
	indices []int
}

// Represents an sliced array node i.e. `books[2:4]`.
type arraySlicedNode struct {
	arrayNode

	// The start index
	start int

	// The end index
	end int
}

// Represents a filtered array node i.e. `books[?(@.isbn)]`.
type arrayFilteredNode struct {
	arrayNode

	// The property to filter with.
	key string

	// The comparison oparator. Can be one of '=', '!', '<', '=<', '=>', '>'.
	op string

	// The value to compare with.
	value any
}

// ----
// node
// ----

// validateSource ensures that the provided data can be used by the node for retrieval or update.
func (n node) validateSource(source any) error {
	if !isMap(source) {
		return SourceValidationError(fmt.Sprintf("Source data is not a map: %#v", source))
	}

	if !mapHasKey(source, n.name) {
		return SourceValidationError(fmt.Sprintf("key '%v' not found", n.name))
	}

	return nil
}

// get returns the value of the provided map data with key same as the name of the node.
func (n node) get(source any) (any, error) {
	if err := n.validateSource(source); err != nil {
		return nil, err
	}

	return source.(map[string]any)[n.name], nil
}

// put updates the value of the provided map data with key same as the name of the node.
func (n node) put(source any, value any) error {
	if err := n.validateSource(source); err != nil {
		return err
	}

	source.(map[string]any)[n.name] = value

	return nil
}

// getName returns the name of the node.
func (n node) getName() string { return n.name }

// ---------
// arrayNode
// ---------

// validateSource ensures that the provided data can be used by the array node for retrieval or update
func (n arrayNode) validateSource(source any) error {
	if err := n.node.validateSource(source); err != nil {
		return err
	}

	value, _ := source.(map[string]any)[n.name]

	if !isSlice(value) {
		return SourceValidationError(fmt.Sprintf("Value of key '%v' is not an array: %#v", n.node.name, source))
	}

	return nil
}

// ----------------
// arrayIndexedNode
// ----------------

// get returns the value of the provided map data with key same as the name of the node.
// The underlying value must be a slice and the returned value will be the slice
// containing only the values of the `source` defined by the indices of the node.
func (n arrayIndexedNode) get(source any) (any, error) {
	if err := n.validateSource(source); err != nil {
		return nil, err
	}

	value := source.(map[string]any)[n.name]

	if len(n.indices) == 0 {
		return value, nil
	}

	var result []any
	for _, i := range n.indices {
		if i < 0 || i >= len(value.([]any)) {
			continue
		}
		result = append(result, value.([]any)[i])
	}

	return result, nil
}

// put updates the value of the provided map data with key same as the name of the node.
// The underlying value must be a slice and the new value will apply on the slice
// defined by the indices of the node.
func (n arrayIndexedNode) put(source any, newVal any) error {
	if err := n.validateSource(source); err != nil {
		return err
	}

	value := source.(map[string]any)[n.name]

	for _, i := range n.indices {
		if i < 0 || i >= len(value.([]any)) {
			continue
		}
		value.([]any)[i] = newVal
	}

	return nil
}

// getName returns the name of the node.
func (n arrayIndexedNode) getName() string { return n.node.name }

// ----------------
// arraySlicedNode
// ----------------

// get returns the value of the provided map data with key same as the name of the node.
// The underlying value must be a slice and the returned value will be the subslice
// defined by the start and end values of the node.
func (n arraySlicedNode) get(source any) (any, error) {
	if err := n.validateSource(source); err != nil {
		return nil, err
	}

	value := source.(map[string]any)[n.name]

	if n.start != 0 && n.end != 0 {
		return value.([]any)[n.start:n.end], nil
	}

	if n.start != 0 && n.end == 0 {
		return value.([]any)[n.start:], nil
	}

	if n.start == 0 && n.end != 0 {
		return value.([]any)[:n.end], nil
	}

	return source, nil
}

// put updates the value of the provided map data with key same as the name of the n.
// The underlying value must be a slice and the new value will apply on the slice
// defined by the indices of the n.
func (n arraySlicedNode) put(source any, newVal any) error {
	if err := n.validateSource(source); err != nil {
		return err
	}

	value := source.(map[string]any)[n.name]

	if n.start != 0 && n.end != 0 {
		value = value.([]any)[n.start:n.end]
	} else if n.start != 0 && n.end == 0 {
		value = value.([]any)[n.start:]
	} else if n.start == 0 && n.end != 0 {
		value = value.([]any)[:n.end]
	} else {
		return nil
	}

	for i, _ := range value.([]any) {
		value.([]any)[i] = newVal
	}

	return nil
}

// getName returns the name of the n.
func (n arraySlicedNode) getName() string { return n.node.name }

// -----------------
// arrayFilteredNode
// -----------------

// toFloat converts any number like value or any string number to float64.
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
	case "==":
		if areFloats {
			return fval1 == fval2
		}
		if isString(val1) && isString(val2) {
			return val1.(string) == val2.(string)
		}
	case "=!":
		if areFloats {
			return fval1 != fval2
		}
		if isString(val1) && isString(val2) {
			return val1.(string) != val2.(string)
		}
	}

	return false
}

// get returns the value of the provided map data with key same as the name of the n.
// The underlying value must be a slice and the returned value will be the subslice
// that satisfies the condition defived by the key, value and operator of the n.
func (n arrayFilteredNode) get(source any) (any, error) {
	if err := n.validateSource(source); err != nil {
		return nil, err
	}

	value := source.(map[string]any)[n.name]

	var filteredVal []any
	for _, item := range value.([]any) {
		value, ok := item.(map[string]any)[n.key]
		if !ok {
			continue
		}

		if len(n.op) == 0 || n.value == nil || assertCondition(value, n.value, n.op) {
			filteredVal = append(filteredVal, item)
		}
	}

	return filteredVal, nil
}

// put updates the value of the provided map data with key same as the name of the n.
// The underlying value must be a slice and the returned value will be the subslice
// that satisfies the condition defived by the key, value and operator of the n.
func (n arrayFilteredNode) put(source any, newVal any) error {
	if err := n.validateSource(source); err != nil {
		return err
	}

	value := source.(map[string]any)[n.name]

	for _, item := range value.([]any) {
		currValue, ok := item.(map[string]any)[n.key]
		if !ok {
			continue
		}

		if len(n.op) == 0 || n.value == nil || assertCondition(currValue, n.value, n.op) {
			item.(map[string]any)[n.key] = newVal
		}
	}

	return nil
}

// getName returns the name of the n.
func (n arrayFilteredNode) getName() string { return n.node.name }

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

// nodeFromJsonPathSubNode checks one by one the existing JSONPath patterns and returns an appropriate node data accessor.
func nodeFromJsonPathSubNode(jsonPathSubNode string) nodeDataAccessor {
	var dict map[string]string

	dict = getMatchDictionary(JSON_PATH_ARRAY_NODE_PATTERN, jsonPathSubNode)
	if len(dict) > 0 {
		return arrayIndexedNode{
			arrayNode: arrayNode{
				node: node{
					name: dict["node"],
				},
			},
		}
	}

	dict = getMatchDictionary(JSON_PATH_INDEXED_ARRAY_NODE_PATTERN, jsonPathSubNode)
	if len(dict) > 0 {
		node := arrayIndexedNode{
			arrayNode: arrayNode{
				node: node{
					name: dict["node"],
				},
			},
		}
		indices := strings.Split(dict["indices"], ",")
		for _, index := range indices {
			indexInt, _ := strconv.Atoi(strings.TrimSpace(index))
			node.indices = append(node.indices, indexInt)
		}

		return node
	}

	dict = getMatchDictionary(JSON_PATH_SLICED_ARRAY_NODE_PATTERN, jsonPathSubNode)
	if len(dict) > 0 {
		node := arraySlicedNode{
			arrayNode: arrayNode{
				node: node{
					name: dict["node"],
				},
			},
		}
		node.start, _ = strconv.Atoi(dict["start"])
		node.end, _ = strconv.Atoi(dict["end"])

		return node
	}

	dict = getMatchDictionary(JSON_PATH_FILTERED_ARRAY_NODE_PATTERN, jsonPathSubNode)
	if len(dict) > 0 {
		return arrayFilteredNode{
			arrayNode: arrayNode{
				node: node{
					name: dict["node"],
				},
			},
			key:   dict["key"],
			op:    dict["op"],
			value: dict["value"],
		}
	}

	dict = getMatchDictionary(JSON_PATH_SIMPLE_NODE_PATTERN, jsonPathSubNode)
	if len(dict) > 0 {
		return node{
			name: dict["node"],
		}
	}

	return nil
}

// isArrayNode returns whether the node is of array type or not.
func isArrayNode(n nodeDataAccessor) bool {
	switch n.(type) {
	case arrayIndexedNode, arraySlicedNode, arrayFilteredNode:
		return true
	}
	return false
}

// walkNodes iterates through a slice of nodes and at the same time descends in the given `data` map object replacing it with the new value.
// The value held in data at the end of the itaration will be returned.
func walkNodes(data any, nodes []nodeDataAccessor) (any, error) {
	var err error

	withReccursiveDescent := false
	for _, n := range nodes {
		if n.getName() == "*" {
			continue
		}

		if n.getName() == "" {
			withReccursiveDescent = true
			continue
		}

		if isSlice(data) {
			var items []any
			for _, item := range data.([]any) {
				value, err := n.get(item)
				if err != nil {
					return nil, err
				}
				items = append(items, value)
			}
			data = items
			continue
		}

		if withReccursiveDescent {
			data = mapGetDeepFlattened(data, n.getName())
			if isArrayNode(n) {
				dataWithkey := map[string]any{n.getName(): data}
				data, err = n.get(dataWithkey)
				if err != nil {
					return nil, err
				}
			}
			continue
		}

		data, err = n.get(data)
		if err != nil {
			return nil, err
		}
	}

	return data, nil
}
