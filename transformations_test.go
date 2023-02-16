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
			expectedErrorMessage:     "Index out of bounds.",
		},
		{
			transformer:              SplitTransformer{Delim: ",", Index: 4},
			value:                    1,
			expectedTransformedValue: nil,
			expectedErrorMessage:     "Value is not a string.",
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
			expectedTransformedValue: nil,
			expectedErrorMessage:     "Value is not a string.",
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
			expectedTransformedValue: nil,
			expectedErrorMessage:     "Value is not an array.",
		},
		{
			transformer:              JoinTransformer{Delim: ","},
			value:                    true,
			expectedTransformedValue: nil,
			expectedErrorMessage:     "Value is not an array.",
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
			transformer:              StringMatchTransformer{},
			value:                    1,
			expectedTransformedValue: nil,
			expectedErrorMessage:     "Value is not a string.",
		},
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
			transformer:              SubStrTransformer{},
			value:                    1,
			expectedTransformedValue: nil,
			expectedErrorMessage:     "Value is not a string.",
		},
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
			expectedErrorMessage:     "Value is not a string.",
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
