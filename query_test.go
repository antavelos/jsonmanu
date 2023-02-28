package jsonmanu

import (
	"fmt"
	"testing"

	gu "github.com/antavelos/go-utils"
	cmp "github.com/google/go-cmp/cmp"
)

type SplitJsonPathTestCase struct {
	jsonPath       string
	expectedTokens []string
}

func TestSplitJsonPath(t *testing.T) {
	testCases := []SplitJsonPathTestCase{
		{
			jsonPath:       "$.library.books",
			expectedTokens: []string{"$", "library", "books"},
		},
		{
			jsonPath:       "$.library.books[*]",
			expectedTokens: []string{"$", "library", "books[*]"},
		},
		{
			jsonPath:       "$.library.books[1,2]",
			expectedTokens: []string{"$", "library", "books[1,2]"},
		},
		{
			jsonPath:       "$.library.books[2:4]",
			expectedTokens: []string{"$", "library", "books[2:4]"},
		},
		{
			jsonPath:       "$.library.books[?(@.price < 10)]",
			expectedTokens: []string{"$", "library", "books[?(@.price < 10)]"},
		},
		{
			jsonPath:       "$..books",
			expectedTokens: []string{"$", "", "books"},
		},
		{
			jsonPath:       "$..*",
			expectedTokens: []string{"$", "", "*"},
		},
	}
	for _, tc := range testCases {
		t.Run(fmt.Sprintf("splitJsonPath(%v)=%v", tc.jsonPath, tc.expectedTokens), func(t *testing.T) {
			tokens := splitJsonPath(tc.jsonPath)
			if !cmp.Equal(tc.expectedTokens, tokens) {
				t.Errorf("Expected tokens '%#v', but got '%#v'", tc.expectedTokens, tokens)
			}
		})
	}
}

type ParseJsonPathTestCase struct {
	jsonPath             string
	expectedNodes        []nodeDataAccessor
	expectedErrorMessage string
}

func TestParse(t *testing.T) {
	testCases := []ParseJsonPathTestCase{
		{
			jsonPath:             "books",
			expectedNodes:        nil,
			expectedErrorMessage: "JSONPath should start with '$.'",
		},
		{
			jsonPath:             "$.books.",
			expectedNodes:        nil,
			expectedErrorMessage: "JSONPath should not end with '.'",
		},
		{
			jsonPath:             "$.books. ",
			expectedNodes:        nil,
			expectedErrorMessage: "Couldn't parse JSONPath substring 1: ' '",
		},
		{
			jsonPath: "$.books[0]",
			expectedNodes: []nodeDataAccessor{
				arrayIndexedNode{
					node:    node{name: "books"},
					indices: []int{0},
				},
			},
			expectedErrorMessage: "",
		},
		{
			jsonPath: "$.library.books[*]",
			expectedNodes: []nodeDataAccessor{
				node{
					name: "library",
				},
				arrayIndexedNode{
					node: node{name: "books"},
				},
			},
			expectedErrorMessage: "",
		},
		{
			jsonPath: "$.library.books[0]",
			expectedNodes: []nodeDataAccessor{
				node{
					name: "library",
				},
				arrayIndexedNode{
					node:    node{name: "books"},
					indices: []int{0},
				},
			},
			expectedErrorMessage: "",
		},
		{
			jsonPath: "$.library.books[0,1,2]",
			expectedNodes: []nodeDataAccessor{
				node{
					name: "library",
				},
				arrayIndexedNode{
					node:    node{name: "books"},
					indices: []int{0, 1, 2},
				},
			},
			expectedErrorMessage: "",
		},
		{
			jsonPath: "$.library.books[1:]",
			expectedNodes: []nodeDataAccessor{
				node{
					name: "library",
				},
				arraySlicedNode{
					node:  node{name: "books"},
					start: 1,
				},
			},
			expectedErrorMessage: "",
		},
		{
			jsonPath: "$.library.books[1:2]",
			expectedNodes: []nodeDataAccessor{
				node{
					name: "library",
				},
				arraySlicedNode{
					node:  node{name: "books"},
					start: 1,
					end:   2,
				},
			},
			expectedErrorMessage: "",
		},
		{
			jsonPath: "$.library.books[:2]",
			expectedNodes: []nodeDataAccessor{
				node{
					name: "library",
				},
				arraySlicedNode{
					node: node{name: "books"},
					end:  2,
				},
			},
			expectedErrorMessage: "",
		},
		{
			jsonPath: "$.library.books[?(@.price < 10)]",
			expectedNodes: []nodeDataAccessor{
				node{
					name: "library",
				},
				arrayFilteredNode{
					node:  node{name: "books"},
					key:   "price",
					op:    "<",
					value: "10",
				},
			},
			expectedErrorMessage: "",
		},
		{
			jsonPath: "$.library.books[?(@.isbn)]",
			expectedNodes: []nodeDataAccessor{
				node{
					name: "library",
				},
				arrayFilteredNode{
					node:  node{name: "books"},
					key:   "isbn",
					op:    "",
					value: "",
				},
			},
			expectedErrorMessage: "",
		},
		{
			jsonPath: "$..books",
			expectedNodes: []nodeDataAccessor{
				node{
					name: "",
				},
				node{
					name: "books",
				},
			},
			expectedErrorMessage: "",
		},
		{
			jsonPath: "$..*",
			expectedNodes: []nodeDataAccessor{
				node{
					name: "",
				},
				node{
					name: "*",
				},
			},
			expectedErrorMessage: "",
		},
	}

	for _, tc := range testCases {
		t.Run(fmt.Sprintf("parseJsonPAth(%v)=%v, %v", tc.jsonPath, tc.expectedNodes, tc.expectedErrorMessage), func(t *testing.T) {
			nodes, err := parseJsonPath(tc.jsonPath)
			if !cmp.Equal(tc.expectedNodes, nodes, cmp.AllowUnexported(node{}, arrayIndexedNode{}, arrayFilteredNode{}, arraySlicedNode{})) {
				t.Errorf("Expected nodes '%#v', but got '%#v'", tc.expectedNodes, nodes)
			}
			if (err == nil && len(tc.expectedErrorMessage) > 0) || (err != nil && err.Error() != tc.expectedErrorMessage) {
				t.Errorf("Expected error message '%#v', but got '%#v'", tc.expectedErrorMessage, err.Error())
			}
		})
	}
}

