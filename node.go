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

// Interface to be implemented by all node like structs for name retrieval.
type namedNode interface {
	getName() string
}

// Interface to be implemented by all node like structs for retrieving and updating JSON node data.
type nodeDataAccessor interface {
	namedNode

	// Retrieves a value out of a given map according to the rules of the node called.
	get(any) any

	// Updates a given map according to the rules of the node called.
	put(any, any) error
}

type DataValidationError string

func (err DataValidationError) Error() string {
	return fmt.Sprintf("DataValidationError error: %s", string(err))
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

// validateSourceData ensures that the provided data can be used by the node for retrieval or update.
func (n node) validateSourceData(sourceData any) (any, error) {
	if !isMap(sourceData) {
		return nil, DataValidationError(fmt.Sprintf("Source data is not a map: %#v", sourceData))
	}

	data, ok := sourceData.(map[string]any)[n.name]
	if !ok {
		return nil, DataValidationError(fmt.Sprintf("key '%v' not found", n.name))
	}

	return data, nil
}

// get returns the value of the provided map data with key same as the name of the node.
func (n node) get(sourceData any) any {
	_, err := n.validateSourceData(sourceData)
	if err != nil {
		return nil
	}

	return sourceData.(map[string]any)[n.name]
}

// put updates the value of the provided map data with key same as the name of the node.
func (n node) put(sourceData any, value any) error {
	if _, err := n.validateSourceData(sourceData); err != nil {
		return err
	}

	sourceData.(map[string]any)[n.name] = value

	return nil
}

// getName returns the name of the node.
func (n node) getName() string { return n.name }

// ---------
// arrayNode
// ---------

// validateSourceData ensures that the provided data can be used by the array node for retrieval or update
func (n arrayNode) validateSourceData(sourceData any) (any, error) {
	data, err := n.node.validateSourceData(sourceData)
	if err != nil {
		return nil, err
	}

	if !isSlice(data) {
		return nil, DataValidationError(fmt.Sprintf("Value of key '%v' is not an array: %#v", n.node.name, sourceData))
	}

	return data, nil
}

// ----------------
// arrayIndexedNode
// ----------------

// get returns the value of the provided map data with key same as the name of the node.
// The underlying value must be a slice and the returned value will be the slice
// containing only the values of the `sourceData` defined by the indices of the node.
func (n arrayIndexedNode) get(sourceData any) any {
	data, err := n.validateSourceData(sourceData)
	if err != nil {
		return nil
	}

	if len(n.indices) == 0 {
		return data
	}

	var result []any
	for _, i := range n.indices {
		if i < 0 || i >= len(data.([]any)) {
			continue
		}
		result = append(result, data.([]any)[i])
	}

	return result
}

// put updates the value of the provided map data with key same as the name of the node.
// The underlying value must be a slice and the new value will apply on the slice
// defined by the indices of the node.
func (n arrayIndexedNode) put(sourceData any, value any) error {
	data, err := n.validateSourceData(sourceData)
	if err != nil {
		return err
	}

	for _, i := range n.indices {
		if i < 0 || i >= len(data.([]any)) {
			continue
		}
		data.([]any)[i] = value
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
func (n arraySlicedNode) get(sourceData any) any {
	data, err := n.validateSourceData(sourceData)
	if err != nil {
		return nil
	}

	if n.start != 0 && n.end != 0 {
		return data.([]any)[n.start:n.end]
	}

	if n.start != 0 && n.end == 0 {
		return data.([]any)[n.start:]
	}

	if n.start == 0 && n.end != 0 {
		return data.([]any)[:n.end]
	}

	return sourceData
}

// put updates the value of the provided map data with key same as the name of the n.
// The underlying value must be a slice and the new value will apply on the slice
// defined by the indices of the n.
func (n arraySlicedNode) put(sourceData any, value any) error {
	data, err := n.validateSourceData(sourceData)
	if err != nil {
		return err
	}

	if n.start != 0 && n.end != 0 {
		data = data.([]any)[n.start:n.end]
	} else if n.start != 0 && n.end == 0 {
		data = data.([]any)[n.start:]
	} else if n.start == 0 && n.end != 0 {
		data = data.([]any)[:n.end]
	} else {
		return nil
	}

	for i, _ := range data.([]any) {
		data.([]any)[i] = value
	}

	return nil
}

// getName returns the name of the n.
func (n arraySlicedNode) getName() string { return n.node.name }

// -----------------
// arrayFilteredNode
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

// get returns the value of the provided map data with key same as the name of the n.
// The underlying value must be a slice and the returned value will be the subslice
// that satisfies the condition defived by the key, value and operator of the n.
func (n arrayFilteredNode) get(sourceData any) any {
	data, err := n.validateSourceData(sourceData)
	if err != nil {
		return nil
	}

	var filtered []any
	for _, item := range data.([]any) {
		value, ok := item.(map[string]any)[n.key]
		if !ok {
			continue
		}

		if len(n.op) == 0 || n.value == nil || assertCondition(value, n.value, n.op) {
			filtered = append(filtered, item)
		}
	}

	return filtered
}

// put updates the value of the provided map data with key same as the name of the n.
// The underlying value must be a slice and the returned value will be the subslice
// that satisfies the condition defived by the key, value and operator of the n.
func (n arrayFilteredNode) put(sourceData any, value any) error {
	data, err := n.validateSourceData(sourceData)
	if err != nil {
		return err
	}

	for _, item := range data.([]any) {
		currValue, ok := item.(map[string]any)[n.key]
		if !ok {
			continue
		}

		if len(n.op) == 0 || n.value == nil || assertCondition(currValue, n.value, n.op) {
			item.(map[string]any)[n.key] = value
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

// nodeFromToken checks one by one the existing token patterns and returns an appropriate node data accessor.
func nodeFromToken(token string) nodeDataAccessor {
	var dict map[string]string

	dict = getMatchDictionary(ARRAY_TOKEN_PATTERN, token)
	if len(dict) > 0 {
		return arrayIndexedNode{
			arrayNode: arrayNode{
				node: node{
					name: dict["node"],
				},
			},
		}
	}

	dict = getMatchDictionary(INDEXED_ARRAY_TOKEN_PATTERN, token)
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

	dict = getMatchDictionary(SLICED_ARRAY_TOKEN_PATTERN, token)
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

	dict = getMatchDictionary(FILTERED_ARRAY_TOKEN_PATTERN, token)
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

	dict = getMatchDictionary(SIMPLE_NODE_TOKEN_PATTERN, token)
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
func walkNodes(data any, nodes []nodeDataAccessor) any {
	withReccursiveDescent := false
	for _, node := range nodes {
		if node.getName() == "*" {
			continue
		}

		if node.getName() == "" {
			withReccursiveDescent = true
			continue
		}

		if isSlice(data) {
			var idata []any
			for _, item := range data.([]any) {
				idata = append(idata, node.get(item))
			}
			data = idata
			continue
		}

		if withReccursiveDescent {
			data = mapGetDeepFlattened(data, node.getName())
			if isArrayNode(node) {
				dataWithkey := map[string]any{node.getName(): data}
				data = node.get(dataWithkey)
			}
			continue
		}

		data = node.get(data)
	}

	return data
}
