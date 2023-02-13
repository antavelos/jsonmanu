package jsonmanu

import (
	"fmt"
	"testing"

	cmp "github.com/google/go-cmp/cmp"
)

type FlattenArrayTestCase struct {
	array             []any
	expectedFlattened []any
}

func TestFlattenArray(t *testing.T) {
	cases := []FlattenArrayTestCase{
		{
			array:             []any{1, 2, []any{3, 4, []any{5, 6}}},
			expectedFlattened: []any{1, 2, 3, 4, 5, 6},
		},
		{
			array:             []any{1, 2, []any{3, 4, []any{5, 6, []any{7, 8}, 9}}},
			expectedFlattened: []any{1, 2, 3, 4, 5, 6, 7, 8, 9},
		},
	}
	for _, tc := range cases {
		t.Run(fmt.Sprintf("flattenArray(%v)=%v", tc.array, tc.expectedFlattened), func(t *testing.T) {
			flattened := flattenArray(tc.array)
			if !cmp.Equal(tc.expectedFlattened, flattened) {
				t.Errorf("Expected '%#v', but got '%#v'", tc.expectedFlattened, flattened)
			}
		})
	}
}

type MapGetDeepFlattenedTestCase struct {
	m              map[string]any
	key            string
	expectedValues []any
}

func TestMapGetDeepFlattened(t *testing.T) {
	cases := []MapGetDeepFlattenedTestCase{
		{
			m: map[string]any{
				"A1": map[string]any{
					"B1": map[string]any{
						"C1": map[string]any{
							"D1": "val1",
						},
					},
					"B2": map[string]any{
						"C1": map[string]any{
							"D1": "val2",
						},
					},
				},
			},
			key:            "D1",
			expectedValues: []any{"val1", "val2"},
		},
		{
			m: map[string]any{
				"A1": map[string]any{
					"B1": map[string]any{
						"C1": []any{
							map[string]any{"D1": "val1"},
							map[string]any{"D1": "val2"},
						},
					},
					"B2": map[string]any{
						"C1": map[string]any{
							"D1": "val3",
						},
					},
				},
			},
			key:            "D1",
			expectedValues: []any{"val1", "val2", "val3"},
		},
	}

	for _, tc := range cases {
		t.Run(fmt.Sprintf("mapGetDeep(%v, %v)=%v", tc.m, tc.key, tc.expectedValues), func(t *testing.T) {
			values := mapGetDeepFlattened(tc.m, tc.key)
			if !compareCounters(counterFromArray(tc.expectedValues), counterFromArray(values)) {
				t.Errorf("Expected '%#v', but got '%#v'", tc.expectedValues, values)
			}
		})
	}
}
