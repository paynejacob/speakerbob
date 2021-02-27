package sound

type Node struct {
	char     byte
	children map[byte]*Node
	soundIds []string
}

func NewNode(char byte) *Node {
	return &Node{char: char, children: make(map[byte]*Node), soundIds: make([]string, 0)}
}

func (r *Node) AddSound(sound Sound) {
	r.soundIds = append(r.soundIds, sound.Id)
}
