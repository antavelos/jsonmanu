package jsonmanu

import (
	"fmt"
	cmp "github.com/google/go-cmp/cmp"
	"testing"
)

// JSONPath				    Result
// $.store.book[*].author	the authors of all books in the store
// $..author				all authors
// $.store.*				all things in store, which are some books and a red bicycle.
// $.store..price			the price of everything in the store.
// $..book[2]				the third book
// $..book[(@.length-1)]
// $..book[-1:]			    the last book in order.
// $..book[0,1]
// $..book[:2]		    	the first two books
// $..book[?(@.isbn)]		filter all books with isbn number
// $..book[?(@.price<10)]	filter all books cheapier than 10
// $..*				    	all Elements in XML document. All members of JSON structure.

type MatchDictionaryTestCase struct {
	pattern      string
	str          string
	expectedDict MatchDictionary
}

func TestGetMatchDictionary(t *testing.T) {
	cases := []MatchDictionaryTestCase{
		{`\w+`, "alex", MatchDictionary{}},
		{`(?P<name>\w+)`, "", MatchDictionary{}},
		{`(?P<name>\w+)`, "-", MatchDictionary{}},
		{`(?P<name>\w+)`, "alex", MatchDictionary{"name": "alex"}},
		{`^(?P<name>\w+) (?P<age>\d+)$`, "alex 40", MatchDictionary{"name": "alex", "age": "40"}},
		{SIMPLE_NODE_TOKEN_PATTERN, "books", MatchDictionary{"node": "books"}},
		{SIMPLE_NODE_TOKEN_PATTERN, "", MatchDictionary{"node": ""}},
		{ARRAY_TOKEN_PATTERN, "books[*]", MatchDictionary{"node": "books"}},
		{INDEXED_ARRAY_TOKEN_PATTERN, "books[1,2]", MatchDictionary{"node": "books", "indices": "1,2"}},
		{SLICED_ARRAY_TOKEN_PATTERN, "books[-1:]", MatchDictionary{"node": "books", "start": "-1", "end": ""}},
		{SLICED_ARRAY_TOKEN_PATTERN, "books[3:7]", MatchDictionary{"node": "books", "start": "3", "end": "7"}},
		{SLICED_ARRAY_TOKEN_PATTERN, "books[:7]", MatchDictionary{"node": "books", "start": "", "end": "7"}},
		{FILTERED_ARRAY_TOKEN_PATTERN, "books[?(@.price)]", MatchDictionary{"node": "books", "key": "price", "op": "", "value": ""}},
		{FILTERED_ARRAY_TOKEN_PATTERN, "books[?(@.price<10)]", MatchDictionary{"node": "books", "key": "price", "op": "<", "value": "10"}},
	}

	for _, tc := range cases {
		t.Run(fmt.Sprintf("GetMatchDictionary(%v, %v)=%v", tc.pattern, tc.str, tc.expectedDict), func(t *testing.T) {
			dict := getMatchDictionary(tc.pattern, tc.str)
			if !cmp.Equal(tc.expectedDict, dict) {
				t.Errorf("Expected '%#v', but got '%#v'", tc.expectedDict, dict)
			}
		})
	}
}

type NodeFromTokenTestCase struct {
	str          string
	expectedNode any
}

