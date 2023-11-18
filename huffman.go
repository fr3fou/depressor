package main

import (
	"encoding/binary"
	"io"
	"unicode/utf8"
)

type HuffmanNode struct {
	Left  *PriorityQueueNode[HuffmanNode, int]
	Right *PriorityQueueNode[HuffmanNode, int]
	Rune  rune
}

func BuildHuffmanTree(text string) PriorityQueueNode[HuffmanNode, int] {
	queue := NewPriorityQueue[HuffmanNode, int]()
	frequency := map[rune]int{}
	for _, r := range text {
		frequency[r]++
	}

	for k, v := range frequency {
		queue.Push(PriorityQueueNode[HuffmanNode, int]{
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
		queue.Push(PriorityQueueNode[HuffmanNode, int]{
			Data: HuffmanNode{
				Left:  &left,
				Right: &right,
			},
			Value: left.Value + right.Value,
		})
	}

	return PriorityQueueNode[HuffmanNode, int]{}
}

func EncodeHuffmanTree(w io.Writer, huffman *PriorityQueueNode[HuffmanNode, int]) error {
	if huffman.Data.Left != nil {
		if err := EncodeHuffmanTree(w, huffman.Data.Left); err != nil {
			return err
		}
	}

	if huffman.Data.Rune != 0 {
		p := make([]byte, utf8.RuneLen(huffman.Data.Rune))
		utf8.EncodeRune(p, huffman.Data.Rune)
		if err := binary.Write(w, binary.BigEndian, p); err != nil {
			return err
		}
	} else {
		if err := binary.Write(w, binary.BigEndian, byte(1)); err != nil {
			return err
		}
	}

	if huffman.Data.Right != nil {
		if err := EncodeHuffmanTree(w, huffman.Data.Right); err != nil {
			return err
		}
	}

	return nil
}

func BuildHuffmanDictionary(huffman *PriorityQueueNode[HuffmanNode, int]) map[rune]uint32 {
	dictionary := map[rune]uint32{}
	buildHuffmanDictionary(huffman, dictionary, 0, 0)
	return dictionary
}

func buildHuffmanDictionary(huffman *PriorityQueueNode[HuffmanNode, int], dictionary map[rune]uint32, depth uint32, state uint32) {
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
