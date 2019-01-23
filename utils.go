package gremlin

import (
	"bytes"
	"strconv"
)

func CharSliceToMap(chars []rune) map[rune]bool {
	charMap := make(map[rune]bool)
	for _, char := range chars {
		charMap[char] = true
	}
	return charMap
}

func InterfaceToString(i interface{}) string {
	s, _ := i.(string)
	return s
}

func CoalesceStrings(s ...string) string {
	for _, v := range s {
		if v != "" {
			return v
		}
	}
	return ""
}

func StringToBool(s string) bool {
	b, _ := strconv.ParseBool(s)
	return b
}

func InterfaceToBool(i interface{}) bool {
	switch i.(interface{}).(type) {
	case string:
		return StringToBool(InterfaceToString(i))
	case int, float64, float32:
		if i == 1 {
			return true
		}
		return false
	case bool:
		return i.(bool)
	default:
		return false
	}
}

func EscapeArgs(args []interface{}, escapeFn func(string) string) []interface{} {
	for idx := range args {
		switch args[idx].(type) {
		case string:
			args[idx] = escapeFn(args[idx].(string))
		}
	}
	return args
}

func EscapeGremlin(value string) string {
	return escapeCharacters(value, ESCAPE_CHARS_GREMLIN)
}

func escapeCharacters(value string, escapeChars map[rune]bool) string {
	var buffer bytes.Buffer

	for _, char := range value {
		if escapeChars[char] {
			buffer.WriteRune(backslash)
		}
		buffer.WriteRune(char)
	}

	return buffer.String()
}
