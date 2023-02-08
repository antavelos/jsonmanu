package jsonmanu

import (
	"fmt"
	"testing"

	"github.com/google/go-cmp/cmp"
)

type SplitTransformerTestCase struct {
	transformer              SplitTransformer
	value                    any
	expectedTransformedValue any
	expectedErrorMessage     string
}

func TestSplitTransformer(t *testing.T) {
	cases := []SplitTransformerTestCase{
		{
			transformer:              SplitTransformer{Delim: ",", Index: -1},
			value:                    "tok1,tok2,tok3",
			expectedTransformedValue: []string{"tok1", "tok2", "tok3"},
			expectedErrorMessage:     "",
		},
		{
			transformer:              SplitTransformer{Delim: ",", Index: 0},
			value:                    "tok1,tok2,tok3",
			expectedTransformedValue: "tok1",
			expectedErrorMessage:     "",
		},
		{
			transformer:              SplitTransformer{Delim: ",", Index: 2},
			value:                    "tok1,tok2,tok3",
			expectedTransformedValue: "tok3",
			expectedErrorMessage:     "",
		},
		{
			transformer:              SplitTransformer{Delim: ".", Index: 0},
			value:                    "tok1,tok2,tok3",
			expectedTransformedValue: "tok1,tok2,tok3",
			expectedErrorMessage:     "",
		},
		{
			transformer:              SplitTransformer{Delim: ",", Index: 4},
			value:                    "tok1,tok2,tok3",
			expectedTransformedValue: nil,
			expectedErrorMessage:     "SplitTransformer error: Index out of bounds.",
		},
		{
			transformer:              SplitTransformer{Delim: ",", Index: 4},
			value:                    1,
			expectedTransformedValue: nil,
			expectedErrorMessage:     "SplitTransformer error: value is not a string.",
		},
	}

	for _, tc := range cases {
		t.Run(fmt.Sprintf("SplitTransformer.transform(%v)=%v", tc.value, tc.expectedTransformedValue), func(t *testing.T) {
			transformedValue, err := tc.transformer.Transform(tc.value)

			if (err == nil && len(tc.expectedErrorMessage) > 0) || (err != nil && err.Error() != tc.expectedErrorMessage) {
				t.Errorf("Expected error message '%#v', but got '%#v'", tc.expectedErrorMessage, err.Error())
			}
			if !cmp.Equal(tc.expectedTransformedValue, transformedValue) {
				t.Errorf("Expected '%#v', but got '%#v'", tc.expectedTransformedValue, transformedValue)
			}
		})
	}
}

type ReplaceTransformerTestCase struct {
	transformer              ReplaceTransformer
	value                    any
	expectedTransformedValue any
	expectedErrorMessage     string
}

func TestReplaceSplitTransformer(t *testing.T) {
	cases := []ReplaceTransformerTestCase{
		{
			transformer:              ReplaceTransformer{OldVal: "lorem", NewVal: "ipsum"},
			value:                    "lorem ipsum lorem ipsum",
			expectedTransformedValue: "ipsum ipsum ipsum ipsum",
			expectedErrorMessage:     "",
		},
		{
			transformer:              ReplaceTransformer{OldVal: "lorem", NewVal: "ipsum"},
			value:                    "ipsum ipsum ipsum ipsum",
			expectedTransformedValue: "ipsum ipsum ipsum ipsum",
			expectedErrorMessage:     "",
		},
		{
			transformer:              ReplaceTransformer{OldVal: "lorem", NewVal: "ipsum"},
			value:                    1,
			expectedTransformedValue: "",
			expectedErrorMessage:     "ReplaceTransformer error: value is not a string.",
		},
	}

	for _, tc := range cases {
		t.Run(fmt.Sprintf("ReplaceTransformer.transform(%v)=%v", tc.value, tc.expectedTransformedValue), func(t *testing.T) {
			transformedValue, err := tc.transformer.Transform(tc.value)

			if (err == nil && len(tc.expectedErrorMessage) > 0) || (err != nil && err.Error() != tc.expectedErrorMessage) {
				t.Errorf("Expected error message '%#v', but got '%#v'", tc.expectedErrorMessage, err.Error())
			}
			if !cmp.Equal(tc.expectedTransformedValue, transformedValue) {
				t.Errorf("Expected '%#v', but got '%#v'", tc.expectedTransformedValue, transformedValue)
			}
		})
	}
}