func TestNodeFromToken(t *testing.T) {
	cases := []NodeFromTokenTestCase{
		{"*", Node{Name: "*"}},
		{"books", Node{Name: "books"}},
		{"books[*]", ArrayIndexedNode{ArrayNode: ArrayNode{Node: Node{Name: "books"}}}},
		{"books[1]", ArrayIndexedNode{ArrayNode: ArrayNode{Node: Node{Name: "books"}}, Indices: []int{1}}},
		{"books[1,2]", ArrayIndexedNode{ArrayNode: ArrayNode{Node: Node{Name: "books"}}, Indices: []int{1, 2}}},
		{"books[-1:]", ArraySlicedNode{ArrayNode: ArrayNode{Node: Node{Name: "books"}}, Start: -1}},
		{"books[1:3]", ArraySlicedNode{ArrayNode: ArrayNode{Node: Node{Name: "books"}}, Start: 1, End: 3}},
		{"books[:3]", ArraySlicedNode{ArrayNode: ArrayNode{Node: Node{Name: "books"}}, End: 3}},
		{"books[?(@.price<10)]", ArrayFilteredNode{ArrayNode: ArrayNode{Node: Node{Name: "books"}}, Key: "price", Op: "<", Value: "10"}},
		{"books[?(@.price<=10)]", ArrayFilteredNode{ArrayNode: ArrayNode{Node: Node{Name: "books"}}, Key: "price", Op: "<=", Value: "10"}},
		{"books[?(@.price>=10)]", ArrayFilteredNode{ArrayNode: ArrayNode{Node: Node{Name: "books"}}, Key: "price", Op: ">=", Value: "10"}},
		{"books[?(@.price>10)]", ArrayFilteredNode{ArrayNode: ArrayNode{Node: Node{Name: "books"}}, Key: "price", Op: ">", Value: "10"}},
		{"books[?(@.price=10)]", ArrayFilteredNode{ArrayNode: ArrayNode{Node: Node{Name: "books"}}, Key: "price", Op: "=", Value: "10"}},
		{"books[?(@.price!10)]", ArrayFilteredNode{ArrayNode: ArrayNode{Node: Node{Name: "books"}}, Key: "price", Op: "!", Value: "10"}},
		{"books[?(@.price)]", ArrayFilteredNode{ArrayNode: ArrayNode{Node: Node{Name: "books"}}, Key: "price", Op: "", Value: ""}},
	}

	for _, tc := range cases {
		t.Run(fmt.Sprintf("nodeFromToken(%v)=%v", tc.str, tc.expectedNode), func(t *testing.T) {
			node := nodeFromToken(tc.str)
			if !cmp.Equal(tc.expectedNode, node) {
				t.Errorf("Expected '%#v', but got '%#v'", tc.expectedNode, node)
			}
		})
	}
}

type NodeDataAccessorGetTestCase struct {
	manager      NodeDataAccessor
	sourceData   any
	expectedData any
}

type NodeDataAccessorPutTestCase struct {
	manager             NodeDataAccessor
	sourceData          any
	value               any
	expectedError       error
	expectedUpdatedData any
}

func TestNodeGet(t *testing.T) {
	testCases := []NodeDataAccessorGetTestCase{
		{
			manager:      Node{"books"},
			sourceData:   map[string]any{"books": []any{1, 2, 3}},
			expectedData: []any{1, 2, 3},
		},
		{
			manager:      Node{"books"},
			sourceData:   map[string]any{"book": []any{1, 2, 3}},
			expectedData: nil,
		},
		{
			manager:      Node{"books"},
			sourceData:   []any{1, 2, 3},
			expectedData: nil,
		},
	}

	for _, tc := range testCases {
		t.Run(fmt.Sprintf("Node.Get(%v)=%v", tc.sourceData, tc.expectedData), func(t *testing.T) {
			data := tc.manager.Get(tc.sourceData)
			if !cmp.Equal(tc.expectedData, data) {
				t.Errorf("Expected '%#v', but got '%#v'", tc.expectedData, data)
			}
		})
	}
}

func TestNodePut(t *testing.T) {
	testCases := []NodeDataAccessorPutTestCase{
		{
			manager:             Node{"price"},
			sourceData:          map[string]any{"price": 10},
			value:               20,
			expectedError:       nil,
			expectedUpdatedData: map[string]any{"price": 20},
		},
		{
			manager:             Node{"author"},
			sourceData:          map[string]any{"author": "Stirner"},
			value:               "Nietzsche",
			expectedError:       nil,
			expectedUpdatedData: map[string]any{"author": "Nietzsche"},
		},
		{
			manager:             Node{"numbers"},
			sourceData:          map[string]any{"numbers": []any{1, 2, 3}},
			value:               []any{2.3, 4.5, 6.7},
			expectedError:       nil,
			expectedUpdatedData: map[string]any{"numbers": []any{2.3, 4.5, 6.7}},
		},
		{
			manager:             Node{"numbers"},
			sourceData:          []any{1, 2, 3},
			value:               []any{2.3, 4.5, 6.7},
			expectedError:       DataValidationError(fmt.Sprintf("Source data is not a map: %#v", []any{1, 2, 3})),
			expectedUpdatedData: []any{1, 2, 3},
		},
		{
			manager:             Node{"numbers"},
			sourceData:          []any{1, 2, 3},
			value:               100,
			expectedError:       DataValidationError(fmt.Sprintf("Source data is not a map: %#v", []any{1, 2, 3})),
			expectedUpdatedData: []any{1, 2, 3},
		},
		{
			manager:             Node{"numbers"},
			sourceData:          map[string]any{"book": []any{1, 2, 3}},
			value:               100,
			expectedError:       DataValidationError(fmt.Sprintf("Key 'numbers' not found")),
			expectedUpdatedData: map[string]any{"book": []any{1, 2, 3}},
		},
	}

	for _, tc := range testCases {
		t.Run(fmt.Sprintf("Node.Put(%v, %v)=%v", tc.sourceData, tc.value, tc.expectedUpdatedData), func(t *testing.T) {
			err := tc.manager.Put(tc.sourceData, tc.value)
			if err != tc.expectedError {
				t.Errorf("Expected error '%#v', but got '%#v'", tc.expectedError, err)
			}
			if !cmp.Equal(tc.expectedUpdatedData, tc.sourceData) {
				t.Errorf("Expected '%#v', but got '%#v'", tc.expectedUpdatedData, tc.sourceData)
			}
		})
	}
}

