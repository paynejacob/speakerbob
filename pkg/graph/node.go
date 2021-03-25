package graph

type Node struct {
	char byte

	parent   *Node
	children map[byte]*Node
	values   map[string]bool
}

func newNode(char byte) *Node {
	return &Node{char: char, children: make(map[byte]*Node, 0), values: make(map[string]bool, 0)}
}

func (n *Node) AddValue(value []byte) {
	n.values[string(value)] = true
}

func (n *Node) HasValue(value []byte) bool {
	var found bool

	_, found = n.values[string(value)]

	return found
}

func (n *Node) RemoveValue(value []byte) {
	delete(n.values, string(value))
}

func (n *Node) Values() [][]byte {
	rvalues := make([][]byte, len(n.values))

	i := 0
	for k, _ := range n.values {
		rvalues[i] = []byte(k)
		i++
	}

	return rvalues
}

func (n *Node) Empty() bool {
	return len(n.values) == 0
}

func (n *Node) AddChild(child *Node) {
	child.parent = n
	n.children[child.char] = child
}

func (n *Node) HasChild(char byte) bool {
	if _, ok := n.children[char]; ok {
		return true
	}

	return false
}

func (n *Node) RemoveChild(char byte) {
	delete(n.children, char)
}
