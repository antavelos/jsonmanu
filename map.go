package jsonmanu

import (
	"errors"
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

type Transformer interface {
	Transform(value any) (any, error)
}

type SplitTransformer struct {
	Delim string
	Index int
}

func (t SplitTransformer) Transform(value any) (any, error) {
	if !isString(value) {
		return nil, errors.New("SplitTransformer error: value is not a string.")
	}

	split := strings.Split(value.(string), t.Delim)

	if t.Index >= len(split) {
		return nil, errors.New("SplitTransformer error: Index out of bounds.")
	}

	if t.Index == -1 {
		return split, nil
	}

	return split[t.Index], nil
}

type JoinTransformer struct {
	Delim string
}

func (t JoinTransformer) Transform(value any) (any, error) {
	if !isSlice(value) {
		return "", errors.New("JoinTransformer error: value is not an array.")
	}

	var strSlice []string
	for item := range iterAny(value, nil) {
		strSlice = append(strSlice, fmt.Sprintf("%v", item))
	}

	return strings.Join(strSlice, t.Delim), nil
}

type ReplaceTransformer struct {
	OldVal string
	NewVal string
}

func (t ReplaceTransformer) Transform(value any) (any, error) {
	if !isString(value) {
		return "", errors.New("ReplaceTransformer error: value is not a string.")
	}

	return strings.Replace(value.(string), t.OldVal, t.NewVal, -1), nil
}

type StringMatchTransformer struct {
	Regex string
}

func (t StringMatchTransformer) Transform(value any) (any, error) {
	re := regexp.MustCompile(t.Regex)
	result := re.FindString(value.(string))

	return result, nil
}

type SubStrTransformer struct {
	Start int
	End   int
}

func (t SubStrTransformer) Transform(value any) (any, error) {
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

type NumberTransformer struct{}

func (t NumberTransformer) Transform(value any) (any, error) {
	if !isString(value) {
		return nil, errors.New("Value should be a string.")
	}

	fv, err := strconv.ParseFloat(value.(string), 1)
	if err != nil {
		return nil, errors.New("Couldn't convert value to number.")
	}

	return fv, nil
}

type DataType int

const (
	Object DataType = iota
	Array
	String
	Number
	Boolean
	Null
)

type JsonNode struct {
	Path string
	Type DataType
}

type Mapper struct {
	SrcNode      JsonNode
	DstNode      JsonNode
	Transformers []Transformer
	asArray      bool
}

// handleSlideTransformation applies the transformation on each element of the slice
func handleSlideTransformation(value any, transformer Transformer) (any, error) {
	var transArray []any
	for item := range iterAny(value, nil) {
		transItem, err := transformer.Transform(item)
		if err != nil {
			return value, err
		}
		transArray = append(transArray, transItem)
	}
	value = transArray

	return value, nil
}

// handleMapper handles the cycle of a mapping of a src value to a dest based on the mapper conf
func handleMapper(src any, dst any, mapper Mapper) error {
	err := validateMapper(mapper)
	if err != nil {
		return fmt.Errorf("Validation error: %v", err)
	}

	srcValue, err := Get(src, mapper.SrcNode.Path)
	if err != nil {
		return fmt.Errorf("Error while getting value from source: %v", err)
	}

	for i, transformer := range mapper.Transformers {
		if isSlice(srcValue) {
			if mapper.asArray {
				srcValue, err = transformer.Transform(srcValue)
			} else {
				srcValue, err = handleSlideTransformation(srcValue, transformer)
			}
		} else {
			srcValue, err = transformer.Transform(srcValue)
		}

		if err != nil {
			return fmt.Errorf("Transformer[%v]: Error while transforming value: %v", i, err)
		}
	}

	err = Put(dst, mapper.DstNode.Path, srcValue)
	if err != nil {
		return fmt.Errorf("Error while putting value in destination: %v", err)
	}

	return nil
}

func validateMapper(mapper Mapper) error {
	if pathHasReccursiveDescent(mapper.DstNode.Path) {
		return fmt.Errorf("Reccursive descent not allowed in destination path")
	}

	return nil
}

func Map(src any, dst any, mappers []Mapper) (errors []error) {
	for i, mapper := range mappers {
		err := handleMapper(src, dst, mapper)
		if err != nil {
			errors = append(errors, fmt.Errorf("Mapper[%v]: %s", i, err.Error()))
		}
	}

	return
}