func TestArrayIndexedNodeGet(t *testing.T) {
	testCases := []NodeDataAccessorGetTestCase{
		{
			manager: ArrayIndexedNode{
				ArrayNode: ArrayNode{Node: Node{Name: "books"}},
				Indices:   []int{0, 2},
			},
			sourceData:   map[string]any{"books": []any{1, 2, 3}},
			expectedData: []any{1, 3},
		},
		{
			manager: ArrayIndexedNode{
				ArrayNode: ArrayNode{Node: Node{Name: "books"}},
				Indices:   []int{0},
			},
			sourceData:   map[string]any{"books": []any{1, 2, 3}},
			expectedData: []any{1},
		},
		{
			manager: ArrayIndexedNode{
				ArrayNode: ArrayNode{Node: Node{Name: "books"}},
				Indices:   []int{0, 4},
			},
			sourceData:   map[string]any{"books": []any{1, 2, 3}},
			expectedData: []any{1},
		},
		{
			manager: ArrayIndexedNode{
				ArrayNode: ArrayNode{Node: Node{Name: "books"}},
				Indices:   []int{0, 4},
			},
			sourceData:   map[string]any{"books": 1},
			expectedData: nil,
		},
		{
			manager: ArrayIndexedNode{
				ArrayNode: ArrayNode{Node: Node{Name: "books"}},
				Indices:   []int{},
			},
			sourceData:   map[string]any{"books": []any{1, 2, 3}},
			expectedData: []any{1, 2, 3},
		},
	}

	for _, tc := range testCases {
		t.Run(fmt.Sprintf("ArrayIndexedNode.Get(%v)=%v", tc.sourceData, tc.expectedData), func(t *testing.T) {
			data := tc.manager.Get(tc.sourceData)
			if !cmp.Equal(tc.expectedData, data) {
				t.Errorf("Expected '%#v', but got '%#v'", tc.expectedData, data)
			}
		})
	}
}

func TestArrayIndexedNodePut(t *testing.T) {
	testCases := []NodeDataAccessorPutTestCase{
		{
			manager: ArrayIndexedNode{
				ArrayNode: ArrayNode{Node: Node{Name: "books"}},
				Indices:   []int{0, 2},
			},
			sourceData:          map[string]any{"books": []any{1, 2, 3}},
			value:               100,
			expectedError:       nil,
			expectedUpdatedData: map[string]any{"books": []any{100, 2, 100}},
		},
		{
			manager: ArrayIndexedNode{
				ArrayNode: ArrayNode{Node: Node{Name: "books"}},
				Indices:   []int{0, 2},
			},
			sourceData:          map[string]any{"books": []any{1, 2, 3}},
			value:               "hundred",
			expectedError:       nil,
			expectedUpdatedData: map[string]any{"books": []any{"hundred", 2, "hundred"}},
		},
		{
			manager: ArrayIndexedNode{
				ArrayNode: ArrayNode{Node: Node{Name: "books"}},
				Indices:   []int{0, 2},
			},
			sourceData:          []any{1, 2, 3},
			value:               100,
			expectedError:       DataValidationError(fmt.Sprintf("Source data is not a map: %#v", []any{1, 2, 3})),
			expectedUpdatedData: []any{1, 2, 3},
		},
		{
			manager: ArrayIndexedNode{
				ArrayNode: ArrayNode{Node: Node{Name: "books"}},
				Indices:   []int{0, 2},
			},
			sourceData:          map[string]any{"book": []any{1, 2, 3}},
			value:               100,
			expectedError:       DataValidationError(fmt.Sprintf("Key 'books' not found")),
			expectedUpdatedData: map[string]any{"book": []any{1, 2, 3}},
		},
		{
			manager: ArrayIndexedNode{
				ArrayNode: ArrayNode{Node: Node{Name: "books"}},
				Indices:   []int{0, 2},
			},
			sourceData:          map[string]any{"books": 1},
			value:               100,
			expectedError:       DataValidationError(fmt.Sprintf("Value of key 'books' is not an array: %#v", map[string]any{"books": 1})),
			expectedUpdatedData: map[string]any{"books": 1},
		},
	}

	for _, tc := range testCases {
		t.Run(fmt.Sprintf("ArrayIndexedNode.Put(%v)=%v", tc.sourceData, tc.expectedError), func(t *testing.T) {
			err := tc.manager.Put(tc.sourceData, tc.value)
			if err != tc.expectedError {
				t.Errorf("Expected error '%#v', but got '%#v'", tc.expectedError, err)
			}
			if !cmp.Equal(tc.expectedUpdatedData, tc.sourceData) {
				t.Errorf("Expected '%#v', but got '%#v'", tc.expectedUpdatedData, tc.sourceData)
			}
		})
	}
}

