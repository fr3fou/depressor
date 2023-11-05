package main

import (
	"encoding/binary"
	"math/bits"
	"os"
)

func main() {
	text := "aabcd"
	if len(os.Args) > 1 && os.Args[1] == "decompress" {
		decompress(text)
	} else {
		compress(text)
	}
}

func compress(text string) {
	huffman := BuildHuffmanTree(text)
	dictionary := map[rune]uint32{}
	BuildHuffmanDictionary(&huffman, dictionary, 0, 0)
	file, err := os.Create("text.maikati")
	if err != nil {
		panic(err)
	}
	defer file.Close()

	var buf []byte
	var carry int8
	for _, r := range text {
		code := dictionary[r]
		buf, carry = encodePrefix(buf, carry, code)
	}

	if err := binary.Write(file, binary.BigEndian, buf); err != nil {
		panic(err)
	}
}

func decompress(text string) {
	file, err := os.Open("text.maikati")
	if err != nil {
		panic(err)
	}

	// todo get tree _from_ file
	huffman := BuildHuffmanTree(text)
}

type HuffmanNode struct {
	Left  *Node[HuffmanNode, int]
	Right *Node[HuffmanNode, int]
	Rune  rune
}

func BuildHuffmanTree(text string) Node[HuffmanNode, int] {
	queue := NewPriorityQueue[HuffmanNode, int]()
	frequency := map[rune]int{}
	for _, r := range text {
		frequency[r]++
	}

	for k, v := range frequency {
		queue.Push(Node[HuffmanNode, int]{
			Data: HuffmanNode{
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
		queue.Push(Node[HuffmanNode, int]{
			Data: HuffmanNode{
				Left:  &left,
				Right: &right,
			},
			Value: left.Value + right.Value,
		})
	}

	return Node[HuffmanNode, int]{}
}

func BuildHuffmanDictionary(huffman *Node[HuffmanNode, int], dictionary map[rune]uint32, depth uint32, state uint32) {
	if huffman.Data.Rune != 0 {
		dictionary[huffman.Data.Rune] = state | (1 << depth) // Mark end of state/prefix code, will be removed before serializing
		return
	}

	if huffman.Data.Left != nil {
		BuildHuffmanDictionary(huffman.Data.Left, dictionary, depth+1, state)
	}

	if huffman.Data.Right != nil {
		BuildHuffmanDictionary(huffman.Data.Right, dictionary, depth+1, state|(1<<depth))
	}
}

// shoutout tsetso
func encodePrefix(buf []byte, carry int8, code uint32) ([]byte, int8) {
	acc := buf
	l := bits.Len32(code)

	// Keep track of two "pointers":
	// `read` is the current bit we're reading from `code` -> [0..l-2]
	// `write` is the current bit we're writing into `acc` -> [7..0]
	read := 0
	write := carry - 1

	// Helper for appending a new byte and resetting the write pointer
	newByte := func() {
		acc = append(acc, 0)
		write = 7
	}

	// If there's no space left over from a previous serialization, start by
	// appending a new byte.
	if carry == 0 {
		newByte()
	}

	for read < l-1 {
		// If the current bit is set in `code`, set it in `acc`:
		if code&(1<<read) != 0 {
			acc[len(acc)-1] |= 1 << write
		}

		// Move the pointers
		read++
		write--

		// If we're not done and we're out of writing space, append a new byte
		if read != l-1 && write < 0 {
			newByte()
		}
	}

	return acc, write + 1
}
