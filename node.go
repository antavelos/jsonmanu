package jsonmanu

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

// Full array JSONPath pattern.
// Example: `books[*]`
const jsonPathArrayNodePattern = `^(?P<node>\w+)\[\*\]$`

// Indexed array JSONPath pattern.
// Examples:
// - `books[2]`
// - `books[1,2]`
const jsonPathIndexedArrayNodePattern = `^(?P<node>\w+)\[(?P<indices>( *\d+,? *)+)\]$`

// Sliced array JSONPath pattern.
// Examples:
// - `books[:2]`
// - `books[3:]`
// - `books[1:2]`
const jsonPathSlicedArrayNodePattern = `^(?P<node>\w+)\[(?P<start>\-?\d*):(?P<end>\-?\d*)\]$`

// Filtered array JSONPath pattern.
// Examples:
// - `books[?(@.isbn)]`
// - `books[?(@.price<10)]`
const jsonPathFilteredArrayNodePattern = `^(?P<node>\w+)\[\?\(@\.(?P<key>\w+)\s*((?P<op>(\<|\>|(!=)|={2}|(<=)|(>=))?)\s*(?P<value>[\w\d]*))?\)\]$`

// Simple JSON node pattern.
const jsonPathSimpleNodePattern = `^(?P<node>(\w*|\*))$`

// Interface to be implemented by all node like structs for name retrieval.
type namedNode interface {
	getName() string
}

// Interface to be implemented by all node like structs for retrieving and updating JSON node data.
type nodeDataAccessor interface {
	namedNode

	// Retrieves a value out of a given map according to the rules of the node called.
	get(map[string]any) (any, error)

	// Updates a given map according to the rules of the node called.
	put(map[string]any, any) error
}

// Represents a simple JSON object node or leaf.
type node struct {
	name string
}

// Represents an indexed array node i.e. `books[2]`.
type arrayIndexedNode struct {
	node

	// Holds the indices
	indices []int
}

// Represents an sliced array node i.e. `books[2:4]`.
type arraySlicedNode struct {
	node

	// The start index
	start int

	// The end index
	end int
}

// Represents a filtered array node i.e. `books[?(@.isbn)]`.
type arrayFilteredNode struct {
	node

	// The property to filter with.
	key string

	// The comparison oparator. Can be one of '=', '!', '<', '=<', '=>', '>'.
	op string

	// The value to compare with.
	value any
}

const (
	dataValidationErrorNotMap int = iota
	dataValidationErrorKeyNotFound
	dataValidationErrorValueNotArray
)

type dataValidationError struct {
	data      any
	key       string
	value     any
	errorType int
}

func (err dataValidationError) Error() string {
	prefix := "dataValidationError"

	switch err.errorType {
	case dataValidationErrorNotMap:
		return fmt.Sprintf("%v: Data is nil.", prefix)
	case dataValidationErrorKeyNotFound:
		return fmt.Sprintf("%v: Source key not found: '%v'", prefix, err.key)
	case dataValidationErrorValueNotArray:
		return fmt.Sprintf("%v: Value of key '%v' is not an array: %#v", prefix, err.key, err.value)
	}

	return prefix
}

// validateSource ensures that the provided data can be used by the node for retrieval or update.
func validateNodeData(n nodeDataAccessor, data map[string]any) error {
	nodeName := n.getName()

	if data == nil {
		return dataValidationError{data: data, errorType: dataValidationErrorNotMap}
	}

	if !mapHasKey(data, nodeName) {
		return dataValidationError{key: nodeName, errorType: dataValidationErrorKeyNotFound}
	}

	if isArrayNode(n) {
		value, _ := data[nodeName]

		if !isSlice(value) {
			return dataValidationError{key: nodeName, value: value, errorType: dataValidationErrorValueNotArray}
		}
	}

	return nil
}

// ----
// node
// ----

// get returns the value of the provided map data with key same as the name of the node.
func (n node) get(data map[string]any) (any, error) {
	if err := validateNodeData(n, data); err != nil {
		return nil, err
	}

	return data[n.name], nil
}

// put updates the value of the provided map data with key same as the name of the node.
func (n node) put(data map[string]any, value any) error {
	err := validateNodeData(n, data)

	// the key not found error is excluded because the key will be created anyway below
	if err != nil && err.(dataValidationError).errorType != dataValidationErrorKeyNotFound {
		return err
	}

	data[n.name] = value

	return nil
}

// getName returns the name of the node.
func (n node) getName() string { return n.name }

// ----------------
// arrayIndexedNode
// ----------------

