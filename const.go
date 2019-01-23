package gremlin

import (
	"time"
)

const (
	singleQuote rune = '\''
	doubleQuote rune = '"'
	backslash   rune = '\\'
	pctSymbol   rune = '%'
	// Gremlin client allows the following:
	//     - all alphanumeric characters
	//     - all whitespace charaters
	//     - the following punctuation: \, ;, ., :, /, -, ?, !, *, (, ), &, _, =, ,, #, ?, !, "
	ARG_REGEX = "^[\\d\\w\\s\\\\;\\.\\:\\/\\-\\?\\!\\*\\(\\)\\&\\_\\=\\,\\#\\?\\!\\'\\>\\<\"]+$"

	// Gremlin stack defaults
	DEFAULT_MAX_CAP             = 10
	DEFAULT_MAX_GREMLIN_RETRIES = 2
	DEFAULT_VERBOSE_LOGGING     = false
	DEFAULT_PING_INTERVAL       = 5

	// Lock defaults
	DEFAULT_LOCK_WAIT_TIME = time.Duration(200 * time.Millisecond)
	DEFAULT_MAX_RETRIES    = 5
)

var ESCAPE_CHARS_GREMLIN = CharSliceToMap([]rune{
	singleQuote,
	backslash,
	pctSymbol,
	doubleQuote,
})
