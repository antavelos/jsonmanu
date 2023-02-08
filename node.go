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
	get(any) (any, error)

	// Updates a given map according to the rules of the node called.
	put(any, any) error
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
	sourceValidationErrorNotMap int = iota
	sourceValidationErrorKeyNotFound
	sourceValidationErrorValueNotArray
)

type SourceValidationError struct {
	source    any
	key       string
	value     any
	errorType int
}

func (err SourceValidationError) Error() string {
	prefix := "SourceValidationError"

	switch err.errorType {
	case sourceValidationErrorNotMap:
		return fmt.Sprintf("%v: Source is not a map: '%#v'", prefix, err.source)
	case sourceValidationErrorKeyNotFound:
		return fmt.Sprintf("%v: Source key not found: '%v'", prefix, err.key)
	case sourceValidationErrorValueNotArray:
		return fmt.Sprintf("%v: Value of key '%v' is not an array: %#v", prefix, err.key, err.value)
	}

	return prefix
}

// validateSource ensures that the provided data can be used by the node for retrieval or update.
func validateNodeSource(n nodeDataAccessor, source any) error {
	nodeName := n.getName()

	if !isMap(source) {
		return SourceValidationError{source: source, errorType: sourceValidationErrorNotMap}
	}

	if !mapHasKey(source, nodeName) {
		return SourceValidationError{key: nodeName, errorType: sourceValidationErrorKeyNotFound}
	}

	if isArrayNode(n) {
		value, _ := source.(map[string]any)[nodeName]

		if !isSlice(value) {
			return SourceValidationError{key: nodeName, value: value, errorType: sourceValidationErrorValueNotArray}
		}
	}

	return nil
}

// ----
// node
// ----

// get returns the value of the provided map data with key same as the name of the node.
func (n node) get(source any) (any, error) {
	if err := validateNodeSource(n, source); err != nil {
		return nil, err
	}

	return source.(map[string]any)[n.name], nil
}

// put updates the value of the provided map data with key same as the name of the node.
func (n node) put(source any, value any) error {
	err := validateNodeSource(n, source)

	// if err != nil {

	// 	switch err.(SourceValidationError).errorType {
	// 	case sourceValidationErrorValueNotArray:
	// 		return err
	// 	}
	// }
	// the key not found error is excluded because the key will be created anyway below
	if err != nil && err.(SourceValidationError).errorType != sourceValidationErrorKeyNotFound {
		return err
	}

	source.(map[string]any)[n.name] = value

	return nil
}

// getName returns the name of the node.
func (n node) getName() string { return n.name }

// ----------------
// arrayIndexedNode
// ----------------

// get returns the value of the provided map data with key same as the name of the node.
// The underlying value must be a slice and the returned value will be the slice
// containing only the values of the `source` defined by the indices of the node.
func (n arrayIndexedNode) get(source any) (any, error) {
	if err := validateNodeSource(n, source); err != nil {
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
	if err := validateNodeSource(n, source); err != nil {
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
	if err := validateNodeSource(n, source); err != nil {
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
	if err := validateNodeSource(n, source); err != nil {
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
func (n arrayFilteredNode) get(source any) (any, error) {
	if err := validateNodeSource(n, source); err != nil {
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
	if err := validateNodeSource(n, source); err != nil {
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
func walkNodes(data any, nodes []nodeDataAccessor) (any, error) {
	var err error

	prevHasReccursiveDescent := false
	for _, n := range nodes {
		if n.getName() == "*" {
			continue
		}

		if n.getName() == "" {
			prevHasReccursiveDescent = true
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

		if prevHasReccursiveDescent {
			data = mapGetDeepFlattened(data, n.getName())
			if isArrayNode(n) {
				dataWithkey := map[string]any{n.getName(): data}
				data, err = n.get(dataWithkey)
				if err != nil {
					return nil, err
				}
			}
			prevHasReccursiveDescent = false
			continue
		}

		data, err = n.get(data)
		if err != nil {
			return nil, err
		}
	}

	return data, nil
}
