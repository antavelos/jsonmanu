package jsonmanu

import (
	"fmt"
	"testing"

	cmp "github.com/google/go-cmp/cmp"
)

type SplitJsonPAthTestCase struct {
	path           string
	expectedTokens []string
}

func TestSplitJsonPath(t *testing.T) {
	testCases := []SplitJsonPAthTestCase{
		{
			path:           "$.library.books",
			expectedTokens: []string{"$", "library", "books"},
		},
		{
			path:           "$.library.books[*]",
			expectedTokens: []string{"$", "library", "books[*]"},
		},
		{
			path:           "$.library.books[1,2]",
			expectedTokens: []string{"$", "library", "books[1,2]"},
		},
		{
			path:           "$.library.books[2:4]",
			expectedTokens: []string{"$", "library", "books[2:4]"},
		},
		{
			path:           "$.library.books[?(@.price < 10)]",
			expectedTokens: []string{"$", "library", "books[?(@.price < 10)]"},
		},
		{
			path:           "$..books",
			expectedTokens: []string{"$", "", "books"},
		},
		{
			path:           "$..*",
			expectedTokens: []string{"$", "", "*"},
		},
	}
	for _, tc := range testCases {
		t.Run(fmt.Sprintf("jsonmanu.splitJsonPath(%v)=%v", tc.path, tc.expectedTokens), func(t *testing.T) {
			tokens := splitJsonPath(tc.path)
			if !cmp.Equal(tc.expectedTokens, tokens) {
				t.Errorf("Expected tokens '%#v', but got '%#v'", tc.expectedTokens, tokens)
			}
		})
	}
}

type JsonmanParseTestCase struct {
	path                 string
	expectedNodes        []nodeDataAccessor
	expectedErrorMessage string
}

func TestParse(t *testing.T) {
	testCases := []JsonmanParseTestCase{
		{
			path:                 "books",
			expectedNodes:        nil,
			expectedErrorMessage: "JSONPath should start with '$.'",
		},
		{
			path:                 "$.books.",
			expectedNodes:        nil,
			expectedErrorMessage: "JSONPath should not end with '.'",
		},
		{
			path:                 "$.books. ",
			expectedNodes:        nil,
			expectedErrorMessage: "Couldn't parse JSONPath substring 1: ' '",
		},
		{
			path: "$.books[0]",
			expectedNodes: []nodeDataAccessor{
				arrayIndexedNode{
					node:    node{name: "books"},
					indices: []int{0},
				},
			},
			expectedErrorMessage: "",
		},
		{
			path: "$.library.books[*]",
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
			path: "$.library.books[0]",
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
			path: "$.library.books[0,1,2]",
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
			path: "$.library.books[1:]",
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
			path: "$.library.books[1:2]",
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
			path: "$.library.books[:2]",
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
			path: "$.library.books[?(@.price < 10)]",
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
			path: "$.library.books[?(@.isbn)]",
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
			path: "$..books",
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
			path: "$..*",
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
		t.Run(fmt.Sprintf("jsonmanu.parse(%v)=%v, %v", tc.path, tc.expectedNodes, tc.expectedErrorMessage), func(t *testing.T) {
			nodes, err := parse(tc.path)
			if !cmp.Equal(tc.expectedNodes, nodes, cmp.AllowUnexported(node{}, arrayIndexedNode{}, arrayFilteredNode{}, arraySlicedNode{})) {
				t.Errorf("Expected nodes '%#v', but got '%#v'", tc.expectedNodes, nodes)
			}
			if (err == nil && len(tc.expectedErrorMessage) > 0) || (err != nil && err.Error() != tc.expectedErrorMessage) {
				t.Errorf("Expected error message '%#v', but got '%#v'", tc.expectedErrorMessage, err.Error())
			}
		})
	}
}

type JsonmanGetTestCase struct {
	path                 string
	data                 any
	expectedErrorMessage string
	expectedData         any
}

func TestGet(t *testing.T) {
	testCases := []JsonmanGetTestCase{
		{
			path: ".books",
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
			path: "$.books.",
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
			path: "$.books",
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
			path: "$.store.books.*",
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
			path: "$.books.*",
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
			path: "$.store.books[*]",
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
			path: "$.books[*]",
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
			path: "$..books[*]",
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
			path: "$..books.*",
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
			path: "$.books[0]",
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
			path: "$.books[0,2]",
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
			path: "$.books[1:]",
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
			path: "$.books[1:3]",
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
			path: "$.books[:3]",
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
			path: "$.books[?(@.price > 10)]",
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
			path: "$.books[?(@.price >= 10)]",
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
			path: "$.books[?(@.price < 10)]",
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
			path: "$.books[?(@.price <= 10)]",
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
			path: "$.books[?(@.price == 10)]",
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
			path: "$.books[?(@.price > 10)].author",
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
			path: "$.books[?(@.author == Nietzsche)]",
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
			path: "$.books[?(@.price)]",
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
			path: "$.books[*].author",
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
			path: "$.books.author",
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
			path: "$.store..author",
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
			path: "$..author",
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
			path: "$..author",
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
			path: "$..price",
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
			path: "$..books[0].author",
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

		t.Run(fmt.Sprintf("(%v) - jsonmanu.Get(%v, %v)=%v, %v", i, tc.data, tc.path, tc.expectedData, tc.expectedErrorMessage),
			func(t *testing.T) {
				data, err := Get(tc.data, tc.path)
				if (err == nil && len(tc.expectedErrorMessage) > 0) || (err != nil && err.Error() != tc.expectedErrorMessage) {
					t.Errorf("Expected error message '%#v', but got '%#v'", tc.expectedErrorMessage, err.Error())
				}
				if !cmp.Equal(tc.expectedData, data) {
					t.Errorf("\n(%v) - Expected:\n '%#v\nbut got\n'%#v'", i, tc.expectedData, data)
				}
			})
	}
}

type JsonmanPutTestCase struct {
	path                 string
	data                 any
	value                any
	expectedErrorMessage string
	expectedUpdatedData  any
}

func TestPut(t *testing.T) {
	testCases := []JsonmanPutTestCase{
		{
			path: ".books",
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
			path: "$.books.",
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
			path: "$.book.author",
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
			path: "$.book[*].author",
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
			path: "$.book[0].author",
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
			path: "$.book[0, 1].author",
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
			path: "$.book[:1].author",
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
			path: "$.book[1:].author",
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
			path: "$.book[1:2].author",
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
			path: "$.store.books[?(@.author == Nietzsche)].price",
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
			path: "$..books[?(@.author == Nietzsche)].price",
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
			path: "$.store..books[?(@.author == Nietzsche)].price",
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
			path: "$..author",
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
			path: "$..author",
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
			path: "$..price",
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
			path: "$.store..price",
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
		t.Run(fmt.Sprintf("(%v) - jsonmanu.Put(%v, %v, %v)=%v", i, tc.data, tc.path, tc.value, tc.expectedErrorMessage), func(t *testing.T) {
			err := Put(tc.data, tc.path, tc.value)
			if (err == nil && len(tc.expectedErrorMessage) > 0) || (err != nil && err.Error() != tc.expectedErrorMessage) {
				t.Errorf("Expected error message '%#v', but got '%#v'", tc.expectedErrorMessage, err.Error())
			}
			if !cmp.Equal(tc.expectedUpdatedData, tc.data) {
				t.Errorf("Expected '%#v', but got '%#v'", tc.expectedUpdatedData, tc.data)
			}
		})
	}
}
