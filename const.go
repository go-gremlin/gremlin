package gremlin

const (
	singleQuote rune = '\''
	doubleQuote rune = '"'
	backslash   rune = '\\'
	pctSymbol   rune = '%'
	// Gremlin client allows the following:
	//     - all alphanumeric characters
	//     - all whitespace charaters
	//     - the following punctuation: \, ;, ., :, /, -, ?, !, *, (, ), &, _, =, ,, #, ?, !, ', "
	ARG_REGEX = "^[\\d\\w\\s\\\\;\\.\\:\\/\\-\\?\\!\\*\\(\\)\\&\\_\\=\\,\\#\\?\\!\\'\\>\\<\"]+$"
)

var ESCAPE_CHARS_GREMLIN = CharSliceToMap([]rune{
	singleQuote,
	backslash,
	pctSymbol,
	doubleQuote,
})
