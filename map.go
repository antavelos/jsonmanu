package jsonmanu

import (
	"fmt"
)

// Mapper holds the configuration of a mapping from a certain data structure to another one.
type Mapper struct {
	// SrcJsonPath is the JsonPath of the source data where data will be retrieved from
	SrcJsonPath string

	// DstJsonPath is the JsonPath of the destination data where data will be put in
	DstJsonPath string

	// Transformations enable optional functionality to be applied on the retrieved value before it's put in the destination data.
	// The transformations will be applied in a chain mode according to their order
	Transformations []Transformation
}

// handleSlideTransformation applies the transformation on each element of the slice
func handleSlideTransformation(value any, transformer Transformer) (any, error) {
	var transArray []any
	i := 0
	for item := range iterAny(value, nil) {
		transItem, err := transformer.Transform(item)
		if err != nil {
			return value, fmt.Errorf("Array[%v]: %v", i, err)
		}
		transArray = append(transArray, transItem)
		i++
	}
	value = transArray

	return value, nil
}

// handleMapper handles the cycle of a mapping of a src value to a dest based on the mapper conf
func handleMapper(src any, dst any, mapper Mapper) error {
	if err := validateMapper(mapper); err != nil {
		return fmt.Errorf("Validation error: %v", err)
	}

	srcValue, err := Get(src, mapper.SrcJsonPath)
	if err != nil {
		return fmt.Errorf("Error while getting value from source: %v", err)
	}

	for i, transformation := range mapper.Transformations {
		if isSlice(srcValue) {
			if transformation.AsArray {
				srcValue, err = transformation.Trsnfmr.Transform(srcValue)
			} else {
				srcValue, err = handleSlideTransformation(srcValue, transformation.Trsnfmr)
			}
		} else {
			srcValue, err = transformation.Trsnfmr.Transform(srcValue)
		}

		if err != nil {
			return fmt.Errorf("Transformation[%v] (%T): %v", i, transformation.Trsnfmr, err)
		}
	}

	if err = Put(dst, mapper.DstJsonPath, srcValue); err != nil {
		return fmt.Errorf("Error while putting value in destination: %v", err)
	}

	return nil
}

// validateMapper validates a mapperconfiguration
func validateMapper(mapper Mapper) error {
	if pathHasReccursiveDescent(mapper.DstJsonPath) {
		return fmt.Errorf("Reccursive descent not allowed in destination path.")
	}

	return nil
}

// Map maps data from a given map to another based on a configuration described in one or more Mapper objects
func Map(src any, dst any, mappers []Mapper) (errors []error) {
	for i, mapper := range mappers {
		if err := handleMapper(src, dst, mapper); err != nil {
			errors = append(errors, fmt.Errorf("Mapper[%v]: %s", i, err.Error()))
		}
	}

	return
}
