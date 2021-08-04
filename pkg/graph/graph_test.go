package graph_test

import (
	"github.com/paynejacob/speakerbob/pkg/graph"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestNewGraph(t *testing.T) {
	g := graph.NewGraph()

	assert.NotNil(t, g)
}

func TestGraph_Write(t *testing.T) {
	var g *graph.Graph

	// Write to empty graph
	{
		g = graph.NewGraph()
		g.writeToken([]byte{4, 5, 6}, []byte{1})

		assert.Equal(t, [][]byte{{1}}, g.Search([]byte{4, 5, 6}))
		assert.Equal(t, [][]byte{{1}}, g.Search([]byte{4, 5}))
		assert.Equal(t, [][]byte{{1}}, g.Search([]byte{4}))
		assert.Equal(t, [][]byte{{1}}, g.Search([]byte{}))
	}

	// Write with no overlap
	{
		g = graph.NewGraph()
		g.writeToken([]byte{4, 5, 6}, []byte{1})
		g.writeToken([]byte{7, 8, 9}, []byte{2})

		assert.Equal(t, [][]byte{{1}}, g.Search([]byte{4, 5, 6}))
		assert.Equal(t, [][]byte{{1}}, g.Search([]byte{4, 5}))
		assert.Equal(t, [][]byte{{1}}, g.Search([]byte{4}))

		assert.Equal(t, [][]byte{{2}}, g.Search([]byte{7, 8, 9}))
		assert.Equal(t, [][]byte{{2}}, g.Search([]byte{7, 8}))
		assert.Equal(t, [][]byte{{2}}, g.Search([]byte{7}))
		assert.Equal(t, [][]byte{{1}, {2}}, g.Search([]byte{}))
	}

	// Write with overlap
	{
		g = graph.NewGraph()
		g.writeToken([]byte{4, 5, 6}, []byte{1})
		g.writeToken([]byte{4, 5, 6, 7}, []byte{2})

		assert.Equal(t, [][]byte{{2}}, g.Search([]byte{4, 5, 6, 7}))
		assert.Equal(t, [][]byte{{1}, {2}}, g.Search([]byte{4, 5, 6}))
		assert.Equal(t, [][]byte{{1}, {2}}, g.Search([]byte{4, 5}))
		assert.Equal(t, [][]byte{{1}, {2}}, g.Search([]byte{4}))
		assert.Equal(t, [][]byte{{1}, {2}}, g.Search([]byte{}))
	}

	// Write twice
	{
		g = graph.NewGraph()
		g.writeToken([]byte{4, 5, 6}, []byte{1})
		g.writeToken([]byte{4, 5, 6}, []byte{1})

		assert.Equal(t, [][]byte{{1}}, g.Search([]byte{4, 5, 6}))
		assert.Equal(t, [][]byte{{1}}, g.Search([]byte{4, 5}))
		assert.Equal(t, [][]byte{{1}}, g.Search([]byte{4}))
		assert.Equal(t, [][]byte{{1}}, g.Search([]byte{}))
	}
}

func TestGraph_Search(t *testing.T) {
	var g *graph.Graph

	// search empty graph
	{
		g = graph.NewGraph()

		assert.Equal(t, [][]byte{}, g.Search([]byte{4}))
		assert.Equal(t, [][]byte{}, g.Search([]byte{}))
	}

	// search empty query
	{
		g = graph.NewGraph()
		g.writeToken([]byte{1}, []byte{1})

		assert.Equal(t, [][]byte{{1}}, g.Search([]byte{}))
	}

	// search partial match
	{
		g = graph.NewGraph()
		g.writeToken([]byte{1, 2, 3}, []byte{1})

		assert.Equal(t, [][]byte{{1}}, g.Search([]byte{}))
		assert.Equal(t, [][]byte{{1}}, g.Search([]byte{1}))
		assert.Equal(t, [][]byte{{1}}, g.Search([]byte{1, 2}))
	}

	// search full match
	{
		g = graph.NewGraph()
		g.writeToken([]byte{1, 2, 3}, []byte{1})

		assert.Equal(t, [][]byte{{1}}, g.Search([]byte{1, 2, 3}))
	}

	// search no match
	{
		g = graph.NewGraph()
		g.writeToken([]byte{1, 2, 3}, []byte{1})

		assert.Equal(t, [][]byte{}, g.Search([]byte{1, 2, 3, 4}))
		assert.Equal(t, [][]byte{}, g.Search([]byte{2, 3}))
	}
}

func TestGraph_Delete(t *testing.T) {
	var g *graph.Graph

	// empty graph
	{
		g = graph.NewGraph()

		g.Delete([]byte{0})
	}

	// missing value
	{
		g = graph.NewGraph()
		g.writeToken([]byte{1, 2, 3}, []byte{1})

		g.Delete([]byte{2})

		assert.Equal(t, [][]byte{{1}}, g.Search([]byte{1, 2, 3}))
		assert.Equal(t, [][]byte{{1}}, g.Search([]byte{1, 2}))
		assert.Equal(t, [][]byte{{1}}, g.Search([]byte{1}))
		assert.Equal(t, [][]byte{{1}}, g.Search([]byte{}))
	}

	// match on all nodes
	{
		g = graph.NewGraph()
		g.writeToken([]byte{1, 2, 3}, []byte{1})
		g.writeToken([]byte{1, 2, 3}, []byte{2})

		g.Delete([]byte{2})

		assert.Equal(t, [][]byte{{1}}, g.Search([]byte{1, 2, 3}))
		assert.Equal(t, [][]byte{{1}}, g.Search([]byte{1, 2}))
		assert.Equal(t, [][]byte{{1}}, g.Search([]byte{1}))
		assert.Equal(t, [][]byte{{1}}, g.Search([]byte{}))
	}

	// match head nodes
	{
		g = graph.NewGraph()
		g.writeToken([]byte{1, 2, 3}, []byte{1})
		g.writeToken([]byte{1, 2}, []byte{2})

		g.Delete([]byte{2})

		assert.Equal(t, [][]byte{{1}}, g.Search([]byte{1, 2, 3}))
		assert.Equal(t, [][]byte{{1}}, g.Search([]byte{1, 2}))
		assert.Equal(t, [][]byte{{1}}, g.Search([]byte{1}))
		assert.Equal(t, [][]byte{{1}}, g.Search([]byte{}))
	}

	// match tail nodes
	{
		g = graph.NewGraph()
		g.writeToken([]byte{1, 2, 3}, []byte{1})
		g.writeToken([]byte{1, 2, 3, 4}, []byte{2})

		g.Delete([]byte{2})

		assert.Equal(t, [][]byte{}, g.Search([]byte{1, 2, 3, 4}))
		assert.Equal(t, [][]byte{{1}}, g.Search([]byte{1, 2, 3}))
		assert.Equal(t, [][]byte{{1}}, g.Search([]byte{1, 2}))
		assert.Equal(t, [][]byte{{1}}, g.Search([]byte{1}))
		assert.Equal(t, [][]byte{{1}}, g.Search([]byte{}))
	}
}
