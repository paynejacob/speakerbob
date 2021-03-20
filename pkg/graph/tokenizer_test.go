package graph_test

import (
	"github.com/paynejacob/speakerbob/pkg/graph"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestTokenize(t *testing.T) {
	// empty string
	{
		in := ""
		out := graph.Tokenize(in)

		assert.Equal(t, [][]byte{}, out)
	}

	// 1 token
	{
		in := "abc"
		out := graph.Tokenize(in)

		assert.Equal(t, [][]byte{{97, 98, 99}}, out)
	}

	// 2 token
	{
		in := "abc def"
		out := graph.Tokenize(in)

		assert.Equal(t, [][]byte{{97, 98, 99}, {100, 101, 102}}, out)
	}

	// special characters
	{
		in := "abc, def"
		out := graph.Tokenize(in)

		assert.Equal(t, [][]byte{{97, 98, 99}, {100, 101, 102}}, out)
	}

	// capital letters
	{
		in := "aBc"
		out := graph.Tokenize(in)

		assert.Equal(t, [][]byte{{97, 98, 99}}, out)

		in = "abC DeF"
		out = graph.Tokenize(in)

		assert.Equal(t, [][]byte{{97, 98, 99}, {100, 101, 102}}, out)
	}
}
