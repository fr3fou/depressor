package huffman

import (
	"github.com/fr3fou/depressor/pq"
)

type Node struct {
	Left  *pq.Node[Node, int]
	Right *pq.Node[Node, int]
	Rune  rune
}

func BuildTree(text string) pq.Node[Node, int] {
	queue := pq.NewPriorityQueue[Node, int]()
	frequency := map[rune]int{}
	for _, r := range text {
		frequency[r]++
	}

	for k, v := range frequency {
		node := Node{
			Rune: k,
		}
		queue.Push(node, v)
	}

	for !queue.Empty() {
		left, leftOk := queue.Pop()
		right, rightOk := queue.Pop()
		if !leftOk || !rightOk {
			return left
		}
		node := Node{
			Left:  &left,
			Right: &right,
		}
		queue.Push(node, left.Value+right.Value)
	}

	return pq.Node[Node, int]{}
}

func BuildDictionary(node pq.Node[Node, int]) map[rune]uint32 {
	dictionary := map[rune]uint32{}
	buildHuffmanDictionary(node, dictionary, 0, 0)
	return dictionary
}

func buildHuffmanDictionary(node pq.Node[Node, int], dictionary map[rune]uint32, depth uint32, state uint32) {
	if node.Data.Rune != 0 {
		dictionary[node.Data.Rune] = state | (1 << depth) // Mark end of state/prefix code, will be removed before serializing
		return
	}

	if node.Data.Left != nil {
		// Set 0 for left
		buildHuffmanDictionary(*node.Data.Left, dictionary, depth+1, state)
	}

	if node.Data.Right != nil {
		// Set 1 for right
		buildHuffmanDictionary(*node.Data.Right, dictionary, depth+1, state|(1<<depth))
	}
}
