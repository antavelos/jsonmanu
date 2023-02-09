package jsonmanu

import (
	"fmt"
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
		return nil, fmt.Errorf("SplitTransformer error: value is not a string.")
	}

	split := strings.Split(value.(string), t.Delim)

	if t.Index >= len(split) {
		return nil, fmt.Errorf("SplitTransformer error: Index out of bounds.")
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
		return "", fmt.Errorf("JoinTransformer error: value is not an array.")
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
		return "", fmt.Errorf("ReplaceTransformer error: value is not a string.")
	}

	return strings.Replace(value.(string), t.OldVal, t.NewVal, -1), nil
}

const (
	Object int = iota
	Array
	String
	Number
	Boolean
	Null
)

type JsonNode struct {
	Path string
	Type int
}

type Mapper struct {
	SrcNode     JsonNode
	DstNode     JsonNode
	Transformer Transformer
}

func handleMapper(src any, dst any, mapper Mapper) error {
	err := validateMapper(mapper)
	if err != nil {
		return fmt.Errorf("Validation error: %v", err)
	}

	srcValue, err := Get(src, mapper.SrcNode.Path)
	if err != nil {
		return fmt.Errorf("Error while getting value from source: %v", err)
	}

	if mapper.Transformer != nil {
		srcValue, err = mapper.Transformer.Transform(srcValue)
		if err != nil {
			return fmt.Errorf("Error while transforming value: %v", err)
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
		return fmt.Errorf("reccursive descent not allowed in destination path")
	}

	return nil
}

func Map(src any, dst any, mappers []Mapper) (errors []error) {
	for i, mapper := range mappers {
		err := handleMapper(src, dst, mapper)
		if err != nil {
			errors = append(errors, fmt.Errorf("mapper[%v]: %s", i, err.Error()))
		}
	}

	return
}
