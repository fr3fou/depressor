package huffman

import (
	"github.com/fr3fou/depressor/pq"
)

type Node struct {
	Left  *pq.Node[Node, int]
	Right *pq.Node[Node, int]
	Rune  rune
}

func BuildTreeFromText(text string) pq.Node[Node, int] {
	queue := pq.NewPriorityQueue[Node, int]()
	frequency := map[rune]int{}
	for _, r := range text {
		frequency[r]++
	}

	for k, v := range frequency {
		queue.Push(pq.Node[Node, int]{
			Data: Node{
				Rune: k,
			},
			Value: v,
		})
	}

	for !queue.Empty() {
		left, leftOk := queue.Pop()
		right, rightOk := queue.Pop()
		if !leftOk || !rightOk {
			return left
		}
		queue.Push(pq.Node[Node, int]{
			Data: Node{
				Left:  &left,
				Right: &right,
			},
			Value: left.Value + right.Value,
		})
	}

	return pq.Node[Node, int]{}
}

func BuildHuffmanDictionary(huffman *pq.Node[Node, int]) map[rune]uint32 {
	dictionary := map[rune]uint32{}
	buildHuffmanDictionary(huffman, dictionary, 0, 0)
	return dictionary
}

func buildHuffmanDictionary(huffman *pq.Node[Node, int], dictionary map[rune]uint32, depth uint32, state uint32) {
	if huffman.Data.Rune != 0 {
		dictionary[huffman.Data.Rune] = state | (1 << depth) // Mark end of state/prefix code, will be removed before serializing
		return
	}

	if huffman.Data.Left != nil {
		// Set 0 for left
		buildHuffmanDictionary(huffman.Data.Left, dictionary, depth+1, state)
	}

	if huffman.Data.Right != nil {
		// Set 1 for right
		buildHuffmanDictionary(huffman.Data.Right, dictionary, depth+1, state|(1<<depth))
	}
}
