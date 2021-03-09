package sound

import (
	"strings"
	"sync"
)

type Index struct {
	m sync.RWMutex

	root *Node
}

func NewIndex() *Index {
	return &Index{root: NewNode(0)}
}

func (i *Index) IndexSound(sound Sound) {
	var char byte
	var found bool

	i.m.Lock()
	defer i.m.Unlock()

	tokens := strings.Split(strings.ToLower(sound.Name), " ")
	for _, token := range tokens {
		root := i.root
		for _, c := range token {
			char = byte(c)

			// new node
			if _, found = root.children[char]; !found {
				root.children[char] = NewNode(char)
			}

			// write sound to this node
			root.children[char].AddSound(sound)

			// move down tree
			root = root.children[char]
		}
	}
}

func (i *Index) Search(query string) []string {
	var char byte
	root := i.root

	i.m.RLock()
	defer i.m.RUnlock()

	// walk tree, return final nodes results
	var found bool
	var nextRoot *Node
	for _, c := range strings.ToLower(query) {
		char = byte(c)
		if nextRoot, found = root.children[char]; found {
			root = nextRoot
		} else {
			return make([]string, 0)
		}
	}

	return root.soundIds
}

func (i *Index) Clear() {
	i.root = NewNode(0)
}
