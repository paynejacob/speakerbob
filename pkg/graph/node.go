package graph

type Node struct {
	char byte

	parent   *Node
	children map[byte]*Node
	values   [][]byte
}

func newNode(char byte) *Node {
	return &Node{char: char, children: make(map[byte]*Node, 0), values: make([][]byte, 0)}
}

func (n *Node) AddValue(value []byte) {
	if n.HasValue(value) {
		return
	}

	n.values = append(n.values, value)
}

func (n *Node) HasValue(value []byte) bool {
	var found bool

	for i := 0; i < len(n.values); i++ {
		if len(n.values[i]) == len(value) {
			found = true
			for ii := 0; ii < len(value); ii++ {
				if n.values[i][ii] != value[ii] {
					found = false
					break
				}
			}

			if found {
				return true
			}
		}
	}

	return false
}

func (n *Node) RemoveValue(value []byte) {
	var found bool
	var pos int

	for i := 0; i < len(n.values); i++ {
		if len(n.values[i]) == len(value) {
			found = true
			for ii := 0; ii < len(value); ii++ {
				if n.values[i][ii] != value[ii] {
					found = false
					break
				}
			}

			if found {
				pos = i
				break
			}
		}
	}

	n.values[len(n.values)-1], n.values[pos] = n.values[pos], n.values[len(n.values)-1]
	n.values = n.values[:len(n.values)-1]
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
