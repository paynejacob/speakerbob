package graph

type Graph struct {
	root *Node
}

func NewGraph() *Graph {
	return &Graph{
		root: newNode(0),
	}
}

func (g *Graph) Write(tokens [][]byte, value []byte) {
	for i := 0; i < len(tokens); i++ {
		g.writeToken(tokens[i], value)
	}
}

func (g *Graph) Search(q []byte) [][]byte {
	var root *Node

	root = g.root
	for _, char := range q {
		// check if this sequence value exists
		if root.HasChild(char) {
			root = root.children[char]
		} else {
			return [][]byte{}
		}
	}

	return root.Values()
}

func (g *Graph) Delete(value []byte) {
	var stack nodeStack
	var root *Node

	stack.Push(g.root)
	for {

		// We have traversed all relevant nodes exit
		if stack.Empty() {
			break
		}

		// get the next node
		root = stack.Pop()

		// if this node has the index it's children may also so we need to check them
		if root.HasValue(value) {
			for k := range root.children {
				stack.Push(root.children[k])
			}
		} else {
			// if this node does not have the value we are done
			continue
		}

		root.RemoveValue(value)

		if root.Empty() {
			if root.parent != nil {
				root.parent.RemoveChild(root.char)
			}
		}
	}
}

func (g *Graph) writeToken(token []byte, value []byte) {
	var root *Node
	var char byte

	root = g.root
	root.AddValue(value)
	for i := range token {
		char = token[i]
		// ensure this Node exists
		if !root.HasChild(char) {
			root.AddChild(newNode(char))
		}

		// move down the tree
		root = root.children[char]

		// write value to Node
		root.AddValue(value)
	}
}
