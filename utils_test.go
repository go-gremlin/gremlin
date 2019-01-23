package gremlin

import (
	"encoding/json"
	"fmt"
	"testing"
)

func TestCharSliceToMap(t *testing.T) {
	given := []rune{
		singleQuote,
		backslash,
		pctSymbol,
		doubleQuote,
	}
	expectedMap := make(map[rune]bool)
	expectedMap[singleQuote] = true
	expectedMap[backslash] = true
	expectedMap[pctSymbol] = true
	expectedMap[doubleQuote] = true
	expected, _ := json.Marshal(expectedMap)

	resultMap := CharSliceToMap(given)
	result, _ := json.Marshal(resultMap)
	if string(result) != string(expected) {
		t.Error("given", given, "expected", expected, "result", result)
	}
}

func TestInterfaceToString(t *testing.T) {
	tests := [][]interface{}{
		{"", ""},
		{"test", "test"},
		{1, ""},
		{true, ""},
	}

	for _, test := range tests {
		given := test[0]
		expected := test[1].(string)
		result := InterfaceToString(given)
		if result != expected {
			t.Error("given", given, "expected", expected, "result", result)
		}
	}
}

func TestCoalesceStrings(t *testing.T) {
	tests := [][][]string{
		{{"first"}, {"first"}},
		{{"", "first"}, {"first"}},
		{{"", "first", "", "second"}, {"first"}},
	}
	for _, test := range tests {
		given := test[0]
		expected := test[1][0]
		result := CoalesceStrings(given...)
		if result != expected {
			t.Error("given", given, "expected", expected, "result", result)
		}
	}
}

func TestEscapeCharacters(t *testing.T) {
	tests := [][]string{
		{"this is a test", "this is a test"},
		{`this is a %`, `this is a \%`},
		{"", ""},
		{`' \ % "`, `\' \\ \% \"`},
	}
	for _, test := range tests {
		given := test[0]
		expected := test[1]
		result := escapeCharacters(given, ESCAPE_CHARS_GREMLIN)
		if result != expected {
			t.Error("given", given, "expected", expected, "result", result)
		}
	}
}

func TestEscapeGremlin(t *testing.T) {
	tests := [][]string{
		{"this is a test", "this is a test"},
		{`this is a %`, `this is a \%`},
		{"", ""},
		{`' \ % "`, `\' \\ \% \"`},
	}
	for _, test := range tests {
		given := test[0]
		expected := test[1]
		result := EscapeGremlin(given)
		if result != expected {
			t.Error("given", given, "expected", expected, "result", result)
		}
	}
}

func TestEscapeArgs(t *testing.T) {
	tests := [][][]interface{}{
		{{"blah"}, {"blah"}},
		{{"blah", 1}, {"blah", 1}},
		{{"blah", 1, `escape '`}, {"blah", 1, `escape \'`}},
	}
	for _, test := range tests {
		given := test[0]
		expected := fmt.Sprintf("%v", test[1])
		result := fmt.Sprintf("%v", EscapeArgs(given, EscapeGremlin))
		if result != expected {
			t.Error("given", given, "expected", expected, "result", result)
		}
	}
}
