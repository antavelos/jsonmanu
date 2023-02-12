package jsonmanu

import (
	"fmt"
	"testing"

	"github.com/google/go-cmp/cmp"
)

type TransformerTestCase struct {
	transformer              Transformer
	value                    any
	expectedTransformedValue any
	expectedErrorMessage     string
}

func TestSplitTransformer(t *testing.T) {
	cases := []TransformerTestCase{
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

func TestReplaceSplitTransformer(t *testing.T) {
	cases := []TransformerTestCase{
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

func TestJoinSplitTransformer(t *testing.T) {
	cases := []TransformerTestCase{
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

func TestStringMatchSplitTransformer(t *testing.T) {
	cases := []TransformerTestCase{
		{
			transformer:              StringMatchTransformer{Regex: `abc`},
			value:                    "123abcdefg",
			expectedTransformedValue: "abc",
			expectedErrorMessage:     "",
		},
		{
			transformer:              StringMatchTransformer{Regex: `$abc`},
			value:                    "123abcdefg",
			expectedTransformedValue: "",
			expectedErrorMessage:     "",
		},
		{
			transformer:              StringMatchTransformer{Regex: `^\d*`},
			value:                    "123abcdefg",
			expectedTransformedValue: "123",
			expectedErrorMessage:     "",
		},
	}

	for _, tc := range cases {
		t.Run(fmt.Sprintf("StringMatchTransformer.transform(%v)=%v", tc.value, tc.expectedTransformedValue), func(t *testing.T) {
			transformedValue, err := tc.transformer.Transform(tc.value)

			if err == nil && len(tc.expectedErrorMessage) > 0 {
				t.Errorf("Expected error message '%#v', but got '%#v'", tc.expectedErrorMessage, err)
			}
			if err != nil && err.Error() != tc.expectedErrorMessage {
				t.Errorf("Expected error message '%#v', but got '%#v'", tc.expectedErrorMessage, err.Error())
			}
			if !cmp.Equal(tc.expectedTransformedValue, transformedValue) {
				t.Errorf("Expected '%#v', but got '%#v'", tc.expectedTransformedValue, transformedValue)
			}
		})
	}
}

func TestSubStrSplitTransformer(t *testing.T) {
	cases := []TransformerTestCase{
		{
			transformer:              SubStrTransformer{Start: -1},
			value:                    "123abcdefg",
			expectedTransformedValue: nil,
			expectedErrorMessage:     "Start index out of bound.",
		},
		{
			transformer:              SubStrTransformer{End: 100},
			value:                    "123abcdefg",
			expectedTransformedValue: nil,
			expectedErrorMessage:     "End index out of bound.",
		},
		{
			transformer:              SubStrTransformer{Start: 0, End: 3},
			value:                    "123abcdefg",
			expectedTransformedValue: "123",
			expectedErrorMessage:     "",
		},
		{
			transformer:              SubStrTransformer{Start: 3},
			value:                    "123abcdefg",
			expectedTransformedValue: "abcdefg",
			expectedErrorMessage:     "",
		},
		{
			transformer:              SubStrTransformer{End: 3},
			value:                    "123abcdefg",
			expectedTransformedValue: "123",
			expectedErrorMessage:     "",
		},
	}

	for _, tc := range cases {
		t.Run(fmt.Sprintf("SubStrMatchTransformer.transform(%v)=%v", tc.value, tc.expectedTransformedValue), func(t *testing.T) {
			transformedValue, err := tc.transformer.Transform(tc.value)

			if err == nil && len(tc.expectedErrorMessage) > 0 {
				t.Errorf("Expected error message '%#v', but got '%#v'", tc.expectedErrorMessage, err)
			}
			if err != nil && err.Error() != tc.expectedErrorMessage {
				t.Errorf("Expected error message '%#v', but got '%#v'", tc.expectedErrorMessage, err.Error())
			}
			if !cmp.Equal(tc.expectedTransformedValue, transformedValue) {
				t.Errorf("Expected '%#v', but got '%#v'", tc.expectedTransformedValue, transformedValue)
			}
		})
	}
}

func TestNumberTransformer(t *testing.T) {
	cases := []TransformerTestCase{
		{
			transformer:              NumberTransformer{},
			value:                    1,
			expectedTransformedValue: nil,
			expectedErrorMessage:     "Value should be a string.",
		},
		{
			transformer:              NumberTransformer{},
			value:                    "invalid",
			expectedTransformedValue: nil,
			expectedErrorMessage:     "Couldn't convert value to number.",
		},
		{
			transformer:              NumberTransformer{},
			value:                    "123",
			expectedTransformedValue: 123.0,
			expectedErrorMessage:     "",
		},
		{
			transformer:              NumberTransformer{},
			value:                    "123.456",
			expectedTransformedValue: 123.456,
			expectedErrorMessage:     "",
		},
	}

	for _, tc := range cases {
		t.Run(fmt.Sprintf("NumberTransformer.transform(%v)=%v", tc.value, tc.expectedTransformedValue), func(t *testing.T) {
			transformedValue, err := tc.transformer.Transform(tc.value)

			if err == nil && len(tc.expectedErrorMessage) > 0 {
				t.Errorf("Expected error message '%#v', but got '%#v'", tc.expectedErrorMessage, err)
			}
			if err != nil && err.Error() != tc.expectedErrorMessage {
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
					SrcJsonPath: "$.library.books.author",
					DstJsonPath: "$.authors",
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
					SrcJsonPath: "$.library.books.author",
					DstJsonPath: "$.authors",
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
					SrcJsonPath: "$.library.books.author",
					DstJsonPath: "$.authors",
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
					SrcJsonPath: "$.library.books.author",
					DstJsonPath: "$.library.authors",
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
					SrcJsonPath: "$.library.books.author",
					DstJsonPath: "$.library.authors",
				},
				Mapper{
					SrcJsonPath: "$.library.categories.name",
					DstJsonPath: "$.library.categories",
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
					SrcJsonPath: "$.library.books.author",
					DstJsonPath: "$.library.authors",
				},
				Mapper{
					SrcJsonPath: "$.library.categories.name",
					DstJsonPath: "$.library.categories",
				},
			},
			expectedDst: nil,
			expectedErrorMessages: []string{
				"Mapper[0]: Error while putting value in destination: DataValidationError: Data is not a map: '<nil>'",
				"Mapper[1]: Error while putting value in destination: DataValidationError: Data is not a map: '<nil>'"},
		},
		{
			src: map[string]any{
				"area": map[string]any{
					"coordinates": "45.4422445 4.3222345",
					"name":        "old name",
					"acronym":     []string{"n", "a", "m", "e"},
				},
			},
			dst: map[string]any{},
			mappers: []Mapper{
				Mapper{
					SrcJsonPath:     "$.area.coordinates",
					DstJsonPath:     "$.area.latitude",
					Transformations: []Transformation{{Trsnfmr: SplitTransformer{Delim: " ", Index: 0}}},
				},
				Mapper{
					SrcJsonPath:     "$.area.coordinates",
					DstJsonPath:     "$.area.longitude",
					Transformations: []Transformation{{Trsnfmr: SplitTransformer{Delim: " ", Index: 1}}},
				},
				Mapper{
					SrcJsonPath:     "$.area.name",
					DstJsonPath:     "$.area.name",
					Transformations: []Transformation{{Trsnfmr: ReplaceTransformer{OldVal: "old", NewVal: "new"}}},
				},
				Mapper{
					SrcJsonPath:     "$.area.acronym",
					DstJsonPath:     "$.area.name_from_acronym",
					Transformations: []Transformation{{Trsnfmr: JoinTransformer{Delim: ""}, AsArray: true}},
				},
			},
			expectedDst: map[string]any{
				"area": map[string]any{
					"latitude":          "45.4422445",
					"longitude":         "4.3222345",
					"name":              "new name",
					"name_from_acronym": "name",
				},
			},
			expectedErrorMessages: []string{},
		},
		{
			src: map[string]any{
				"books": []any{
					map[string]any{
						"title":   "Title 1",
						"summary": "lorem ipsum 23/12/1889",
					},
					map[string]any{
						"title":   "Title 2",
						"summary": "lorem ipsum 15/03/1733",
					},
					map[string]any{
						"title":   "Title 3",
						"summary": "lorem ipsum 12/01/1901",
					},
				},
			},
			dst: map[string]any{},
			mappers: []Mapper{
				Mapper{
					SrcJsonPath: "$.books.summary",
					DstJsonPath: "$.dates",
					Transformations: []Transformation{
						{Trsnfmr: StringMatchTransformer{Regex: `\d{2}/\d{2}/\d{4}`}},
						{Trsnfmr: SplitTransformer{Delim: "/", Index: 2}},
						{Trsnfmr: NumberTransformer{}},
					},
				},
			},
			expectedDst: map[string]any{
				"dates": []any{1889.0, 1733.0, 1901.0},
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
				},
			},
			dst: map[string]any{"authors": nil},
			mappers: []Mapper{
				Mapper{
					SrcJsonPath: "$.library.books.author",
					DstJsonPath: "$.authors",
					Transformations: []Transformation{
						Transformation{Trsnfmr: JoinTransformer{Delim: ", "}, AsArray: true},
					},
				},
			},
			expectedDst:           map[string]any{"authors": "Nietzsche, Stirner"},
			expectedErrorMessages: []string{},
		},
	}

	for i, tc := range cases {
		t.Run(fmt.Sprintf("[%v] Map(%v, %v, %v)=%v", i, tc.src, tc.dst, tc.mappers, tc.expectedErrorMessages), func(t *testing.T) {
			errors := Map(tc.src, tc.dst, tc.mappers)
			if len(errors) != len(tc.expectedErrorMessages) {
				t.Errorf("Expected error messages '%#v', but got '%#v'", tc.expectedErrorMessages, errors)
			}

			for i, err := range errors {
				if err == nil && len(tc.expectedErrorMessages[i]) > 0 {
					t.Errorf("Expected error message '%#v', but got '%#v'", tc.expectedErrorMessages[i], err)
				}
				if err != nil && err.Error() != tc.expectedErrorMessages[i] {
					t.Errorf("Expected error message '%#v', but got '%#v'", tc.expectedErrorMessages[i], err.Error())
				}
			}
			if !cmp.Equal(tc.expectedDst, tc.dst) {
				t.Errorf("Expected '%s', but got '%s'", prettify(tc.expectedDst), prettify(tc.dst))
			}
		})
	}
}