type GetTestCase struct {
	jsonPath             string
	data                 map[string]any
	expectedErrorMessage string
	expectedData         any
}

func TestGet(t *testing.T) {
	testCases := []GetTestCase{
		{
			jsonPath: ".books",
			data: map[string]any{
				"books": []any{
					map[string]any{"author": "Nietzsche", "title": "Book1"},
					map[string]any{"author": "Nietzsche", "title": "Book2"},
					map[string]any{"author": "Nietzsche", "title": "Book3"},
				},
			},
			expectedErrorMessage: "JSONPath should start with '$.'",
			expectedData:         nil,
		},
		{
			jsonPath: "$.books.",
			data: map[string]any{
				"books": []any{
					map[string]any{"author": "Nietzsche", "title": "Book1"},
					map[string]any{"author": "Nietzsche", "title": "Book2"},
					map[string]any{"author": "Nietzsche", "title": "Book3"},
				},
			},
			expectedErrorMessage: "JSONPath should not end with '.'",
			expectedData:         nil,
		},
		{
			jsonPath: "$.books",
			data: map[string]any{
				"books": []any{
					map[string]any{"author": "Nietzsche", "title": "Book1"},
					map[string]any{"author": "Nietzsche", "title": "Book2"},
					map[string]any{"author": "Nietzsche", "title": "Book3"},
				},
			},
			expectedErrorMessage: "",
			expectedData: []any{
				map[string]any{"author": "Nietzsche", "title": "Book1"},
				map[string]any{"author": "Nietzsche", "title": "Book2"},
				map[string]any{"author": "Nietzsche", "title": "Book3"},
			},
		},
		{
			jsonPath: "$.store.books.*",
			data: map[string]any{
				"store": map[string]any{
					"books": []any{
						map[string]any{"author": "Nietzsche", "title": "Book1"},
						map[string]any{"author": "Nietzsche", "title": "Book2"},
						map[string]any{"author": "Nietzsche", "title": "Book3"},
					},
				},
			},
			expectedErrorMessage: "",
			expectedData: []any{
				map[string]any{"author": "Nietzsche", "title": "Book1"},
				map[string]any{"author": "Nietzsche", "title": "Book2"},
				map[string]any{"author": "Nietzsche", "title": "Book3"},
			},
		},
		{
			jsonPath: "$.books.*",
			data: map[string]any{
				"books": []any{
					map[string]any{"author": "Nietzsche", "title": "Book1"},
					map[string]any{"author": "Nietzsche", "title": "Book2"},
					map[string]any{"author": "Nietzsche", "title": "Book3"},
				},
			},
			expectedErrorMessage: "",
			expectedData: []any{
				map[string]any{"author": "Nietzsche", "title": "Book1"},
				map[string]any{"author": "Nietzsche", "title": "Book2"},
				map[string]any{"author": "Nietzsche", "title": "Book3"},
			},
		},
		{
			jsonPath: "$.store.books[*]",
			data: map[string]any{
				"store": map[string]any{
					"books": []any{
						map[string]any{"author": "Nietzsche", "title": "Book1"},
						map[string]any{"author": "Nietzsche", "title": "Book2"},
						map[string]any{"author": "Nietzsche", "title": "Book3"},
					},
				},
			},
			expectedErrorMessage: "",
			expectedData: []any{
				map[string]any{"author": "Nietzsche", "title": "Book1"},
				map[string]any{"author": "Nietzsche", "title": "Book2"},
				map[string]any{"author": "Nietzsche", "title": "Book3"},
			},
		},
		{
			jsonPath: "$.books[*]",
			data: map[string]any{
				"books": []any{
					map[string]any{"author": "Nietzsche", "title": "Book1"},
					map[string]any{"author": "Nietzsche", "title": "Book2"},
					map[string]any{"author": "Nietzsche", "title": "Book3"},
				},
			},
			expectedErrorMessage: "",
			expectedData: []any{
				map[string]any{"author": "Nietzsche", "title": "Book1"},
				map[string]any{"author": "Nietzsche", "title": "Book2"},
				map[string]any{"author": "Nietzsche", "title": "Book3"},
			},
		},
		{
			jsonPath: "$..books[*]",
			data: map[string]any{
				"books": []any{
					map[string]any{"author": "Nietzsche", "title": "Book1"},
					map[string]any{"author": "Nietzsche", "title": "Book2"},
					map[string]any{"author": "Nietzsche", "title": "Book3"},
				},
			},
			expectedErrorMessage: "",
			expectedData: []any{
				map[string]any{"author": "Nietzsche", "title": "Book1"},
				map[string]any{"author": "Nietzsche", "title": "Book2"},
				map[string]any{"author": "Nietzsche", "title": "Book3"},
			},
		},
		{
			jsonPath: "$..books.*",
			data: map[string]any{
				"books": []any{
					map[string]any{"author": "Nietzsche", "title": "Book1"},
					map[string]any{"author": "Nietzsche", "title": "Book2"},
					map[string]any{"author": "Nietzsche", "title": "Book3"},
				},
			},
			expectedErrorMessage: "",
			expectedData: []any{
				map[string]any{"author": "Nietzsche", "title": "Book1"},
				map[string]any{"author": "Nietzsche", "title": "Book2"},
				map[string]any{"author": "Nietzsche", "title": "Book3"},
			},
		},
		{
			jsonPath: "$.books[0]",
			data: map[string]any{
				"books": []any{
					map[string]any{"author": "Nietzsche", "title": "Book1"},
					map[string]any{"author": "Nietzsche", "title": "Book2"},
					map[string]any{"author": "Nietzsche", "title": "Book3"},
				},
			},
			expectedErrorMessage: "",
			expectedData: []any{
				map[string]any{"author": "Nietzsche", "title": "Book1"},
			},
		},
		{
			jsonPath: "$.books[0,2]",
			data: map[string]any{
				"books": []any{
					map[string]any{"author": "Nietzsche", "title": "Book1"},
					map[string]any{"author": "Nietzsche", "title": "Book2"},
					map[string]any{"author": "Nietzsche", "title": "Book3"},
				},
			},
			expectedErrorMessage: "",
			expectedData: []any{
				map[string]any{"author": "Nietzsche", "title": "Book1"},
				map[string]any{"author": "Nietzsche", "title": "Book3"},
			},
		},
		{
			jsonPath: "$.books[1:]",
			data: map[string]any{
				"books": []any{
					map[string]any{"author": "Nietzsche", "title": "Book1"},
					map[string]any{"author": "Nietzsche", "title": "Book2"},
					map[string]any{"author": "Nietzsche", "title": "Book3"},
				},
			},
			expectedErrorMessage: "",
			expectedData: []any{
				map[string]any{"author": "Nietzsche", "title": "Book2"},
				map[string]any{"author": "Nietzsche", "title": "Book3"},
			},
		},
		{
			jsonPath: "$.books[1:3]",
			data: map[string]any{
				"books": []any{
					map[string]any{"author": "Nietzsche", "title": "Book1"},
					map[string]any{"author": "Nietzsche", "title": "Book2"},
					map[string]any{"author": "Nietzsche", "title": "Book3"},
					map[string]any{"author": "Nietzsche", "title": "Book4"},
				},
			},
			expectedErrorMessage: "",
			expectedData: []any{
				map[string]any{"author": "Nietzsche", "title": "Book2"},
				map[string]any{"author": "Nietzsche", "title": "Book3"},
			},
		},
		{
			jsonPath: "$.books[:3]",
			data: map[string]any{
				"books": []any{
					map[string]any{"author": "Nietzsche", "title": "Book1"},
					map[string]any{"author": "Nietzsche", "title": "Book2"},
					map[string]any{"author": "Nietzsche", "title": "Book3"},
					map[string]any{"author": "Nietzsche", "title": "Book4"},
				},
			},
			expectedErrorMessage: "",
			expectedData: []any{
				map[string]any{"author": "Nietzsche", "title": "Book1"},
				map[string]any{"author": "Nietzsche", "title": "Book2"},
				map[string]any{"author": "Nietzsche", "title": "Book3"},
			},
		},
		{
			jsonPath: "$.books[?(@.price > 10)]",
			data: map[string]any{
				"books": []any{
					map[string]any{"author": "Nietzsche", "title": "Book1", "price": 15},
					map[string]any{"author": "Nietzsche", "title": "Book2", "price": 20},
					map[string]any{"author": "Nietzsche", "title": "Book3", "price": 5},
					map[string]any{"author": "Nietzsche", "title": "Book4", "price": 10},
				},
			},
			expectedErrorMessage: "",
			expectedData: []any{
				map[string]any{"author": "Nietzsche", "title": "Book1", "price": 15},
				map[string]any{"author": "Nietzsche", "title": "Book2", "price": 20},
			},
		},
		{
			jsonPath: "$.books[?(@.price >= 10)]",
			data: map[string]any{
				"books": []any{
					map[string]any{"author": "Nietzsche", "title": "Book1", "price": 15},
					map[string]any{"author": "Nietzsche", "title": "Book2", "price": 20},
					map[string]any{"author": "Nietzsche", "title": "Book3", "price": 5},
					map[string]any{"author": "Nietzsche", "title": "Book4", "price": 10},
				},
			},
			expectedErrorMessage: "",
			expectedData: []any{
				map[string]any{"author": "Nietzsche", "title": "Book1", "price": 15},
				map[string]any{"author": "Nietzsche", "title": "Book2", "price": 20},
				map[string]any{"author": "Nietzsche", "title": "Book4", "price": 10},
			},
		},
		{
			jsonPath: "$.books[?(@.price < 10)]",
			data: map[string]any{
				"books": []any{
					map[string]any{"author": "Nietzsche", "title": "Book1", "price": 15},
					map[string]any{"author": "Nietzsche", "title": "Book2", "price": 20},
					map[string]any{"author": "Nietzsche", "title": "Book3", "price": 5},
					map[string]any{"author": "Nietzsche", "title": "Book4", "price": 10},
				},
			},
			expectedErrorMessage: "",
			expectedData: []any{
				map[string]any{"author": "Nietzsche", "title": "Book3", "price": 5},
			},
		},
		{
			jsonPath: "$.books[?(@.price <= 10)]",
			data: map[string]any{
				"books": []any{
					map[string]any{"author": "Nietzsche", "title": "Book1", "price": 15},
					map[string]any{"author": "Nietzsche", "title": "Book2", "price": 20},
					map[string]any{"author": "Nietzsche", "title": "Book3", "price": 5},
					map[string]any{"author": "Nietzsche", "title": "Book4", "price": 10},
				},
			},
			expectedErrorMessage: "",
			expectedData: []any{
				map[string]any{"author": "Nietzsche", "title": "Book3", "price": 5},
				map[string]any{"author": "Nietzsche", "title": "Book4", "price": 10},
			},
		},
		{
			jsonPath: "$.books[?(@.price == 10)]",
			data: map[string]any{
				"books": []any{
					map[string]any{"author": "Nietzsche", "title": "Book1", "price": 15},
					map[string]any{"author": "Nietzsche", "title": "Book2", "price": 20},
					map[string]any{"author": "Nietzsche", "title": "Book3", "price": 5},
					map[string]any{"author": "Nietzsche", "title": "Book4", "price": 10},
				},
			},
			expectedErrorMessage: "",
			expectedData: []any{
				map[string]any{"author": "Nietzsche", "title": "Book4", "price": 10},
			},
		},
		{
			jsonPath: "$.books[?(@.price > 10)].author",
			data: map[string]any{
				"books": []any{
					map[string]any{"author": "Nietzsche", "title": "Book1", "price": 15},
					map[string]any{"author": "Nietzsche", "title": "Book2", "price": 20},
					map[string]any{"author": "Stirner", "title": "Book3", "price": 5},
					map[string]any{"author": "Stirner", "title": "Book4", "price": 10},
				},
			},
			expectedErrorMessage: "",
			expectedData:         []any{"Nietzsche", "Nietzsche"},
		},
		{
			jsonPath: "$.books[?(@.author == Nietzsche)]",
			data: map[string]any{
				"books": []any{
					map[string]any{"author": "Nietzsche", "title": "Book1", "price": 15},
					map[string]any{"author": "Nietzsche", "title": "Book2", "price": 20},
					map[string]any{"author": "Stirner", "title": "Book1", "price": 15},
					map[string]any{"author": "Nietzsche", "title": "Book3", "price": 5},
					map[string]any{"author": "Stirner", "title": "Book2", "price": 10},
					map[string]any{"author": "Nietzsche", "title": "Book4", "price": 10},
				},
			},
			expectedErrorMessage: "",
			expectedData: []any{
				map[string]any{"author": "Nietzsche", "title": "Book1", "price": 15},
				map[string]any{"author": "Nietzsche", "title": "Book2", "price": 20},
				map[string]any{"author": "Nietzsche", "title": "Book3", "price": 5},
				map[string]any{"author": "Nietzsche", "title": "Book4", "price": 10},
			},
		},
		{
			jsonPath: "$.books[?(@.price)]",
			data: map[string]any{
				"books": []any{
					map[string]any{"author": "Nietzsche", "title": "Book1"},
					map[string]any{"author": "Nietzsche", "title": "Book2", "price": 20},
					map[string]any{"author": "Nietzsche", "title": "Book3"},
					map[string]any{"author": "Nietzsche", "title": "Book4", "price": 10},
				},
			},
			expectedErrorMessage: "",
			expectedData: []any{
				map[string]any{"author": "Nietzsche", "title": "Book2", "price": 20},
				map[string]any{"author": "Nietzsche", "title": "Book4", "price": 10},
			},
		},
		{
			jsonPath: "$.books[*].author",
			data: map[string]any{
				"books": []any{
					map[string]any{"author": "Nietzsche", "title": "Book1", "price": 15},
					map[string]any{"author": "Nietzsche", "title": "Book2", "price": 20},
					map[string]any{"author": "Stirner", "title": "Book1", "price": 15},
					map[string]any{"author": "Nietzsche", "title": "Book3", "price": 5},
					map[string]any{"author": "Stirner", "title": "Book2", "price": 10},
					map[string]any{"author": "Nietzsche", "title": "Book4", "price": 10},
				},
			},
			expectedErrorMessage: "",
			expectedData:         []any{"Nietzsche", "Nietzsche", "Stirner", "Nietzsche", "Stirner", "Nietzsche"},
		},
		{
			jsonPath: "$.books.author",
			data: map[string]any{
				"books": []any{
					map[string]any{"author": "Nietzsche", "title": "Book1", "price": 15},
					map[string]any{"author": "Nietzsche", "title": "Book2", "price": 20},
					map[string]any{"author": "Stirner", "title": "Book1", "price": 15},
					map[string]any{"author": "Nietzsche", "title": "Book3", "price": 5},
					map[string]any{"author": "Stirner", "title": "Book2", "price": 10},
					map[string]any{"author": "Nietzsche", "title": "Book4", "price": 10},
				},
			},
			expectedErrorMessage: "",
			expectedData:         []any{"Nietzsche", "Nietzsche", "Stirner", "Nietzsche", "Stirner", "Nietzsche"},
		},
		{
			jsonPath: "$.store..author",
			data: map[string]any{
				"store": map[string]any{
					"books": []any{
						map[string]any{"author": "Nietzsche", "title": "Book1", "price": 15},
						map[string]any{"author": "Nietzsche", "title": "Book2", "price": 20},
						map[string]any{"author": "Stirner", "title": "Book1", "price": 15},
						map[string]any{"author": "Nietzsche", "title": "Book3", "price": 5},
						map[string]any{"author": "Stirner", "title": "Book2", "price": 10},
						map[string]any{"author": "Nietzsche", "title": "Book4", "price": 10},
					},
				},
			},
			expectedErrorMessage: "",
			expectedData:         []any{"Nietzsche", "Nietzsche", "Stirner", "Nietzsche", "Stirner", "Nietzsche"},
		},
		{
			jsonPath: "$..author",
			data: map[string]any{
				"store": map[string]any{
					"book": map[string]any{
						"author": "Nietzsche",
					},
				},
			},
			expectedErrorMessage: "",
			expectedData:         []any{"Nietzsche"},
		},
		{
			jsonPath: "$..author",
			data: map[string]any{
				"store": map[string]any{
					"books": []any{
						map[string]any{"author": "Nietzsche", "title": "Book1", "price": 15},
						map[string]any{"author": "Nietzsche", "title": "Book2", "price": 20},
						map[string]any{"author": "Stirner", "title": "Book1", "price": 15},
						map[string]any{"author": "Nietzsche", "title": "Book3", "price": 5},
						map[string]any{"author": "Stirner", "title": "Book2", "price": 10},
						map[string]any{"author": "Nietzsche", "title": "Book4", "price": 10},
					},
				},
			},
			expectedErrorMessage: "",
			expectedData:         []any{"Nietzsche", "Nietzsche", "Stirner", "Nietzsche", "Stirner", "Nietzsche"},
		},
		{
			jsonPath: "$..price",
			data: map[string]any{
				"store": map[string]any{
					"books": []any{
						map[string]any{"author": "Nietzsche", "title": "Book1", "price": 15},
						map[string]any{"author": "Nietzsche", "title": "Book2", "price": 20},
						map[string]any{"author": "Stirner", "title": "Book1", "price": 15},
						map[string]any{"author": "Nietzsche", "title": "Book3", "price": 5},
						map[string]any{"author": "Stirner", "title": "Book2", "price": 10},
						map[string]any{"author": "Nietzsche", "title": "Book4", "price": 10},
					},
				},
			},
			expectedErrorMessage: "",
			expectedData:         []any{15, 20, 15, 5, 10, 10},
		},
		{
			jsonPath: "$..books[0].author",
			data: map[string]any{
				"store": map[string]any{
					"books": []any{
						map[string]any{"author": "Nietzsche", "title": "Book1", "price": 15},
						map[string]any{"author": "Nietzsche", "title": "Book2", "price": 20},
						map[string]any{"author": "Stirner", "title": "Book1", "price": 15},
						map[string]any{"author": "Nietzsche", "title": "Book3", "price": 5},
						map[string]any{"author": "Stirner", "title": "Book2", "price": 10},
						map[string]any{"author": "Nietzsche", "title": "Book4", "price": 10},
					},
				},
			},
			expectedErrorMessage: "",
			expectedData:         []any{"Nietzsche"},
		},
	}

	for i, tc := range testCases {

		t.Run(fmt.Sprintf("(%v) - Get(%v, %v)=%v, %v", i, tc.data, tc.jsonPath, tc.expectedData, tc.expectedErrorMessage),
			func(t *testing.T) {
				data, err := Get(tc.data, tc.jsonPath)
				if (err == nil && len(tc.expectedErrorMessage) > 0) || (err != nil && err.Error() != tc.expectedErrorMessage) {
					t.Errorf("Expected error message '%#v', but got '%#v'", tc.expectedErrorMessage, err.Error())
				}
				if !cmp.Equal(tc.expectedData, data) {
					t.Errorf("\n(%v) - Expected:\n '%#v\nbut got\n'%#v'", i, tc.expectedData, data)
				}
			})
	}
}