func TestArraySlicedNodeGet(t *testing.T) {
	testCases := []NodeDataAccessorGetTestCase{
		{
			manager: ArraySlicedNode{
				ArrayNode: ArrayNode{Node: Node{Name: "books"}},
				Start:     0,
				End:       1,
			},
			sourceData:   map[string]any{"books": []any{1, 2, 3}},
			expectedData: []any{1},
		},
		{
			manager: ArraySlicedNode{
				ArrayNode: ArrayNode{Node: Node{Name: "books"}},
				Start:     1,
			},
			sourceData:   map[string]any{"books": []any{1, 2, 3}},
			expectedData: []any{2, 3},
		},
		{
			manager: ArraySlicedNode{
				ArrayNode: ArrayNode{Node: Node{Name: "books"}},
				End:       2,
			},
			sourceData:   map[string]any{"books": []any{1, 2, 3}},
			expectedData: []any{1, 2},
		},
		{
			manager: ArraySlicedNode{
				ArrayNode: ArrayNode{Node: Node{Name: "books"}},
			},
			sourceData:   map[string]any{"books": []any{1, 2, 3}},
			expectedData: map[string]any{"books": []any{1, 2, 3}},
		},
		{
			manager: ArraySlicedNode{
				ArrayNode: ArrayNode{Node: Node{Name: "book"}},
			},
			sourceData:   map[string]any{"books": []any{1, 2, 3}},
			expectedData: nil,
		},
		{
			manager: ArraySlicedNode{
				ArrayNode: ArrayNode{Node: Node{Name: "books"}},
				Start:     0,
				End:       1,
			},
			sourceData:   map[string]any{"books": 1},
			expectedData: nil,
		},
	}

	for _, tc := range testCases {
		t.Run(fmt.Sprintf("ArraySlicedNode.Get(%v)=%v", tc.sourceData, tc.expectedData), func(t *testing.T) {
			data := tc.manager.Get(tc.sourceData)
			if !cmp.Equal(tc.expectedData, data) {
				t.Errorf("Expected '%#v', but got '%#v'", tc.expectedData, data)
			}
		})
	}
}

