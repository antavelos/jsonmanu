package jsonmanu

import (
	"fmt"

	"github.com/google/go-cmp/cmp"

	// "github.com/google/go-cmp/cmp/cmpopts"
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
		{JSON_PATH_SIMPLE_NODE_PATTERN, "books", MatchDictionary{"node": "books"}},
		{JSON_PATH_SIMPLE_NODE_PATTERN, "", MatchDictionary{"node": ""}},
		{JSON_PATH_ARRAY_NODE_PATTERN, "books[*]", MatchDictionary{"node": "books"}},
		{JSON_PATH_INDEXED_ARRAY_NODE_PATTERN, "books[1,2]", MatchDictionary{"node": "books", "indices": "1,2"}},
		{JSON_PATH_SLICED_ARRAY_NODE_PATTERN, "books[-1:]", MatchDictionary{"node": "books", "start": "-1", "end": ""}},
		{JSON_PATH_SLICED_ARRAY_NODE_PATTERN, "books[3:7]", MatchDictionary{"node": "books", "start": "3", "end": "7"}},
		{JSON_PATH_SLICED_ARRAY_NODE_PATTERN, "books[:7]", MatchDictionary{"node": "books", "start": "", "end": "7"}},
		{JSON_PATH_FILTERED_ARRAY_NODE_PATTERN, "books[?(@.price)]", MatchDictionary{"node": "books", "key": "price", "op": "", "value": ""}},
		{JSON_PATH_FILTERED_ARRAY_NODE_PATTERN, "books[?(@.price < 10)]", MatchDictionary{"node": "books", "key": "price", "op": "<", "value": "10"}},
		{JSON_PATH_FILTERED_ARRAY_NODE_PATTERN, "books[?(@.price > 10)]", MatchDictionary{"node": "books", "key": "price", "op": ">", "value": "10"}},
		{JSON_PATH_FILTERED_ARRAY_NODE_PATTERN, "books[?(@.price >= 10)]", MatchDictionary{"node": "books", "key": "price", "op": ">=", "value": "10"}},
		{JSON_PATH_FILTERED_ARRAY_NODE_PATTERN, "books[?(@.price <= 10)]", MatchDictionary{"node": "books", "key": "price", "op": "<=", "value": "10"}},
		{JSON_PATH_FILTERED_ARRAY_NODE_PATTERN, "books[?(@.price == 10)]", MatchDictionary{"node": "books", "key": "price", "op": "==", "value": "10"}},
		{JSON_PATH_FILTERED_ARRAY_NODE_PATTERN, "books[?(@.price != 10)]", MatchDictionary{"node": "books", "key": "price", "op": "!=", "value": "10"}},
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

func TestNnodeFromJsonPathSubNode(t *testing.T) {
	cases := []NodeFromTokenTestCase{
		{"*", node{name: "*"}},
		{"books", node{name: "books"}},
		{"books[*]", arrayIndexedNode{node: node{name: "books"}}},
		{"books[1]", arrayIndexedNode{node: node{name: "books"}, indices: []int{1}}},
		{"books[1,2]", arrayIndexedNode{node: node{name: "books"}, indices: []int{1, 2}}},
		{"books[-1:]", arraySlicedNode{node: node{name: "books"}, start: -1}},
		{"books[1:3]", arraySlicedNode{node: node{name: "books"}, start: 1, end: 3}},
		{"books[:3]", arraySlicedNode{node: node{name: "books"}, end: 3}},
		{"books[?(@.price < 10)]", arrayFilteredNode{node: node{name: "books"}, key: "price", op: "<", value: "10"}},
		{"books[?(@.price <= 10)]", arrayFilteredNode{node: node{name: "books"}, key: "price", op: "<=", value: "10"}},
		{"books[?(@.price >= 10)]", arrayFilteredNode{node: node{name: "books"}, key: "price", op: ">=", value: "10"}},
		{"books[?(@.price > 10)]", arrayFilteredNode{node: node{name: "books"}, key: "price", op: ">", value: "10"}},
		{"books[?(@.price == 10)]", arrayFilteredNode{node: node{name: "books"}, key: "price", op: "==", value: "10"}},
		{"books[?(@.price != 10)]", arrayFilteredNode{node: node{name: "books"}, key: "price", op: "!=", value: "10"}},
		{"books[?(@.price)]", arrayFilteredNode{node: node{name: "books"}, key: "price", op: "", value: ""}},
	}

	for _, tc := range cases {
		t.Run(fmt.Sprintf("nodeFromJsonPathSubNode(%v)=%v", tc.str, tc.expectedNode), func(t *testing.T) {
			n := nodeFromJsonPathSubNode(tc.str)
			if !cmp.Equal(tc.expectedNode, n, cmp.AllowUnexported(node{}, arrayIndexedNode{}, arrayFilteredNode{}, arraySlicedNode{})) {
				t.Errorf("Expected '%#v', but got '%#v'", tc.expectedNode, n)
			}
		})
	}
}

type NodeDataAccessorGetTestCase struct {
	manager       nodeDataAccessor
	sourceData    any
	expectedData  any
	expectedError error
}

type NodeDataAccessorPutTestCase struct {
	manager             nodeDataAccessor
	sourceData          any
	value               any
	expectedError       error
	expectedUpdatedData any
}

func TestNodeGet(t *testing.T) {
	testCases := []NodeDataAccessorGetTestCase{
		{
			manager:       node{"books"},
			sourceData:    map[string]any{"books": []any{1, 2, 3}},
			expectedData:  []any{1, 2, 3},
			expectedError: nil,
		},
		{
			manager:       node{"books"},
			sourceData:    map[string]any{"book": []any{1, 2, 3}},
			expectedData:  nil,
			expectedError: SourceValidationError("key 'books' not found"),
		},
		{
			manager:       node{"books"},
			sourceData:    []any{1, 2, 3},
			expectedData:  nil,
			expectedError: SourceValidationError("Source data is not a map: []interface {}{1, 2, 3}"),
		},
	}

	for _, tc := range testCases {
		t.Run(fmt.Sprintf("node.get(%v)=%v", tc.sourceData, tc.expectedData), func(t *testing.T) {
			data, err := tc.manager.get(tc.sourceData)
			if err != tc.expectedError {
				t.Errorf("Expected error '%#v', but got '%#v'", tc.expectedError, err)
			}
			if !cmp.Equal(tc.expectedData, data) {
				t.Errorf("Expected '%#v', but got '%#v'", tc.expectedData, data)
			}
		})
	}
}

func TestNodePut(t *testing.T) {
	testCases := []NodeDataAccessorPutTestCase{
		{
			manager:             node{"price"},
			sourceData:          map[string]any{"price": 10},
			value:               20,
			expectedError:       nil,
			expectedUpdatedData: map[string]any{"price": 20},
		},
		{
			manager:             node{"author"},
			sourceData:          map[string]any{"author": "Stirner"},
			value:               "Nietzsche",
			expectedError:       nil,
			expectedUpdatedData: map[string]any{"author": "Nietzsche"},
		},
		{
			manager:             node{"numbers"},
			sourceData:          map[string]any{"numbers": []any{1, 2, 3}},
			value:               []any{2.3, 4.5, 6.7},
			expectedError:       nil,
			expectedUpdatedData: map[string]any{"numbers": []any{2.3, 4.5, 6.7}},
		},
		{
			manager:             node{"numbers"},
			sourceData:          []any{1, 2, 3},
			value:               []any{2.3, 4.5, 6.7},
			expectedError:       SourceValidationError(fmt.Sprintf("Source data is not a map: %#v", []any{1, 2, 3})),
			expectedUpdatedData: []any{1, 2, 3},
		},
		{
			manager:             node{"numbers"},
			sourceData:          []any{1, 2, 3},
			value:               100,
			expectedError:       SourceValidationError(fmt.Sprintf("Source data is not a map: %#v", []any{1, 2, 3})),
			expectedUpdatedData: []any{1, 2, 3},
		},
		{
			manager:             node{"numbers"},
			sourceData:          map[string]any{"book": []any{1, 2, 3}},
			value:               100,
			expectedError:       SourceValidationError(fmt.Sprintf("key 'numbers' not found")),
			expectedUpdatedData: map[string]any{"book": []any{1, 2, 3}},
		},
	}

	for _, tc := range testCases {
		t.Run(fmt.Sprintf("node.put(%v, %v)=%v", tc.sourceData, tc.value, tc.expectedUpdatedData), func(t *testing.T) {
			err := tc.manager.put(tc.sourceData, tc.value)
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
			manager: arrayIndexedNode{
				node:    node{name: "books"},
				indices: []int{0, 2},
			},
			sourceData:    []any{1, 2, 3},
			expectedData:  nil,
			expectedError: SourceValidationError(fmt.Sprintf("Source data is not a map: %#v", []any{1, 2, 3})),
		},
		{
			manager: arrayIndexedNode{
				node:    node{name: "books"},
				indices: []int{0, 2},
			},
			sourceData:    map[string]any{"books": []any{1, 2, 3}},
			expectedData:  []any{1, 3},
			expectedError: nil,
		},
		{
			manager: arrayIndexedNode{
				node:    node{name: "books"},
				indices: []int{0},
			},
			sourceData:    map[string]any{"books": []any{1, 2, 3}},
			expectedData:  []any{1},
			expectedError: nil,
		},
		{
			manager: arrayIndexedNode{
				node:    node{name: "books"},
				indices: []int{0, -1, 4},
			},
			sourceData:    map[string]any{"books": []any{1, 2, 3}},
			expectedData:  []any{1},
			expectedError: nil,
		},
		{
			manager: arrayIndexedNode{
				node:    node{name: "books"},
				indices: []int{0, 4},
			},
			sourceData:    map[string]any{"books": []any{1, 2, 3}},
			expectedData:  []any{1},
			expectedError: nil,
		},
		{
			manager: arrayIndexedNode{
				node:    node{name: "books"},
				indices: []int{0, 4},
			},
			sourceData:    map[string]any{"books": 1},
			expectedData:  nil,
			expectedError: SourceValidationError(fmt.Sprintf("Value of key 'books' is not an array: 1")),
		},
		{
			manager: arrayIndexedNode{
				node:    node{name: "books"},
				indices: []int{},
			},
			sourceData:    map[string]any{"books": []any{1, 2, 3}},
			expectedData:  []any{1, 2, 3},
			expectedError: nil,
		},
	}

	for _, tc := range testCases {
		t.Run(fmt.Sprintf("arrayIndexedNode.get(%v)=%v", tc.sourceData, tc.expectedData), func(t *testing.T) {
			data, err := tc.manager.get(tc.sourceData)

			if err != tc.expectedError {
				t.Errorf("Expected error '%#v', but got '%#v'", tc.expectedError, err)
			}

			if !cmp.Equal(tc.expectedData, data) {
				t.Errorf("Expected '%#v', but got '%#v'", tc.expectedData, data)
			}
		})
	}
}

func TestArrayIndexedNodePut(t *testing.T) {
	testCases := []NodeDataAccessorPutTestCase{
		{
			manager: arrayIndexedNode{
				node:    node{name: "books"},
				indices: []int{0, 2},
			},
			sourceData:          map[string]any{"books": []any{1, 2, 3}},
			value:               100,
			expectedError:       nil,
			expectedUpdatedData: map[string]any{"books": []any{100, 2, 100}},
		},
		{
			manager: arrayIndexedNode{
				node:    node{name: "books"},
				indices: []int{0, 2},
			},
			sourceData:          map[string]any{"books": []any{1, 2, 3}},
			value:               "hundred",
			expectedError:       nil,
			expectedUpdatedData: map[string]any{"books": []any{"hundred", 2, "hundred"}},
		},
		{
			manager: arrayIndexedNode{
				node:    node{name: "books"},
				indices: []int{0, 2},
			},
			sourceData:          []any{1, 2, 3},
			value:               100,
			expectedError:       SourceValidationError(fmt.Sprintf("Source data is not a map: %#v", []any{1, 2, 3})),
			expectedUpdatedData: []any{1, 2, 3},
		},
		{
			manager: arrayIndexedNode{
				node:    node{name: "books"},
				indices: []int{0, 2},
			},
			sourceData:          map[string]any{"book": []any{1, 2, 3}},
			value:               100,
			expectedError:       SourceValidationError(fmt.Sprintf("key 'books' not found")),
			expectedUpdatedData: map[string]any{"book": []any{1, 2, 3}},
		},
		{
			manager: arrayIndexedNode{
				node:    node{name: "books"},
				indices: []int{0, 2},
			},
			sourceData:          map[string]any{"books": 1},
			value:               100,
			expectedError:       SourceValidationError(fmt.Sprintf("Value of key 'books' is not an array: 1")),
			expectedUpdatedData: map[string]any{"books": 1},
		},
	}

	for _, tc := range testCases {
		t.Run(fmt.Sprintf("arrayIndexedNode.put(%v)=%v", tc.sourceData, tc.expectedError), func(t *testing.T) {
			err := tc.manager.put(tc.sourceData, tc.value)
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
			manager: arraySlicedNode{
				node:  node{name: "books"},
				start: 0,
				end:   1,
			},
			sourceData:    map[string]any{"books": []any{1, 2, 3}},
			expectedData:  []any{1},
			expectedError: nil,
		},
		{
			manager: arraySlicedNode{
				node:  node{name: "books"},
				start: 1,
			},
			sourceData:    map[string]any{"books": []any{1, 2, 3}},
			expectedData:  []any{2, 3},
			expectedError: nil,
		},
		{
			manager: arraySlicedNode{
				node: node{name: "books"},
				end:  2,
			},
			sourceData:    map[string]any{"books": []any{1, 2, 3}},
			expectedData:  []any{1, 2},
			expectedError: nil,
		},
		{
			manager: arraySlicedNode{
				node: node{name: "books"},
			},
			sourceData:    map[string]any{"books": []any{1, 2, 3}},
			expectedData:  map[string]any{"books": []any{1, 2, 3}},
			expectedError: nil,
		},
		{
			manager: arraySlicedNode{
				node: node{name: "book"},
			},
			sourceData:    map[string]any{"books": []any{1, 2, 3}},
			expectedData:  nil,
			expectedError: SourceValidationError(fmt.Sprintf("key 'book' not found")),
		},
		{
			manager: arraySlicedNode{
				node:  node{name: "books"},
				start: 0,
				end:   1,
			},
			sourceData:    map[string]any{"books": 1},
			expectedData:  nil,
			expectedError: SourceValidationError(fmt.Sprintf("Value of key 'books' is not an array: 1")),
		},
	}

	for _, tc := range testCases {
		t.Run(fmt.Sprintf("arraySlicedNode.get(%v)=%v", tc.sourceData, tc.expectedData), func(t *testing.T) {
			data, err := tc.manager.get(tc.sourceData)

			if err != tc.expectedError {
				t.Errorf("Expected error '%#v', but got '%#v'", tc.expectedError, err)
			}

			if !cmp.Equal(tc.expectedData, data) {
				t.Errorf("Expected '%#v', but got '%#v'", tc.expectedData, data)
			}
		})
	}
}

func TestArraySlicedNodePut(t *testing.T) {
	testCases := []NodeDataAccessorPutTestCase{
		{
			manager: arraySlicedNode{
				node:  node{name: "books"},
				start: 0,
				end:   1,
			},
			sourceData:          map[string]any{"books": []any{1, 2, 3}},
			value:               100,
			expectedError:       nil,
			expectedUpdatedData: map[string]any{"books": []any{100, 2, 3}},
		},
		{
			manager: arraySlicedNode{
				node: node{name: "books"},
				end:  1,
			},
			sourceData:          map[string]any{"books": []any{1, 2, 3}},
			value:               100,
			expectedError:       nil,
			expectedUpdatedData: map[string]any{"books": []any{100, 2, 3}},
		},
		{
			manager: arraySlicedNode{
				node:  node{name: "books"},
				start: 1,
				end:   2,
			},
			sourceData:          map[string]any{"books": []any{1, 2, 3}},
			value:               100,
			expectedError:       nil,
			expectedUpdatedData: map[string]any{"books": []any{1, 100, 3}},
		},
		{
			manager: arraySlicedNode{
				node:  node{name: "books"},
				start: 1,
			},
			sourceData:          map[string]any{"books": []any{1, 2, 3}},
			value:               100,
			expectedError:       nil,
			expectedUpdatedData: map[string]any{"books": []any{1, 100, 100}},
		},
		{
			manager: arraySlicedNode{
				node: node{name: "books"},
			},
			sourceData:          map[string]any{"books": []any{1, 2, 3}},
			value:               100,
			expectedError:       nil,
			expectedUpdatedData: map[string]any{"books": []any{1, 2, 3}},
		},
		{
			manager: arraySlicedNode{
				node:  node{name: "books"},
				start: 1,
				end:   2,
			},
			sourceData:          []any{1, 2, 3},
			value:               100,
			expectedError:       SourceValidationError(fmt.Sprintf("Source data is not a map: %#v", []any{1, 2, 3})),
			expectedUpdatedData: []any{1, 2, 3},
		},
		{
			manager: arraySlicedNode{
				node:  node{name: "books"},
				start: 1,
				end:   2,
			},
			sourceData:          map[string]any{"book": []any{1, 2, 3}},
			value:               100,
			expectedError:       SourceValidationError(fmt.Sprintf("key 'books' not found")),
			expectedUpdatedData: map[string]any{"book": []any{1, 2, 3}},
		},
		{
			manager: arraySlicedNode{
				node:  node{name: "books"},
				start: 1,
				end:   2,
			},
			sourceData:          map[string]any{"books": 1},
			value:               100,
			expectedError:       SourceValidationError(fmt.Sprintf("Value of key 'books' is not an array: 1")),
			expectedUpdatedData: map[string]any{"books": 1},
		},
	}

	for _, tc := range testCases {
		t.Run(fmt.Sprintf("arraySlicedNode.put(%v)=%v", tc.sourceData, tc.expectedError), func(t *testing.T) {
			err := tc.manager.put(tc.sourceData, tc.value)
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
			manager: arrayFilteredNode{
				node:  node{name: "books"},
				key:   "price",
				op:    "<",
				value: 10,
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
			expectedError: nil,
		},
		{
			manager: arrayFilteredNode{
				node:  node{name: "books"},
				key:   "price",
				op:    ">",
				value: 5,
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
			expectedError: nil,
		},
		{
			manager: arrayFilteredNode{
				node:  node{name: "books"},
				key:   "price",
				op:    "==",
				value: 5,
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
			expectedError: nil,
		},
		{
			manager: arrayFilteredNode{
				node:  node{name: "books"},
				key:   "price",
				op:    "!=",
				value: 5,
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
			expectedError: nil,
		},
		{
			manager: arrayFilteredNode{
				node:  node{name: "books"},
				key:   "price",
				op:    ">",
				value: "5",
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
			expectedError: nil,
		},
		{
			manager: arrayFilteredNode{
				node:  node{name: "books"},
				key:   "price",
				op:    ">",
				value: "20.1",
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
			expectedError: nil,
		},
		{
			manager: arrayFilteredNode{
				node:  node{name: "books"},
				key:   "price",
				op:    ">",
				value: 20.1,
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
			expectedError: nil,
		},
		{
			manager: arrayFilteredNode{
				node:  node{name: "books"},
				key:   "price",
				op:    ">",
				value: "20.1",
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
			expectedError: nil,
		},
		{
			manager: arrayFilteredNode{
				node: node{name: "books"},
				key:  "price",
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
			expectedError: nil,
		},
		{
			manager: arrayFilteredNode{
				node:  node{name: "books"},
				key:   "author",
				op:    "==",
				value: "Nietzsche",
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
			expectedError: nil,
		},
		{
			manager: arrayFilteredNode{
				node:  node{name: "books"},
				key:   "author",
				op:    "!=",
				value: "Nietzsche",
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
			expectedError: nil,
		},
		{
			manager: arrayFilteredNode{
				node:  node{name: "books"},
				key:   "author",
				op:    "!=",
				value: "Nietzsche",
			},
			sourceData:    []any{1, 2, 3},
			expectedError: SourceValidationError(fmt.Sprintf("Source data is not a map: %#v", []any{1, 2, 3})),
		},
		{
			manager: arrayFilteredNode{
				node:  node{name: "books"},
				key:   "author",
				op:    "!=",
				value: "Nietzsche",
			},
			sourceData:    map[string]any{"book": []any{1, 2, 3}},
			expectedError: SourceValidationError(fmt.Sprintf("key 'books' not found")),
		},
		{
			manager: arrayFilteredNode{
				node:  node{name: "books"},
				key:   "author",
				op:    "!=",
				value: "Nietzsche",
			},
			sourceData:    map[string]any{"books": 1},
			expectedError: SourceValidationError(fmt.Sprintf("Value of key 'books' is not an array: 1")),
		},
	}

	for i, tc := range testCases {
		t.Run(fmt.Sprintf("%d: arrayFilteredNode.get(%v)=%v", i, tc.sourceData, tc.expectedData), func(t *testing.T) {
			data, err := tc.manager.get(tc.sourceData)

			if err != tc.expectedError {
				t.Errorf("Expected error '%#v', but got '%#v'", tc.expectedError, err)
			}

			if !cmp.Equal(tc.expectedData, data) {
				t.Errorf("Expected '%#v', but got '%#v'", tc.expectedData, data)
			}
		})
	}
}

func TestArrayFilteredNodePut(t *testing.T) {
	testCases := []NodeDataAccessorPutTestCase{
		{
			manager: arrayFilteredNode{
				node:  node{name: "books"},
				key:   "price",
				op:    "<",
				value: 10,
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
			manager: arrayFilteredNode{
				node:  node{name: "books"},
				key:   "price",
				op:    ">",
				value: 5,
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
			manager: arrayFilteredNode{
				node:  node{name: "books"},
				key:   "price",
				op:    "==",
				value: 5,
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
			manager: arrayFilteredNode{
				node:  node{name: "books"},
				key:   "price",
				op:    "==",
				value: "5",
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
			manager: arrayFilteredNode{
				node:  node{name: "books"},
				key:   "price",
				op:    "!=",
				value: 5,
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
			manager: arrayFilteredNode{
				node:  node{name: "books"},
				key:   "price",
				op:    ">",
				value: "20.1",
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
			manager: arrayFilteredNode{
				node:  node{name: "books"},
				key:   "price",
				op:    ">",
				value: 20.1,
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
			manager: arrayFilteredNode{
				node: node{name: "books"},
				key:  "price",
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
			manager: arrayFilteredNode{
				node:  node{name: "books"},
				key:   "author",
				op:    "==",
				value: "Nietzsche",
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
			manager: arrayFilteredNode{
				node:  node{name: "books"},
				key:   "author",
				op:    "!=",
				value: "Nietzsche",
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
		{
			manager: arrayFilteredNode{
				node:  node{name: "books"},
				key:   "author",
				op:    "!=",
				value: "Nietzsche",
			},
			sourceData:          []any{1, 2, 3},
			value:               100,
			expectedError:       SourceValidationError(fmt.Sprintf("Source data is not a map: %#v", []any{1, 2, 3})),
			expectedUpdatedData: []any{1, 2, 3},
		},
		{
			manager: arrayFilteredNode{
				node:  node{name: "books"},
				key:   "author",
				op:    "!=",
				value: "Nietzsche",
			},
			sourceData:          map[string]any{"book": []any{1, 2, 3}},
			value:               100,
			expectedError:       SourceValidationError(fmt.Sprintf("key 'books' not found")),
			expectedUpdatedData: map[string]any{"book": []any{1, 2, 3}},
		},
		{
			manager: arrayFilteredNode{
				node:  node{name: "books"},
				key:   "author",
				op:    "!=",
				value: "Nietzsche",
			},
			sourceData:          map[string]any{"books": 1},
			value:               100,
			expectedError:       SourceValidationError(fmt.Sprintf("Value of key 'books' is not an array: 1")),
			expectedUpdatedData: map[string]any{"books": 1},
		},
	}

	for _, tc := range testCases {
		t.Run(fmt.Sprintf("arrayFilteredNode.put(%v)=%v", tc.sourceData, tc.expectedError), func(t *testing.T) {
			err := tc.manager.put(tc.sourceData, tc.value)
			if err != tc.expectedError {
				t.Errorf("Expected error '%#v', but got '%#v'", tc.expectedError, err)
			}
			if !cmp.Equal(tc.expectedUpdatedData, tc.sourceData) {
				t.Errorf("Expected '%#v', but got '%#v'", tc.expectedUpdatedData, tc.sourceData)
			}
		})
	}
}
