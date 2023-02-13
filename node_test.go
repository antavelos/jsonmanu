package jsonmanu

import (
	"fmt"

	"testing"

	"github.com/google/go-cmp/cmp"
)

type MatchDictionaryTestCase struct {
	pattern      string
	str          string
	expectedDict matchDictionary
}

func TestGetMatchDictionary(t *testing.T) {
	cases := []MatchDictionaryTestCase{
		{`\w+`, "alex", matchDictionary{}},
		{`(?P<name>\w+)`, "", matchDictionary{}},
		{`(?P<name>\w+)`, "-", matchDictionary{}},
		{`(?P<name>\w+)`, "alex", matchDictionary{"name": "alex"}},
		{`^(?P<name>\w+) (?P<age>\d+)$`, "alex 40", matchDictionary{"name": "alex", "age": "40"}},
		{jsonPathSimpleNodePattern, "books", matchDictionary{"node": "books"}},
		{jsonPathSimpleNodePattern, "", matchDictionary{"node": ""}},
		{jsonPathArrayNodePattern, "books[*]", matchDictionary{"node": "books"}},
		{jsonPathIndexedArrayNodePattern, "books[1,2]", matchDictionary{"node": "books", "indices": "1,2"}},
		{jsonPathSlicedArrayNodePattern, "books[-1:]", matchDictionary{"node": "books", "start": "-1", "end": ""}},
		{jsonPathSlicedArrayNodePattern, "books[3:7]", matchDictionary{"node": "books", "start": "3", "end": "7"}},
		{jsonPathSlicedArrayNodePattern, "books[:7]", matchDictionary{"node": "books", "start": "", "end": "7"}},
		{jsonPathFilteredArrayNodePattern, "books[?(@.price)]", matchDictionary{"node": "books", "key": "price", "op": "", "value": ""}},
		{jsonPathFilteredArrayNodePattern, "books[?(@.price < 10)]", matchDictionary{"node": "books", "key": "price", "op": "<", "value": "10"}},
		{jsonPathFilteredArrayNodePattern, "books[?(@.price > 10)]", matchDictionary{"node": "books", "key": "price", "op": ">", "value": "10"}},
		{jsonPathFilteredArrayNodePattern, "books[?(@.price >= 10)]", matchDictionary{"node": "books", "key": "price", "op": ">=", "value": "10"}},
		{jsonPathFilteredArrayNodePattern, "books[?(@.price <= 10)]", matchDictionary{"node": "books", "key": "price", "op": "<=", "value": "10"}},
		{jsonPathFilteredArrayNodePattern, "books[?(@.price == 10)]", matchDictionary{"node": "books", "key": "price", "op": "==", "value": "10"}},
		{jsonPathFilteredArrayNodePattern, "books[?(@.price != 10)]", matchDictionary{"node": "books", "key": "price", "op": "!=", "value": "10"}},
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
	manager              nodeDataAccessor
	sourceData           map[string]any
	expectedData         any
	expectedErrorMessage string
}

type NodeDataAccessorPutTestCase struct {
	manager              nodeDataAccessor
	sourceData           map[string]any
	value                any
	expectedErrorMessage string
	expectedUpdatedData  any
}

func TestNodeGet(t *testing.T) {
	testCases := []NodeDataAccessorGetTestCase{
		{
			manager:              node{"books"},
			sourceData:           map[string]any{"books": []any{1, 2, 3}},
			expectedData:         []any{1, 2, 3},
			expectedErrorMessage: "",
		},
		{
			manager:              node{"books"},
			sourceData:           map[string]any{"book": []any{1, 2, 3}},
			expectedData:         nil,
			expectedErrorMessage: "DataValidationError: Source key not found: 'books'",
		},
	}

	for i, tc := range testCases {
		t.Run(fmt.Sprintf("[%v]: node.get(%v)=%v", i, tc.sourceData, tc.expectedData), func(t *testing.T) {
			data, err := tc.manager.get(tc.sourceData)
			if (err == nil && len(tc.expectedErrorMessage) > 0) || (err != nil && err.Error() != tc.expectedErrorMessage) {
				t.Errorf("Expected error message '%#v', but got '%#v'", tc.expectedErrorMessage, err.Error())
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
			manager:              node{"price"},
			sourceData:           map[string]any{"price": 10},
			value:                20,
			expectedErrorMessage: "",
			expectedUpdatedData:  map[string]any{"price": 20},
		},
		{
			manager:              node{"author"},
			sourceData:           map[string]any{"author": "Stirner"},
			value:                "Nietzsche",
			expectedErrorMessage: "",
			expectedUpdatedData:  map[string]any{"author": "Nietzsche"},
		},
		{
			manager:              node{"numbers"},
			sourceData:           map[string]any{"numbers": []any{1, 2, 3}},
			value:                []any{2.3, 4.5, 6.7},
			expectedErrorMessage: "",
			expectedUpdatedData:  map[string]any{"numbers": []any{2.3, 4.5, 6.7}},
		},
	}

	for _, tc := range testCases {
		t.Run(fmt.Sprintf("node.put(%v, %v)=%v", tc.sourceData, tc.value, tc.expectedUpdatedData), func(t *testing.T) {
			err := tc.manager.put(tc.sourceData, tc.value)
			if (err == nil && len(tc.expectedErrorMessage) > 0) || (err != nil && err.Error() != tc.expectedErrorMessage) {
				t.Errorf("Expected error '%#v', but got '%#v'", tc.expectedErrorMessage, err.Error())
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
			sourceData:           map[string]any{"books": []any{1, 2, 3}},
			expectedData:         []any{1, 3},
			expectedErrorMessage: "",
		},
		{
			manager: arrayIndexedNode{
				node:    node{name: "books"},
				indices: []int{0},
			},
			sourceData:           map[string]any{"books": []any{1, 2, 3}},
			expectedData:         []any{1},
			expectedErrorMessage: "",
		},
		{
			manager: arrayIndexedNode{
				node:    node{name: "books"},
				indices: []int{0, -1, 4},
			},
			sourceData:           map[string]any{"books": []any{1, 2, 3}},
			expectedData:         []any{1},
			expectedErrorMessage: "",
		},
		{
			manager: arrayIndexedNode{
				node:    node{name: "books"},
				indices: []int{0, 4},
			},
			sourceData:           map[string]any{"books": []any{1, 2, 3}},
			expectedData:         []any{1},
			expectedErrorMessage: "",
		},
		{
			manager: arrayIndexedNode{
				node:    node{name: "books"},
				indices: []int{0, 4},
			},
			sourceData:           map[string]any{"books": 1},
			expectedData:         nil,
			expectedErrorMessage: "DataValidationError: Value of key 'books' is not an array: 1",
		},
		{
			manager: arrayIndexedNode{
				node:    node{name: "books"},
				indices: []int{},
			},
			sourceData:           map[string]any{"books": []any{1, 2, 3}},
			expectedData:         []any{1, 2, 3},
			expectedErrorMessage: "",
		},
	}

	for _, tc := range testCases {
		t.Run(fmt.Sprintf("arrayIndexedNode.get(%v)=%v", tc.sourceData, tc.expectedData), func(t *testing.T) {
			data, err := tc.manager.get(tc.sourceData)

			if (err == nil && len(tc.expectedErrorMessage) > 0) || (err != nil && err.Error() != tc.expectedErrorMessage) {
				t.Errorf("Expected error '%#v', but got '%#v'", tc.expectedErrorMessage, err.Error())
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
			sourceData:           map[string]any{"books": []any{1, 2, 3}},
			value:                100,
			expectedErrorMessage: "",
			expectedUpdatedData:  map[string]any{"books": []any{100, 2, 100}},
		},
		{
			manager: arrayIndexedNode{
				node:    node{name: "books"},
				indices: []int{0, 2},
			},
			sourceData:           map[string]any{"books": []any{1, 2, 3}},
			value:                "hundred",
			expectedErrorMessage: "",
			expectedUpdatedData:  map[string]any{"books": []any{"hundred", 2, "hundred"}},
		},
		{
			manager: arrayIndexedNode{
				node:    node{name: "books"},
				indices: []int{0, 2},
			},
			sourceData:           map[string]any{"book": []any{1, 2, 3}},
			value:                100,
			expectedErrorMessage: "DataValidationError: Source key not found: 'books'",
			expectedUpdatedData:  map[string]any{"book": []any{1, 2, 3}},
		},
		{
			manager: arrayIndexedNode{
				node:    node{name: "books"},
				indices: []int{0, 2},
			},
			sourceData:           map[string]any{"books": 1},
			value:                100,
			expectedErrorMessage: "DataValidationError: Value of key 'books' is not an array: 1",
			expectedUpdatedData:  map[string]any{"books": 1},
		},
	}

	for _, tc := range testCases {
		t.Run(fmt.Sprintf("arrayIndexedNode.put(%v)=%v", tc.sourceData, tc.expectedErrorMessage), func(t *testing.T) {
			err := tc.manager.put(tc.sourceData, tc.value)
			if (err == nil && len(tc.expectedErrorMessage) > 0) || (err != nil && err.Error() != tc.expectedErrorMessage) {
				t.Errorf("Expected error '%#v', but got '%#v'", tc.expectedErrorMessage, err.Error())
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
			sourceData:           map[string]any{"books": []any{1, 2, 3}},
			expectedData:         []any{1},
			expectedErrorMessage: "",
		},
		{
			manager: arraySlicedNode{
				node:  node{name: "books"},
				start: 1,
			},
			sourceData:           map[string]any{"books": []any{1, 2, 3}},
			expectedData:         []any{2, 3},
			expectedErrorMessage: "",
		},
		{
			manager: arraySlicedNode{
				node: node{name: "books"},
				end:  2,
			},
			sourceData:           map[string]any{"books": []any{1, 2, 3}},
			expectedData:         []any{1, 2},
			expectedErrorMessage: "",
		},
		{
			manager: arraySlicedNode{
				node: node{name: "books"},
			},
			sourceData:           map[string]any{"books": []any{1, 2, 3}},
			expectedData:         map[string]any{"books": []any{1, 2, 3}},
			expectedErrorMessage: "",
		},
		{
			manager: arraySlicedNode{
				node: node{name: "book"},
			},
			sourceData:           map[string]any{"books": []any{1, 2, 3}},
			expectedData:         nil,
			expectedErrorMessage: "DataValidationError: Source key not found: 'book'",
		},
		{
			manager: arraySlicedNode{
				node:  node{name: "books"},
				start: 0,
				end:   1,
			},
			sourceData:           map[string]any{"books": 1},
			expectedData:         nil,
			expectedErrorMessage: "DataValidationError: Value of key 'books' is not an array: 1",
		},
	}

	for _, tc := range testCases {
		t.Run(fmt.Sprintf("arraySlicedNode.get(%v)=%v", tc.sourceData, tc.expectedData), func(t *testing.T) {
			data, err := tc.manager.get(tc.sourceData)

			if (err == nil && len(tc.expectedErrorMessage) > 0) || (err != nil && err.Error() != tc.expectedErrorMessage) {
				t.Errorf("Expected error '%#v', but got '%#v'", tc.expectedErrorMessage, err.Error())
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
			sourceData:           map[string]any{"books": []any{1, 2, 3}},
			value:                100,
			expectedErrorMessage: "",
			expectedUpdatedData:  map[string]any{"books": []any{100, 2, 3}},
		},
		{
			manager: arraySlicedNode{
				node: node{name: "books"},
				end:  1,
			},
			sourceData:           map[string]any{"books": []any{1, 2, 3}},
			value:                100,
			expectedErrorMessage: "",
			expectedUpdatedData:  map[string]any{"books": []any{100, 2, 3}},
		},
		{
			manager: arraySlicedNode{
				node:  node{name: "books"},
				start: 1,
				end:   2,
			},
			sourceData:           map[string]any{"books": []any{1, 2, 3}},
			value:                100,
			expectedErrorMessage: "",
			expectedUpdatedData:  map[string]any{"books": []any{1, 100, 3}},
		},
		{
			manager: arraySlicedNode{
				node:  node{name: "books"},
				start: 1,
			},
			sourceData:           map[string]any{"books": []any{1, 2, 3}},
			value:                100,
			expectedErrorMessage: "",
			expectedUpdatedData:  map[string]any{"books": []any{1, 100, 100}},
		},
		{
			manager: arraySlicedNode{
				node: node{name: "books"},
			},
			sourceData:           map[string]any{"books": []any{1, 2, 3}},
			value:                100,
			expectedErrorMessage: "",
			expectedUpdatedData:  map[string]any{"books": []any{1, 2, 3}},
		},
		{
			manager: arraySlicedNode{
				node:  node{name: "books"},
				start: 1,
				end:   2,
			},
			sourceData:           map[string]any{"book": []any{1, 2, 3}},
			value:                100,
			expectedErrorMessage: "DataValidationError: Source key not found: 'books'",
			expectedUpdatedData:  map[string]any{"book": []any{1, 2, 3}},
		},
		{
			manager: arraySlicedNode{
				node:  node{name: "books"},
				start: 1,
				end:   2,
			},
			sourceData:           map[string]any{"books": 1},
			value:                100,
			expectedErrorMessage: "DataValidationError: Value of key 'books' is not an array: 1",
			expectedUpdatedData:  map[string]any{"books": 1},
		},
	}

	for _, tc := range testCases {
		t.Run(fmt.Sprintf("arraySlicedNode.put(%v)=%v", tc.sourceData, tc.expectedErrorMessage), func(t *testing.T) {
			err := tc.manager.put(tc.sourceData, tc.value)
			if (err == nil && len(tc.expectedErrorMessage) > 0) || (err != nil && err.Error() != tc.expectedErrorMessage) {
				t.Errorf("Expected error '%#v', but got '%#v'", tc.expectedErrorMessage, err.Error())
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
			expectedErrorMessage: "",
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
			expectedErrorMessage: "",
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
			expectedErrorMessage: "",
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
			expectedErrorMessage: "",
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
			expectedErrorMessage: "",
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
			expectedErrorMessage: "",
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
			expectedErrorMessage: "",
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
			expectedErrorMessage: "",
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
			expectedErrorMessage: "",
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
			expectedErrorMessage: "",
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
			expectedErrorMessage: "",
		},
		{
			manager: arrayFilteredNode{
				node:  node{name: "books"},
				key:   "author",
				op:    "!=",
				value: "Nietzsche",
			},
			sourceData:           map[string]any{"book": []any{1, 2, 3}},
			expectedErrorMessage: "DataValidationError: Source key not found: 'books'",
		},
		{
			manager: arrayFilteredNode{
				node:  node{name: "books"},
				key:   "author",
				op:    "!=",
				value: "Nietzsche",
			},
			sourceData:           map[string]any{"books": 1},
			expectedErrorMessage: "DataValidationError: Value of key 'books' is not an array: 1",
		},
	}

	for i, tc := range testCases {
		t.Run(fmt.Sprintf("%d: arrayFilteredNode.get(%v)=%v", i, tc.sourceData, tc.expectedData), func(t *testing.T) {
			data, err := tc.manager.get(tc.sourceData)

			if (err == nil && len(tc.expectedErrorMessage) > 0) || (err != nil && err.Error() != tc.expectedErrorMessage) {
				t.Errorf("Expected error '%#v', but got '%#v'", tc.expectedErrorMessage, err.Error())
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
			value:                100,
			expectedErrorMessage: "",
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
			value:                100,
			expectedErrorMessage: "",
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
			value:                100,
			expectedErrorMessage: "",
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
			value:                100,
			expectedErrorMessage: "",
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
			value:                100,
			expectedErrorMessage: "",
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
			value:                100,
			expectedErrorMessage: "",
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
			value:                100,
			expectedErrorMessage: "",
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
			value:                100,
			expectedErrorMessage: "",
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
			value:                "Fr. Nietzsche",
			expectedErrorMessage: "",
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
			value:                "Not Nietzsche",
			expectedErrorMessage: "",
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
			sourceData:           map[string]any{"book": []any{1, 2, 3}},
			value:                100,
			expectedErrorMessage: "DataValidationError: Source key not found: 'books'",
			expectedUpdatedData:  map[string]any{"book": []any{1, 2, 3}},
		},
		{
			manager: arrayFilteredNode{
				node:  node{name: "books"},
				key:   "author",
				op:    "!=",
				value: "Nietzsche",
			},
			sourceData:           map[string]any{"books": 1},
			value:                100,
			expectedErrorMessage: "DataValidationError: Value of key 'books' is not an array: 1",
			expectedUpdatedData:  map[string]any{"books": 1},
		},
	}

	for _, tc := range testCases {
		t.Run(fmt.Sprintf("arrayFilteredNode.put(%v)=%v", tc.sourceData, tc.expectedErrorMessage), func(t *testing.T) {
			err := tc.manager.put(tc.sourceData, tc.value)
			if (err == nil && len(tc.expectedErrorMessage) > 0) || (err != nil && err.Error() != tc.expectedErrorMessage) {
				t.Errorf("Expected error '%#v', but got '%#v'", tc.expectedErrorMessage, err.Error())
			}
			if !cmp.Equal(tc.expectedUpdatedData, tc.sourceData) {
				t.Errorf("Expected '%#v', but got '%#v'", tc.expectedUpdatedData, tc.sourceData)
			}
		})
	}
}