func TestArraySlicedNodePut(t *testing.T) {
	testCases := []NodeDataAccessorPutTestCase{
		{
			manager: ArraySlicedNode{
				ArrayNode: ArrayNode{Node: Node{Name: "books"}},
				Start:     0,
				End:       1,
			},
			sourceData:          map[string]any{"books": []any{1, 2, 3}},
			value:               100,
			expectedError:       nil,
			expectedUpdatedData: map[string]any{"books": []any{100, 2, 3}},
		},
		{
			manager: ArraySlicedNode{
				ArrayNode: ArrayNode{Node: Node{Name: "books"}},
				End:       1,
			},
			sourceData:          map[string]any{"books": []any{1, 2, 3}},
			value:               100,
			expectedError:       nil,
			expectedUpdatedData: map[string]any{"books": []any{100, 2, 3}},
		},
		{
			manager: ArraySlicedNode{
				ArrayNode: ArrayNode{Node: Node{Name: "books"}},
				Start:     1,
				End:       2,
			},
			sourceData:          map[string]any{"books": []any{1, 2, 3}},
			value:               100,
			expectedError:       nil,
			expectedUpdatedData: map[string]any{"books": []any{1, 100, 3}},
		},
		{
			manager: ArraySlicedNode{
				ArrayNode: ArrayNode{Node: Node{Name: "books"}},
				Start:     1,
			},
			sourceData:          map[string]any{"books": []any{1, 2, 3}},
			value:               100,
			expectedError:       nil,
			expectedUpdatedData: map[string]any{"books": []any{1, 100, 100}},
		},
		{
			manager: ArraySlicedNode{
				ArrayNode: ArrayNode{Node: Node{Name: "books"}},
			},
			sourceData:          map[string]any{"books": []any{1, 2, 3}},
			value:               100,
			expectedError:       nil,
			expectedUpdatedData: map[string]any{"books": []any{1, 2, 3}},
		},
		{
			manager: ArraySlicedNode{
				ArrayNode: ArrayNode{Node: Node{Name: "books"}},
				Start:     1,
				End:       2,
			},
			sourceData:          []any{1, 2, 3},
			value:               100,
			expectedError:       DataValidationError(fmt.Sprintf("Source data is not a map: %#v", []any{1, 2, 3})),
			expectedUpdatedData: []any{1, 2, 3},
		},
		{
			manager: ArraySlicedNode{
				ArrayNode: ArrayNode{Node: Node{Name: "books"}},
				Start:     1,
				End:       2,
			},
			sourceData:          map[string]any{"book": []any{1, 2, 3}},
			value:               100,
			expectedError:       DataValidationError(fmt.Sprintf("Key 'books' not found")),
			expectedUpdatedData: map[string]any{"book": []any{1, 2, 3}},
		},
		{
			manager: ArraySlicedNode{
				ArrayNode: ArrayNode{Node: Node{Name: "books"}},
				Start:     1,
				End:       2,
			},
			sourceData:          map[string]any{"books": 1},
			value:               100,
			expectedError:       DataValidationError(fmt.Sprintf("Value of key 'books' is not an array: %#v", map[string]any{"books": 1})),
			expectedUpdatedData: map[string]any{"books": 1},
		},
	}

	for _, tc := range testCases {
		t.Run(fmt.Sprintf("ArraySlicedNode.Put(%v)=%v", tc.sourceData, tc.expectedError), func(t *testing.T) {
			err := tc.manager.Put(tc.sourceData, tc.value)
			if err != tc.expectedError {
				t.Errorf("Expected error '%#v', but got '%#v'", tc.expectedError, err)
			}
			if !cmp.Equal(tc.expectedUpdatedData, tc.sourceData) {
				t.Errorf("Expected '%#v', but got '%#v'", tc.expectedUpdatedData, tc.sourceData)
			}
		})
	}
}

