package jsonmanu

import (
	"errors"
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

// Transformer is a type to be used in order to apply some logic on a given value.
type Transformer interface {

	// Transform will apply some logic on the provided value and will return it transformed unless an error occurs in the process.
	Transform(value any) (any, error)
}

// Transformation holds a transformar to be applied along with extra relevant logic.
type Transformation struct {

	// Transformer applies a predefined logic on a specific value
	Trsnfmr Transformer

	// AsArray determines whether a retrieved data array value will be retrieved as a whole array or not.
	// An array retrieved value is such either because a grouping occured due to a parent array higher in the data structure
	// or because it is actually an array in the original data. By default the transformation will apply to each element of
	// the array unless this flag is set as true.
	AsArray bool
}

// SplitTransformer will split a string value based on the provided delimeter and from the occured array it will pick the element
// defined by the provided index.
type SplitTransformer struct {

	// Delim is the delimeter value found in the string to be split, i.e. `,` or `-` etc
	Delim string

	// Index is the index of the element to be picked.
	Index int
}

// SplitTransformer Transform applies the split transformation:
// It expects a string value.
// If expects an index within the bounds of the occured array.
// If the provided index is -1 then the whole occured array will be returned.
func (t SplitTransformer) Transform(value any) (any, error) {
	if !isString(value) {
		return nil, errors.New("Value is not a string.")
	}

	split := strings.Split(value.(string), t.Delim)

	if t.Index >= len(split) {
		return nil, errors.New("Index out of bounds.")
	}

	if t.Index == -1 {
		return split, nil
	}

	return split[t.Index], nil
}

// JoinTransformer joins the values of an array based on the provided delimiter.
type JoinTransformer struct {

	// Delim is the delimeter value to be used, i.e. `,` or `-` etc
	Delim string
}

// JoinTransformer Transform applies the join transformation:
// It expects an array value which can be of any kind.
// The array elements must implement the String method of the Stringer interface.
func (t JoinTransformer) Transform(value any) (any, error) {
	if !isSlice(value) {
		return nil, errors.New("Value is not an array.")
	}

	var strSlice []string
	for item := range iterAny(value, nil) {
		strSlice = append(strSlice, fmt.Sprintf("%v", item))
	}

	return strings.Join(strSlice, t.Delim), nil
}

// ReplaceTransformer replaces a substring in a string value with another.
type ReplaceTransformer struct {

	// OldVal is the substring to be replaced.
	OldVal string

	// NewVal is the value to replace the OldValue with.
	NewVal string
}

// ReplaceTransformer Transform applies the replace transformation:
// It expects a string value.
func (t ReplaceTransformer) Transform(value any) (any, error) {
	if !isString(value) {
		return nil, errors.New("Value is not a string.")
	}

	return strings.Replace(value.(string), t.OldVal, t.NewVal, -1), nil
}

// StringMatchTransformer is used to match a substring within a string.
type StringMatchTransformer struct {

	// Regex is the regular expression to be used for the string matching.
	Regex string
}

// StringMatchTransformer Transform applies the string match transformation:
// It expects a string value.
// It will return the first matched substring found in the provided value.
func (t StringMatchTransformer) Transform(value any) (any, error) {
	if !isString(value) {
		return nil, errors.New("Value is not a string.")
	}

	re := regexp.MustCompile(t.Regex)
	result := re.FindString(value.(string))

	return result, nil
}

// SubStrTransformer will return a slice of a string value based on the provided indices.
type SubStrTransformer struct {

	// Start is the starting index.
	Start int

	// End is the ending index.
	End int
}

// SubStrTransformer Transform applies the substring transformation:
// It expectd a string value.
// The indices must be withing the length of the string value.
// If End index is not provided the value will be sliced from Start index to the end of the value.
func (t SubStrTransformer) Transform(value any) (any, error) {
	if !isString(value) {
		return nil, errors.New("Value is not a string.")
	}

	if t.Start < 0 {
		return nil, errors.New("Start index out of bound.")
	}

	if t.End >= len(value.(string)) {
		return nil, errors.New("End index out of bound.")
	}

	if t.End == 0 {
		t.End = len(value.(string))
	}
	return value.(string)[t.Start:t.End], nil
}

// NumberTransformer converts a string value to float64.
type NumberTransformer struct{}

// NumberTransformer Transform applies the number transformation.
// It expects a numerical string value i.e. "234", "23.434" etc
// It will returned value will be float64 so "123.2" will be transformed to 123.2 and "123" will be transformed to 123.0.
func (t NumberTransformer) Transform(value any) (any, error) {
	if !isString(value) {
		return nil, errors.New("Value is not a string.")
	}

	fv, err := strconv.ParseFloat(value.(string), 1)
	if err != nil {
		return nil, errors.New("Couldn't convert value to number.")
	}

	return fv, nil
}
