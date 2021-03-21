package graph

import (
	"bytes"
	"regexp"
	"strings"
)

var tokenRegexp = regexp.MustCompile(`(\w+){1}`)

func Tokenize(in string) [][]byte {
	if len(in) == 0 {
		return [][]byte{}
	}

	tokens := make([][]byte, strings.Count(in, " ")+1)

	var cursor int
	var next bool

	for _, c := range in {
		if c == 32 { // whitespace
			next = true
			continue
		}

		if (c > 64 && c < 91) || (c > 96 && c < 123) {
			if next {
				cursor += 1
				next = false
			}
			tokens[cursor] = append(tokens[cursor], bytes.ToLower([]byte{byte(c)})[0])
		}
	}

	return tokens
}