// get returns the value of the provided map data with key same as the name of the node.
// The underlying value must be a slice and the returned value will be the slice
// containing only the values of the `data` defined by the indices of the node.
func (n arrayIndexedNode) get(data map[string]any) (any, error) {
	if err := validateNodeData(n, data); err != nil {
		return nil, err
	}

	value := data[n.name]

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
func (n arrayIndexedNode) put(data map[string]any, newVal any) error {
	if err := validateNodeData(n, data); err != nil {
		return err
	}

	value := data[n.name]

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
func (n arraySlicedNode) get(data map[string]any) (any, error) {
	if err := validateNodeData(n, data); err != nil {
		return nil, err
	}

	value := data[n.name]

	if n.start != 0 && n.end != 0 {
		return value.([]any)[n.start:n.end], nil
	}

	if n.start != 0 && n.end == 0 {
		return value.([]any)[n.start:], nil
	}

	if n.start == 0 && n.end != 0 {
		return value.([]any)[:n.end], nil
	}

	return data, nil
}

// put updates the value of the provided map data with key same as the name of the n.
// The underlying value must be a slice and the new value will apply on the slice
// defined by the indices of the n.
func (n arraySlicedNode) put(data map[string]any, newVal any) error {
	if err := validateNodeData(n, data); err != nil {
		return err
	}

	value := data[n.name]

	if n.start != 0 && n.end != 0 {
		value = value.([]any)[n.start:n.end]
	} else if n.start != 0 && n.end == 0 {
		value = value.([]any)[n.start:]
	} else if n.start == 0 && n.end != 0 {
		value = value.([]any)[:n.end]
	} else {
		return nil
	}

	for i := range value.([]any) {
		value.([]any)[i] = newVal
	}

	return nil
}

// getName returns the name of the n.
func (n arraySlicedNode) getName() string { return n.node.name }

// -----------------
// arrayFilteredNode
// -----------------

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
	case "!=":
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
func (n arrayFilteredNode) get(data map[string]any) (any, error) {
	if err := validateNodeData(n, data); err != nil {
		return nil, err
	}

	value := data[n.name]

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
func (n arrayFilteredNode) put(data map[string]any, newVal any) error {
	if err := validateNodeData(n, data); err != nil {
		return err
	}

	value := data[n.name]

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

type matchDictionary map[string]string

// getMatchDictionary returns a map of placeholders and their values found in a string given a pattern with placeholders in it.
func getMatchDictionary(patt string, s string) (dict matchDictionary) {
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

	dict = getMatchDictionary(jsonPathArrayNodePattern, jsonPathSubNode)
	if len(dict) > 0 {
		return arrayIndexedNode{
			node: node{
				name: dict["node"],
			},
		}
	}

	dict = getMatchDictionary(jsonPathIndexedArrayNodePattern, jsonPathSubNode)
	if len(dict) > 0 {
		node := arrayIndexedNode{
			node: node{
				name: dict["node"],
			},
		}
		indices := strings.Split(dict["indices"], ",")
		for _, index := range indices {
			indexInt, _ := strconv.Atoi(strings.TrimSpace(index))
			node.indices = append(node.indices, indexInt)
		}

		return node
	}

	dict = getMatchDictionary(jsonPathSlicedArrayNodePattern, jsonPathSubNode)
	if len(dict) > 0 {
		node := arraySlicedNode{
			node: node{
				name: dict["node"],
			},
		}
		node.start, _ = strconv.Atoi(dict["start"])
		node.end, _ = strconv.Atoi(dict["end"])

		return node
	}

	dict = getMatchDictionary(jsonPathFilteredArrayNodePattern, jsonPathSubNode)
	if len(dict) > 0 {
		return arrayFilteredNode{
			node: node{
				name: dict["node"],
			},
			key:   dict["key"],
			op:    dict["op"],
			value: dict["value"],
		}
	}

	dict = getMatchDictionary(jsonPathSimpleNodePattern, jsonPathSubNode)
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
func walkNodes(data map[string]any, nodes []nodeDataAccessor) (walkedData any, err error) {
	walkedData = data

	prevHasReccursiveDescent := false
	for _, n := range nodes {
		if n.getName() == "*" {
			continue
		}

		if n.getName() == "" {
			prevHasReccursiveDescent = true
			continue
		}

		if isSlice(walkedData) {
			var items []any
			for _, item := range walkedData.([]any) {
				value, err := n.get(item.(map[string]any))
				if err != nil {
					return nil, err
				}
				items = append(items, value)
			}
			walkedData = items
			continue
		}

		if prevHasReccursiveDescent {
			walkedData = mapGetDeepFlattened(walkedData, n.getName())
			if isArrayNode(n) {
				walkedDataWithkey := map[string]any{n.getName(): walkedData}
				walkedData, err = n.get(walkedDataWithkey)
				if err != nil {
					return nil, err
				}
			}
			prevHasReccursiveDescent = false
			continue
		}

		walkedData, err = n.get(walkedData.(map[string]any))
		if err != nil {
			return nil, err
		}
	}

	return walkedData, nil
}