func TestArrayFilteredNodeGet(t *testing.T) {
	testCases := []NodeDataAccessorGetTestCase{
		{
			manager: ArrayFilteredNode{
				ArrayNode: ArrayNode{Node: Node{Name: "books"}},
				Key:       "price",
				Op:        "<",
				Value:     10,
			},
			sourceData: map[string]any{
				"books": []any{
					map[string]any{"price": 20, "title": "Book1"},
					map[string]any{"price": 5, "title": "Book2"},
					map[string]any{"price": 50, "title": "Book3"},
				},
			},
			expectedData: []any{
				map[string]any{"price": 5, "title": "Book2"},
			},
		},
		{
			manager: ArrayFilteredNode{
				ArrayNode: ArrayNode{Node: Node{Name: "books"}},
				Key:       "price",
				Op:        ">",
				Value:     5,
			},
			sourceData: map[string]any{
				"books": []any{
					map[string]any{"price": 20, "title": "Book1"},
					map[string]any{"price": 5, "title": "Book2"},
					map[string]any{"price": 50, "title": "Book3"},
				},
			},
			expectedData: []any{
				map[string]any{"price": 20, "title": "Book1"},
				map[string]any{"price": 50, "title": "Book3"},
			},
		},
		{
			manager: ArrayFilteredNode{
				ArrayNode: ArrayNode{Node: Node{Name: "books"}},
				Key:       "price",
				Op:        "=",
				Value:     5,
			},
			sourceData: map[string]any{
				"books": []any{
					map[string]any{"price": 20, "title": "Book1"},
					map[string]any{"price": 5, "title": "Book2"},
					map[string]any{"price": 50, "title": "Book3"},
				},
			},
			expectedData: []any{
				map[string]any{"price": 5, "title": "Book2"},
			},
		},
		{
			manager: ArrayFilteredNode{
				ArrayNode: ArrayNode{Node: Node{Name: "books"}},
				Key:       "price",
				Op:        "!",
				Value:     5,
			},
			sourceData: map[string]any{
				"books": []any{
					map[string]any{"price": 20, "title": "Book1"},
					map[string]any{"price": 5, "title": "Book2"},
					map[string]any{"price": 50, "title": "Book3"},
				},
			},
			expectedData: []any{
				map[string]any{"price": 20, "title": "Book1"},
				map[string]any{"price": 50, "title": "Book3"},
			},
		},
		{
			manager: ArrayFilteredNode{
				ArrayNode: ArrayNode{Node: Node{Name: "books"}},
				Key:       "price",
				Op:        ">",
				Value:     "5",
			},
			sourceData: map[string]any{
				"books": []any{
					map[string]any{"price": 20, "title": "Book1"},
					map[string]any{"price": 5, "title": "Book2"},
					map[string]any{"price": 50, "title": "Book3"},
				},
			},
			expectedData: []any{
				map[string]any{"price": 20, "title": "Book1"},
				map[string]any{"price": 50, "title": "Book3"},
			},
		},
		{
			manager: ArrayFilteredNode{
				ArrayNode: ArrayNode{Node: Node{Name: "books"}},
				Key:       "price",
				Op:        ">",
				Value:     "20.1",
			},
			sourceData: map[string]any{
				"books": []any{
					map[string]any{"price": 20, "title": "Book1"},
					map[string]any{"price": 5, "title": "Book2"},
					map[string]any{"price": 50, "title": "Book3"},
				},
			},
			expectedData: []any{
				map[string]any{"price": 50, "title": "Book3"},
			},
		},
		{
			manager: ArrayFilteredNode{
				ArrayNode: ArrayNode{Node: Node{Name: "books"}},
				Key:       "price",
				Op:        ">",
				Value:     20.1,
			},
			sourceData: map[string]any{
				"books": []any{
					map[string]any{"price": "20", "title": "Book1"},
					map[string]any{"price": "5", "title": "Book2"},
					map[string]any{"price": "50", "title": "Book3"},
				},
			},
			expectedData: []any{
				map[string]any{"price": "50", "title": "Book3"},
			},
		},
		{
			manager: ArrayFilteredNode{
				ArrayNode: ArrayNode{Node: Node{Name: "books"}},
				Key:       "price",
				Op:        ">",
				Value:     "20.1",
			},
			sourceData: map[string]any{
				"books": []any{
					map[string]any{"price": "20", "title": "Book1"},
					map[string]any{"price": "5", "title": "Book2"},
					map[string]any{"price": "50", "title": "Book3"},
				},
			},
			expectedData: []any{
				map[string]any{"price": "50", "title": "Book3"},
			},
		},
		{
			manager: ArrayFilteredNode{
				ArrayNode: ArrayNode{Node: Node{Name: "books"}},
				Key:       "price",
			},
			sourceData: map[string]any{
				"books": []any{
					map[string]any{"price": 20, "title": "Book1"},
					map[string]any{"price": 5, "title": "Book2"},
					map[string]any{"title": "Book3"},
				},
			},
			expectedData: []any{
				map[string]any{"price": 20, "title": "Book1"},
				map[string]any{"price": 5, "title": "Book2"},
			},
		},
		{
			manager: ArrayFilteredNode{
				ArrayNode: ArrayNode{Node: Node{Name: "books"}},
				Key:       "author",
				Op:        "=",
				Value:     "Nietzsche",
			},
			sourceData: map[string]any{
				"books": []any{
					map[string]any{"author": "Nietzsche", "title": "Book1"},
					map[string]any{"author": "Nietzsche", "title": "Book2"},
					map[string]any{"author": "Stirner", "title": "Book3"},
				},
			},
			expectedData: []any{
				map[string]any{"author": "Nietzsche", "title": "Book1"},
				map[string]any{"author": "Nietzsche", "title": "Book2"},
			},
		},
		{
			manager: ArrayFilteredNode{
				ArrayNode: ArrayNode{Node: Node{Name: "books"}},
				Key:       "author",
				Op:        "!",
				Value:     "Nietzsche",
			},
			sourceData: map[string]any{
				"books": []any{
					map[string]any{"author": "Nietzsche", "title": "Book1"},
					map[string]any{"author": "Nietzsche", "title": "Book2"},
					map[string]any{"author": "Stirner", "title": "Book3"},
				},
			},
			expectedData: []any{
				map[string]any{"author": "Stirner", "title": "Book3"},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(fmt.Sprintf("ArrayFilteredNode.Get(%v)=%v", tc.sourceData, tc.expectedData), func(t *testing.T) {
			data := tc.manager.Get(tc.sourceData)
			if !cmp.Equal(tc.expectedData, data) {
				t.Errorf("Expected '%#v', but got '%#v'", tc.expectedData, data)
			}
		})
	}
}

func TestArrayFilteredNodePut(t *testing.T) {
	testCases := []NodeDataAccessorPutTestCase{
		{
			manager: ArrayFilteredNode{
				ArrayNode: ArrayNode{Node: Node{Name: "books"}},
				Key:       "price",
				Op:        "<",
				Value:     10,
			},
			sourceData: map[string]any{
				"books": []any{
					map[string]any{"price": 20, "title": "Book1"},
					map[string]any{"price": 5, "title": "Book2"},
					map[string]any{"price": 50, "title": "Book3"},
				},
			},
			value:         100,
			expectedError: nil,
			expectedUpdatedData: map[string]any{
				"books": []any{
					map[string]any{"price": 20, "title": "Book1"},
					map[string]any{"price": 100, "title": "Book2"},
					map[string]any{"price": 50, "title": "Book3"},
				},
			},
		},
		{
			manager: ArrayFilteredNode{
				ArrayNode: ArrayNode{Node: Node{Name: "books"}},
				Key:       "price",
				Op:        ">",
				Value:     5,
			},
			sourceData: map[string]any{
				"books": []any{
					map[string]any{"price": 20, "title": "Book1"},
					map[string]any{"price": 5, "title": "Book2"},
					map[string]any{"price": 50, "title": "Book3"},
				},
			},
			value:         100,
			expectedError: nil,
			expectedUpdatedData: map[string]any{
				"books": []any{
					map[string]any{"price": 100, "title": "Book1"},
					map[string]any{"price": 5, "title": "Book2"},
					map[string]any{"price": 100, "title": "Book3"},
				},
			},
		},
		{
			manager: ArrayFilteredNode{
				ArrayNode: ArrayNode{Node: Node{Name: "books"}},
				Key:       "price",
				Op:        "=",
				Value:     5,
			},
			sourceData: map[string]any{
				"books": []any{
					map[string]any{"price": 20, "title": "Book1"},
					map[string]any{"price": 5, "title": "Book2"},
					map[string]any{"price": 50, "title": "Book3"},
				},
			},
			value:         100,
			expectedError: nil,
			expectedUpdatedData: map[string]any{
				"books": []any{
					map[string]any{"price": 20, "title": "Book1"},
					map[string]any{"price": 100, "title": "Book2"},
					map[string]any{"price": 50, "title": "Book3"},
				},
			},
		},
		{
			manager: ArrayFilteredNode{
				ArrayNode: ArrayNode{Node: Node{Name: "books"}},
				Key:       "price",
				Op:        "=",
				Value:     "5",
			},
			sourceData: map[string]any{
				"books": []any{
					map[string]any{"price": 20, "title": "Book1"},
					map[string]any{"price": 5, "title": "Book2"},
					map[string]any{"price": 50, "title": "Book3"},
				},
			},
			value:         100,
			expectedError: nil,
			expectedUpdatedData: map[string]any{
				"books": []any{
					map[string]any{"price": 20, "title": "Book1"},
					map[string]any{"price": 100, "title": "Book2"},
					map[string]any{"price": 50, "title": "Book3"},
				},
			},
		},
		{
			manager: ArrayFilteredNode{
				ArrayNode: ArrayNode{Node: Node{Name: "books"}},
				Key:       "price",
				Op:        "!",
				Value:     5,
			},
			sourceData: map[string]any{
				"books": []any{
					map[string]any{"price": 20, "title": "Book1"},
					map[string]any{"price": 5, "title": "Book2"},
					map[string]any{"price": 50, "title": "Book3"},
				},
			},
			value:         100,
			expectedError: nil,
			expectedUpdatedData: map[string]any{
				"books": []any{
					map[string]any{"price": 100, "title": "Book1"},
					map[string]any{"price": 5, "title": "Book2"},
					map[string]any{"price": 100, "title": "Book3"},
				},
			},
		},
		{
			manager: ArrayFilteredNode{
				ArrayNode: ArrayNode{Node: Node{Name: "books"}},
				Key:       "price",
				Op:        ">",
				Value:     "20.1",
			},
			sourceData: map[string]any{
				"books": []any{
					map[string]any{"price": 20, "title": "Book1"},
					map[string]any{"price": 5, "title": "Book2"},
					map[string]any{"price": 50, "title": "Book3"},
				},
			},
			value:         100,
			expectedError: nil,
			expectedUpdatedData: map[string]any{
				"books": []any{
					map[string]any{"price": 20, "title": "Book1"},
					map[string]any{"price": 5, "title": "Book2"},
					map[string]any{"price": 100, "title": "Book3"},
				},
			},
		},
		{
			manager: ArrayFilteredNode{
				ArrayNode: ArrayNode{Node: Node{Name: "books"}},
				Key:       "price",
				Op:        ">",
				Value:     20.1,
			},
			sourceData: map[string]any{
				"books": []any{
					map[string]any{"price": 20, "title": "Book1"},
					map[string]any{"price": 5, "title": "Book2"},
					map[string]any{"price": 50, "title": "Book3"},
				},
			},
			value:         100,
			expectedError: nil,
			expectedUpdatedData: map[string]any{
				"books": []any{
					map[string]any{"price": 20, "title": "Book1"},
					map[string]any{"price": 5, "title": "Book2"},
					map[string]any{"price": 100, "title": "Book3"},
				},
			},
		},
		{
			manager: ArrayFilteredNode{
				ArrayNode: ArrayNode{Node: Node{Name: "books"}},
				Key:       "price",
			},
			sourceData: map[string]any{
				"books": []any{
					map[string]any{"price": 20, "title": "Book1"},
					map[string]any{"price": 5, "title": "Book2"},
					map[string]any{"title": "Book3"},
				},
			},
			value:         100,
			expectedError: nil,
			expectedUpdatedData: map[string]any{
				"books": []any{
					map[string]any{"price": 100, "title": "Book1"},
					map[string]any{"price": 100, "title": "Book2"},
					map[string]any{"title": "Book3"},
				},
			},
		},
		{
			manager: ArrayFilteredNode{
				ArrayNode: ArrayNode{Node: Node{Name: "books"}},
				Key:       "author",
				Op:        "=",
				Value:     "Nietzsche",
			},
			sourceData: map[string]any{
				"books": []any{
					map[string]any{"author": "Nietzsche", "title": "Book1"},
					map[string]any{"author": "Nietzsche", "title": "Book2"},
					map[string]any{"author": "Stirner", "title": "Book3"},
				},
			},
			value:         "Fr. Nietzsche",
			expectedError: nil,
			expectedUpdatedData: map[string]any{
				"books": []any{
					map[string]any{"author": "Fr. Nietzsche", "title": "Book1"},
					map[string]any{"author": "Fr. Nietzsche", "title": "Book2"},
					map[string]any{"author": "Stirner", "title": "Book3"},
				},
			},
		},
		{
			manager: ArrayFilteredNode{
				ArrayNode: ArrayNode{Node: Node{Name: "books"}},
				Key:       "author",
				Op:        "!",
				Value:     "Nietzsche",
			},
			sourceData: map[string]any{
				"books": []any{
					map[string]any{"author": "Nietzsche", "title": "Book1"},
					map[string]any{"author": "Nietzsche", "title": "Book2"},
					map[string]any{"author": "Stirner", "title": "Book3"},
				},
			},
			value:         "Not Nietzsche",
			expectedError: nil,
			expectedUpdatedData: map[string]any{
				"books": []any{
					map[string]any{"author": "Nietzsche", "title": "Book1"},
					map[string]any{"author": "Nietzsche", "title": "Book2"},
					map[string]any{"author": "Not Nietzsche", "title": "Book3"},
				},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(fmt.Sprintf("ArrayFilteredNode.Put(%v)=%v", tc.sourceData, tc.expectedError), func(t *testing.T) {
			err := tc.manager.Put(tc.sourceData, tc.value)
			if err != tc.expectedError {
				t.Errorf("Expected error '%#v', but got '%#v'", tc.expectedError, err)
			}
			if !cmp.Equal(tc.expectedUpdatedData, tc.sourceData) {
				t.Errorf("Expected '%#v', but got '%#v'", tc.expectedUpdatedData, tc.sourceData)
			}
		})
	}
}
