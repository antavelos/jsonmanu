package jsonmanu

import (
	"fmt"
	"testing"

	"github.com/google/go-cmp/cmp"
)

type MapTestCase struct {
	src                   map[string]any
	dst                   map[string]any
	mappers               []Mapper
	expectedDst           map[string]any
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
					SrcJsonPath: "$.invalid.path",
					DstJsonPath: "$.library.authors",
				},
				Mapper{
					SrcJsonPath: "$.invalid.path",
					DstJsonPath: "$.library.categories",
				},
			},
			expectedDst: nil,
			expectedErrorMessages: []string{
				"Mapper[0]: Error while getting value from source: DataValidationError: Source key not found: 'invalid'",
				"Mapper[1]: Error while getting value from source: DataValidationError: Source key not found: 'invalid'",
			},
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
				"Mapper[0]: Error while putting value in destination: DataValidationError: Data is nil.",
				"Mapper[1]: Error while putting value in destination: DataValidationError: Data is nil.",
			},
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
					DstJsonPath: "$..authors",
					Transformations: []Transformation{
						Transformation{Trsnfmr: JoinTransformer{Delim: ", "}, AsArray: true},
					},
				},
			},
			expectedDst:           map[string]any{"authors": nil},
			expectedErrorMessages: []string{"Mapper[0]: Validation error: Reccursive descent not allowed in destination path."},
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
						Transformation{Trsnfmr: NumberTransformer{}},
					},
				},
			},
			expectedDst:           map[string]any{"authors": nil},
			expectedErrorMessages: []string{"Mapper[0]: Transformation[0] (jsonmanu.NumberTransformer): Array[0]: Couldn't convert value to number."},
		},
		{
			src: map[string]any{
				"library": map[string]any{
					"books": []any{
						map[string]any{"author": "Nietzsche", "price": 10},
						map[string]any{"author": "Stirner", "price": 15},
					},
				},
			},
			dst: map[string]any{"authors": nil},
			mappers: []Mapper{
				Mapper{
					SrcJsonPath: "$.library.books.price",
					DstJsonPath: "$.prices",
					Transformations: []Transformation{
						Transformation{Trsnfmr: StringMatchTransformer{}},
					},
				},
			},
			expectedDst:           map[string]any{"authors": nil},
			expectedErrorMessages: []string{"Mapper[0]: Transformation[0] (jsonmanu.StringMatchTransformer): Array[0]: Value is not a string."},
		},
		{
			src: map[string]any{
				"library": map[string]any{
					"books": []any{
						map[string]any{"author": "Nietzsche", "price": 10},
						map[string]any{"author": "Stirner", "price": 15},
					},
				},
			},
			dst: map[string]any{"authors": nil},
			mappers: []Mapper{
				Mapper{
					SrcJsonPath: "$.library.books.price",
					DstJsonPath: "$.prices",
					Transformations: []Transformation{
						Transformation{Trsnfmr: SubStrTransformer{}},
					},
				},
			},
			expectedDst:           map[string]any{"authors": nil},
			expectedErrorMessages: []string{"Mapper[0]: Transformation[0] (jsonmanu.SubStrTransformer): Array[0]: Value is not a string."},
		},
	}[5:7]

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
