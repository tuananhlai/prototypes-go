package bplustree

// BPlusTree is a minimal B+ tree for integer keys with interface{} values.
// It supports Insert and Get. Delete and rebalancing are intentionally omitted.
type BPlusTree struct {
	order int
	root  *node
}

type node struct {
	isLeaf   bool
	keys     []int
	children []*node
	values   []interface{}
	next     *node
}

// New creates a B+ tree with the given order (minimum 3).
func New(order int) *BPlusTree {
	if order < 3 {
		order = 3
	}
	return &BPlusTree{
		order: order,
		root:  &node{isLeaf: true},
	}
}

// Get returns the value for key if present.
func (t *BPlusTree) Get(key int) (interface{}, bool) {
	if t.root == nil {
		return nil, false
	}
	leaf := findLeaf(t.root, key)
	for i, k := range leaf.keys {
		if k == key {
			return leaf.values[i], true
		}
	}
	return nil, false
}

// Insert inserts or replaces the value for key.
func (t *BPlusTree) Insert(key int, value interface{}) {
	if t.root == nil {
		t.root = &node{isLeaf: true}
	}
	leaf := findLeaf(t.root, key)
	insertIntoLeaf(leaf, key, value)
	if len(leaf.keys) > t.maxKeys() {
		separator, right := splitLeaf(leaf, t.order)
		t.insertIntoParent(leaf, separator, right)
	}
}

func (t *BPlusTree) maxKeys() int {
	return t.order - 1
}

func findLeaf(n *node, key int) *node {
	for !n.isLeaf {
		i := 0
		for i < len(n.keys) && key >= n.keys[i] {
			i++
		}
		n = n.children[i]
	}
	return n
}

func insertIntoLeaf(leaf *node, key int, value interface{}) {
	i := 0
	for i < len(leaf.keys) && leaf.keys[i] < key {
		i++
	}
	if i < len(leaf.keys) && leaf.keys[i] == key {
		leaf.values[i] = value
		return
	}
	leaf.keys = append(leaf.keys, 0)
	leaf.values = append(leaf.values, nil)
	copy(leaf.keys[i+1:], leaf.keys[i:])
	copy(leaf.values[i+1:], leaf.values[i:])
	leaf.keys[i] = key
	leaf.values[i] = value
}

func splitLeaf(leaf *node, order int) (int, *node) {
	split := (order + 1) / 2
	right := &node{
		isLeaf: true,
		keys:   append([]int(nil), leaf.keys[split:]...),
		values: append([]interface{}(nil), leaf.values[split:]...),
		next:   leaf.next,
	}
	leaf.keys = leaf.keys[:split]
	leaf.values = leaf.values[:split]
	leaf.next = right
	return right.keys[0], right
}

func splitInternal(n *node, order int) (int, *node) {
	splitIndex := order / 2
	separator := n.keys[splitIndex]
	right := &node{
		isLeaf:   false,
		keys:     append([]int(nil), n.keys[splitIndex+1:]...),
		children: append([]*node(nil), n.children[splitIndex+1:]...),
	}
	n.keys = n.keys[:splitIndex]
	n.children = n.children[:splitIndex+1]
	return separator, right
}

func (t *BPlusTree) insertIntoParent(left *node, separator int, right *node) {
	if t.root == left {
		t.root = &node{
			isLeaf:   false,
			keys:     []int{separator},
			children: []*node{left, right},
		}
		return
	}

	parent, parentIndex := findParent(t.root, left)
	if parent == nil {
		return
	}
	parent.keys = append(parent.keys, 0)
	copy(parent.keys[parentIndex+1:], parent.keys[parentIndex:])
	parent.keys[parentIndex] = separator
	parent.children = append(parent.children, nil)
	copy(parent.children[parentIndex+2:], parent.children[parentIndex+1:])
	parent.children[parentIndex+1] = right

	if len(parent.keys) > t.maxKeys() {
		sep, newRight := splitInternal(parent, t.order)
		t.insertIntoParent(parent, sep, newRight)
	}
}

func findParent(root, child *node) (*node, int) {
	if root == nil || root.isLeaf {
		return nil, -1
	}
	for i, c := range root.children {
		if c == child {
			return root, i
		}
		if !c.isLeaf {
			if parent, index := findParent(c, child); parent != nil {
				return parent, index
			}
		}
	}
	return nil, -1
}