type JoinTransformerTestCase struct {
	transformer              JoinTransformer
	value                    any
	expectedTransformedValue string
	expectedErrorMessage     string
}

func TestJoinSplitTransformer(t *testing.T) {
	cases := []JoinTransformerTestCase{
		{
			transformer:              JoinTransformer{Delim: ","},
			value:                    []string{"1", "2", "3"},
			expectedTransformedValue: "1,2,3",
			expectedErrorMessage:     "",
		},
		{
			transformer:              JoinTransformer{Delim: ","},
			value:                    []int{1, 2, 3},
			expectedTransformedValue: "1,2,3",
			expectedErrorMessage:     "",
		},
		{
			transformer:              JoinTransformer{Delim: ","},
			value:                    []uint{1, 2, 3},
			expectedTransformedValue: "1,2,3",
			expectedErrorMessage:     "",
		},
		{
			transformer:              JoinTransformer{Delim: "  "},
			value:                    []float64{1.3, 2.5, 3.7},
			expectedTransformedValue: "1.3  2.5  3.7",
			expectedErrorMessage:     "",
		},
		{
			transformer:              JoinTransformer{Delim: ", "},
			value:                    []bool{true, false, true},
			expectedTransformedValue: "true, false, true",
			expectedErrorMessage:     "",
		},
		{
			transformer:              JoinTransformer{Delim: ", "},
			value:                    []any{1, false, "something"},
			expectedTransformedValue: "1, false, something",
			expectedErrorMessage:     "",
		},
		{
			transformer:              JoinTransformer{Delim: ","},
			value:                    1,
			expectedTransformedValue: "",
			expectedErrorMessage:     "JoinTransformer error: value is not an array.",
		},
		{
			transformer:              JoinTransformer{Delim: ","},
			value:                    true,
			expectedTransformedValue: "",
			expectedErrorMessage:     "JoinTransformer error: value is not an array.",
		},
	}

	for _, tc := range cases {
		t.Run(fmt.Sprintf("JoinTransformer.transform(%v)=%v", tc.value, tc.expectedTransformedValue), func(t *testing.T) {
			transformedValue, err := tc.transformer.Transform(tc.value)

			if (err == nil && len(tc.expectedErrorMessage) > 0) || (err != nil && err.Error() != tc.expectedErrorMessage) {
				t.Errorf("Expected error message '%#v', but got '%#v'", tc.expectedErrorMessage, err.Error())
			}
			if !cmp.Equal(tc.expectedTransformedValue, transformedValue) {
				t.Errorf("Expected '%#v', but got '%#v'", tc.expectedTransformedValue, transformedValue)
			}
		})
	}
}

type MapTestCase struct {
	src                   any
	dst                   any
	mappers               []Mapper
	expectedDst           any
	expectedErrorMessages []string
}

