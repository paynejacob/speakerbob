package graph

type nodeStack struct {
	nodes []*Node
}

func (s *nodeStack) Empty() bool {
	return len(s.nodes) == 0
}

func (s *nodeStack) Pop() *Node {
	r := s.nodes[0]

	s.nodes = s.nodes[1:]

	return r
}

func (s *nodeStack) Push(nodes ...*Node) {
	s.nodes = append(s.nodes, nodes...)
}