type PutTestCase struct {
	jsonPath             string
	data                 map[string]any
	value                any
	expectedErrorMessage string
	expectedUpdatedData  map[string]any
}

func TestPut(t *testing.T) {
	testCases := []PutTestCase{
		{
			jsonPath: ".books",
			data: map[string]any{
				"books": []any{
					map[string]any{"author": "Nietzsche", "title": "Book1"},
					map[string]any{"author": "Nietzsche", "title": "Book2"},
					map[string]any{"author": "Nietzsche", "title": "Book3"},
				},
			},
			value:                1,
			expectedErrorMessage: "JSONPath should start with '$.'",
			expectedUpdatedData: map[string]any{
				"books": []any{
					map[string]any{"author": "Nietzsche", "title": "Book1"},
					map[string]any{"author": "Nietzsche", "title": "Book2"},
					map[string]any{"author": "Nietzsche", "title": "Book3"},
				},
			},
		},
		{
			jsonPath: "$.books.",
			data: map[string]any{
				"books": []any{
					map[string]any{"author": "Nietzsche", "title": "Book1"},
					map[string]any{"author": "Nietzsche", "title": "Book2"},
					map[string]any{"author": "Nietzsche", "title": "Book3"},
				},
			},
			value:                1,
			expectedErrorMessage: "JSONPath should not end with '.'",
			expectedUpdatedData: map[string]any{
				"books": []any{
					map[string]any{"author": "Nietzsche", "title": "Book1"},
					map[string]any{"author": "Nietzsche", "title": "Book2"},
					map[string]any{"author": "Nietzsche", "title": "Book3"},
				},
			},
		},
		{
			jsonPath: "$.book.author",
			data: map[string]any{
				"book": map[string]any{
					"author": "Someone",
				},
			},
			value:                "Someone else",
			expectedErrorMessage: "",
			expectedUpdatedData: map[string]any{
				"book": map[string]any{
					"author": "Someone else",
				},
			},
		},
		{
			jsonPath: "$.book[*].author",
			data: map[string]any{
				"book": []any{
					map[string]any{"author": "Someone"},
					map[string]any{"author": "Someone"},
				},
			},
			value:                "Someone else",
			expectedErrorMessage: "",
			expectedUpdatedData: map[string]any{
				"book": []any{
					map[string]any{"author": "Someone else"},
					map[string]any{"author": "Someone else"},
				},
			},
		},
		{
			jsonPath: "$.book[0].author",
			data: map[string]any{
				"book": []any{
					map[string]any{"author": "Someone"},
					map[string]any{"author": "Someone"},
				},
			},
			value:                "Someone else",
			expectedErrorMessage: "",
			expectedUpdatedData: map[string]any{
				"book": []any{
					map[string]any{"author": "Someone else"},
					map[string]any{"author": "Someone"},
				},
			},
		},
		{
			jsonPath: "$.book[0, 1].author",
			data: map[string]any{
				"book": []any{
					map[string]any{"author": "Someone"},
					map[string]any{"author": "Someone"},
				},
			},
			value:                "Someone else",
			expectedErrorMessage: "",
			expectedUpdatedData: map[string]any{
				"book": []any{
					map[string]any{"author": "Someone else"},
					map[string]any{"author": "Someone else"},
				},
			},
		},
		{
			jsonPath: "$.book[:1].author",
			data: map[string]any{
				"book": []any{
					map[string]any{"author": "Someone"},
					map[string]any{"author": "Someone"},
					map[string]any{"author": "Someone"},
				},
			},
			value:                "Someone else",
			expectedErrorMessage: "",
			expectedUpdatedData: map[string]any{
				"book": []any{
					map[string]any{"author": "Someone else"},
					map[string]any{"author": "Someone"},
					map[string]any{"author": "Someone"},
				},
			},
		},
		{
			jsonPath: "$.book[1:].author",
			data: map[string]any{
				"book": []any{
					map[string]any{"author": "Someone"},
					map[string]any{"author": "Someone"},
					map[string]any{"author": "Someone"},
				},
			},
			value:                "Someone else",
			expectedErrorMessage: "",
			expectedUpdatedData: map[string]any{
				"book": []any{
					map[string]any{"author": "Someone"},
					map[string]any{"author": "Someone else"},
					map[string]any{"author": "Someone else"},
				},
			},
		},
		{
			jsonPath: "$.book[1:2].author",
			data: map[string]any{
				"book": []any{
					map[string]any{"author": "Someone"},
					map[string]any{"author": "Someone"},
					map[string]any{"author": "Someone"},
				},
			},
			value:                "Someone else",
			expectedErrorMessage: "",
			expectedUpdatedData: map[string]any{
				"book": []any{
					map[string]any{"author": "Someone"},
					map[string]any{"author": "Someone else"},
					map[string]any{"author": "Someone"},
				},
			},
		},
		{
			jsonPath: "$.store.books[?(@.author == Nietzsche)].price",
			data: map[string]any{
				"store": map[string]any{
					"books": []any{
						map[string]any{"author": "Nietzsche", "title": "Book1", "price": 15},
						map[string]any{"author": "Nietzsche", "title": "Book2", "price": 20},
						map[string]any{"author": "Stirner", "title": "Book1", "price": 15},
						map[string]any{"author": "Nietzsche", "title": "Book3", "price": 5},
						map[string]any{"author": "Stirner", "title": "Book2", "price": 10},
						map[string]any{"author": "Nietzsche", "title": "Book4", "price": 10},
					},
				},
			},
			value:                5,
			expectedErrorMessage: "",
			expectedUpdatedData: map[string]any{
				"store": map[string]any{
					"books": []any{
						map[string]any{"author": "Nietzsche", "title": "Book1", "price": 5},
						map[string]any{"author": "Nietzsche", "title": "Book2", "price": 5},
						map[string]any{"author": "Stirner", "title": "Book1", "price": 15},
						map[string]any{"author": "Nietzsche", "title": "Book3", "price": 5},
						map[string]any{"author": "Stirner", "title": "Book2", "price": 10},
						map[string]any{"author": "Nietzsche", "title": "Book4", "price": 5},
					},
				},
			},
		},
		{
			jsonPath: "$..books[?(@.author == Nietzsche)].price",
			data: map[string]any{
				"store": map[string]any{
					"books": []any{
						map[string]any{"author": "Nietzsche", "title": "Book1", "price": 15},
						map[string]any{"author": "Nietzsche", "title": "Book2", "price": 20},
						map[string]any{"author": "Stirner", "title": "Book1", "price": 15},
						map[string]any{"author": "Nietzsche", "title": "Book3", "price": 5},
						map[string]any{"author": "Stirner", "title": "Book2", "price": 10},
						map[string]any{"author": "Nietzsche", "title": "Book4", "price": 10},
					},
				},
			},
			value:                5,
			expectedErrorMessage: "",
			expectedUpdatedData: map[string]any{
				"store": map[string]any{
					"books": []any{
						map[string]any{"author": "Nietzsche", "title": "Book1", "price": 5},
						map[string]any{"author": "Nietzsche", "title": "Book2", "price": 5},
						map[string]any{"author": "Stirner", "title": "Book1", "price": 15},
						map[string]any{"author": "Nietzsche", "title": "Book3", "price": 5},
						map[string]any{"author": "Stirner", "title": "Book2", "price": 10},
						map[string]any{"author": "Nietzsche", "title": "Book4", "price": 5},
					},
				},
			},
		},
		{
			jsonPath: "$.store..books[?(@.author == Nietzsche)].price",
			data: map[string]any{
				"store": map[string]any{
					"library": map[string]any{
						"books": []any{
							map[string]any{"author": "Nietzsche", "title": "Book1", "price": 15},
							map[string]any{"author": "Nietzsche", "title": "Book2", "price": 20},
							map[string]any{"author": "Stirner", "title": "Book1", "price": 15},
							map[string]any{"author": "Nietzsche", "title": "Book3", "price": 5},
							map[string]any{"author": "Stirner", "title": "Book2", "price": 10},
							map[string]any{"author": "Nietzsche", "title": "Book4", "price": 10},
						},
					},
				},
			},
			value:                5,
			expectedErrorMessage: "",
			expectedUpdatedData: map[string]any{
				"store": map[string]any{
					"library": map[string]any{
						"books": []any{
							map[string]any{"author": "Nietzsche", "title": "Book1", "price": 5},
							map[string]any{"author": "Nietzsche", "title": "Book2", "price": 5},
							map[string]any{"author": "Stirner", "title": "Book1", "price": 15},
							map[string]any{"author": "Nietzsche", "title": "Book3", "price": 5},
							map[string]any{"author": "Stirner", "title": "Book2", "price": 10},
							map[string]any{"author": "Nietzsche", "title": "Book4", "price": 5},
						},
					},
				},
			},
		},
		{
			jsonPath: "$..author",
			data: map[string]any{
				"book": []any{
					map[string]any{"author": "Someone"},
					map[string]any{"author": "Someone"},
				},
			},
			value:                "Someone else",
			expectedErrorMessage: "",
			expectedUpdatedData: map[string]any{
				"book": []any{
					map[string]any{"author": "Someone else"},
					map[string]any{"author": "Someone else"},
				},
			},
		},
		{
			jsonPath: "$..author",
			data: map[string]any{
				"store1": map[string]any{
					"books": []any{
						map[string]any{"author": "Someone"},
						map[string]any{"author": "Someone"},
					},
				},
				"store2": map[string]any{
					"books": []any{
						map[string]any{"author": "Someone"},
						map[string]any{"author": "Someone"},
					},
				},
			},
			value:                "Someone else",
			expectedErrorMessage: "",
			expectedUpdatedData: map[string]any{
				"store1": map[string]any{
					"books": []any{
						map[string]any{"author": "Someone else"},
						map[string]any{"author": "Someone else"},
					},
				},
				"store2": map[string]any{
					"books": []any{
						map[string]any{"author": "Someone else"},
						map[string]any{"author": "Someone else"},
					},
				},
			},
		},
		{
			jsonPath: "$..price",
			data: map[string]any{
				"store": map[string]any{
					"books": []any{
						map[string]any{"author": "Nietzsche", "title": "Book1", "price": 15},
						map[string]any{"author": "Nietzsche", "title": "Book2", "price": 20},
						map[string]any{"author": "Stirner", "title": "Book1", "price": 15},
						map[string]any{"author": "Nietzsche", "title": "Book3", "price": 5},
						map[string]any{"author": "Stirner", "title": "Book2", "price": 10},
						map[string]any{"author": "Nietzsche", "title": "Book4", "price": 10},
					},
				},
			},
			value:                5,
			expectedErrorMessage: "",
			expectedUpdatedData: map[string]any{
				"store": map[string]any{
					"books": []any{
						map[string]any{"author": "Nietzsche", "title": "Book1", "price": 5},
						map[string]any{"author": "Nietzsche", "title": "Book2", "price": 5},
						map[string]any{"author": "Stirner", "title": "Book1", "price": 5},
						map[string]any{"author": "Nietzsche", "title": "Book3", "price": 5},
						map[string]any{"author": "Stirner", "title": "Book2", "price": 5},
						map[string]any{"author": "Nietzsche", "title": "Book4", "price": 5},
					},
				},
			},
		},
		{
			jsonPath: "$.store..price",
			data: map[string]any{
				"store": map[string]any{
					"books": []any{
						map[string]any{"author": "Nietzsche", "title": "Book1", "price": 15},
						map[string]any{"author": "Nietzsche", "title": "Book2", "price": 20},
						map[string]any{"author": "Stirner", "title": "Book1", "price": 15},
						map[string]any{"author": "Nietzsche", "title": "Book3", "price": 5},
						map[string]any{"author": "Stirner", "title": "Book2", "price": 10},
						map[string]any{"author": "Nietzsche", "title": "Book4", "price": 10},
					},
				},
			},
			value:                5,
			expectedErrorMessage: "",
			expectedUpdatedData: map[string]any{
				"store": map[string]any{
					"books": []any{
						map[string]any{"author": "Nietzsche", "title": "Book1", "price": 5},
						map[string]any{"author": "Nietzsche", "title": "Book2", "price": 5},
						map[string]any{"author": "Stirner", "title": "Book1", "price": 5},
						map[string]any{"author": "Nietzsche", "title": "Book3", "price": 5},
						map[string]any{"author": "Stirner", "title": "Book2", "price": 5},
						map[string]any{"author": "Nietzsche", "title": "Book4", "price": 5},
					},
				},
			},
		},
	}

	for i, tc := range testCases {
		t.Run(fmt.Sprintf("(%v) - Put(%v, %v, %v)=%v", i, tc.data, tc.jsonPath, tc.value, tc.expectedErrorMessage), func(t *testing.T) {
			err := Put(tc.data, tc.jsonPath, tc.value)
			if (err == nil && len(tc.expectedErrorMessage) > 0) || (err != nil && err.Error() != tc.expectedErrorMessage) {
				t.Errorf("Expected error message '%#v', but got '%#v'", tc.expectedErrorMessage, err.Error())
			}
			if !cmp.Equal(tc.expectedUpdatedData, tc.data) {
				t.Errorf("Expected '%#s', but got '%#s'", gu.Prettify(tc.expectedUpdatedData), gu.Prettify(tc.data))
			}
		})
	}
}