func TestMap(t *testing.T) {
	cases := []MapTestCase{
		{
			src: map[string]any{
				"library": map[string]any{
					"books": []any{
						map[string]any{"author": "Nietzsche"},
						map[string]any{"author": "Stirner"},
					},
				},
			},
			dst: map[string]any{"authors": nil},
			mappers: []Mapper{
				Mapper{
					SrcNode: JsonNode{Path: "$.library.books.author", Type: String},
					DstNode: JsonNode{Path: "$.authors", Type: Array},
				},
			},
			expectedDst:           map[string]any{"authors": []any{"Nietzsche", "Stirner"}},
			expectedErrorMessages: []string{},
		},
		{
			src: map[string]any{
				"library": map[string]any{
					"books": []any{
						map[string]any{"author": "Nietzsche"},
						map[string]any{"author": "Stirner"},
					},
				},
			},
			dst: map[string]any{},
			mappers: []Mapper{
				Mapper{
					SrcNode: JsonNode{Path: "$.library.books.author", Type: String},
					DstNode: JsonNode{Path: "$.authors", Type: Array},
				},
			},
			expectedDst:           map[string]any{"authors": []any{"Nietzsche", "Stirner"}},
			expectedErrorMessages: []string{},
		},
		{
			src: map[string]any{
				"library": map[string]any{
					"books": []any{
						map[string]any{"author": "Nietzsche"},
						map[string]any{"author": "Stirner"},
					},
				},
			},
			dst: map[string]any{"authors": []int{1, 2, 3}},
			mappers: []Mapper{
				Mapper{
					SrcNode: JsonNode{Path: "$.library.books.author", Type: String},
					DstNode: JsonNode{Path: "$.authors", Type: Array},
				},
			},
			expectedDst:           map[string]any{"authors": []any{"Nietzsche", "Stirner"}},
			expectedErrorMessages: []string{},
		},
		{
			src: map[string]any{
				"library": map[string]any{
					"books": []any{
						map[string]any{"author": "Nietzsche"},
						map[string]any{"author": "Stirner"},
					},
				},
			},
			dst: map[string]any{"authors": nil},
			mappers: []Mapper{
				Mapper{
					SrcNode: JsonNode{Path: "$.library.books.author", Type: String},
					DstNode: JsonNode{Path: "$.library.authors", Type: Array},
				},
			},
			expectedDst:           map[string]any{"authors": nil, "library": map[string]any{"authors": []any{"Nietzsche", "Stirner"}}},
			expectedErrorMessages: []string{},
		},
		{
			src: map[string]any{
				"library": map[string]any{
					"books": []any{
						map[string]any{"author": "Nietzsche"},
						map[string]any{"author": "Stirner"},
					},
					"categories": []any{
						map[string]any{"name": "Philosophy"},
						map[string]any{"name": "Literature"},
					},
				},
			},
			dst: make(map[string]any),
			mappers: []Mapper{
				Mapper{
					SrcNode: JsonNode{Path: "$.library.books.author", Type: String},
					DstNode: JsonNode{Path: "$.library.authors", Type: Array},
				},
				Mapper{
					SrcNode: JsonNode{Path: "$.library.categories.name", Type: String},
					DstNode: JsonNode{Path: "$.library.categories", Type: Array},
				},
			},
			expectedDst: map[string]any{
				"library": map[string]any{
					"authors": []any{
						"Nietzsche", "Stirner",
					},
					"categories": []any{
						"Philosophy", "Literature",
					},
				},
			},
			expectedErrorMessages: []string{},
		},
		{
			src: map[string]any{
				"library": map[string]any{
					"books": []any{
						map[string]any{"author": "Nietzsche"},
						map[string]any{"author": "Stirner"},
					},
					"categories": []any{
						map[string]any{"name": "Philosophy"},
						map[string]any{"name": "Literature"},
					},
				},
			},
			dst: nil,
			mappers: []Mapper{
				Mapper{
					SrcNode: JsonNode{Path: "$.library.books.author", Type: String},
					DstNode: JsonNode{Path: "$.library.authors", Type: Array},
				},
				Mapper{
					SrcNode: JsonNode{Path: "$.library.categories.name", Type: String},
					DstNode: JsonNode{Path: "$.library.categories", Type: Array},
				},
			},
			expectedDst: nil,
			expectedErrorMessages: []string{
				"mapper[0]: Error while putting value in destination: SourceValidationError: Source is not a map: '<nil>'",
				"mapper[1]: Error while putting value in destination: SourceValidationError: Source is not a map: '<nil>'"},
		},
	}

	for i, tc := range cases {
		t.Run(fmt.Sprintf("[%v] Map(%v, %v, %v)=%v", i, tc.src, tc.dst, tc.mappers, tc.expectedErrorMessages), func(t *testing.T) {
			errors := Map(tc.src, tc.dst, tc.mappers)

			for i, err := range errors {
				if (err == nil && len(tc.expectedErrorMessages[i]) > 0) || (err != nil && err.Error() != tc.expectedErrorMessages[i]) {
					t.Errorf("Expected error message '%#v', but got '%#v'", tc.expectedErrorMessages[i], err.Error())
				}
			}
			if !cmp.Equal(tc.expectedDst, tc.dst) {
				t.Errorf("Expected '%s', but got '%s'", prettify(tc.expectedDst), prettify(tc.dst))
			}
		})
	}
}
