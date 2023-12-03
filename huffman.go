package main

import (
	"encoding/binary"
	"errors"
	"io"
	"unicode/utf8"
)

type HuffmanNode struct {
	Left  *PriorityQueueNode[HuffmanNode, int]
	Right *PriorityQueueNode[HuffmanNode, int]
	Rune  rune
}

func BuildHuffmanTreeFromText(text string) PriorityQueueNode[HuffmanNode, int] {
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

func writeNode(w io.Writer, huffman *PriorityQueueNode[HuffmanNode, int]) error {
	if huffman.Data.Rune == 0 {
		if err := binary.Write(w, binary.BigEndian, byte(1)); err != nil {
			return err
		}
		return nil
	}

	p := make([]byte, utf8.RuneLen(huffman.Data.Rune))
	utf8.EncodeRune(p, huffman.Data.Rune)
	if err := binary.Write(w, binary.BigEndian, p); err != nil {
		return err
	}
	return nil
}

func EncodeHuffmanTree(w io.Writer, huffman *PriorityQueueNode[HuffmanNode, int]) error {
	if huffman == nil {
		return nil
	}

	if err := writeNode(w, huffman); err != nil {
		return err
	}

	if err := EncodeHuffmanTree(w, huffman.Data.Left); err != nil {
		return err
	}

	if err := EncodeHuffmanTree(w, huffman.Data.Right); err != nil {
		return err
	}

	return nil
}

type Scanner interface {
	io.ByteScanner
	io.RuneScanner
}

func DecodeHuffmanTree(scanner Scanner) (*PriorityQueueNode[HuffmanNode, int], int, error) {
	s := []*PriorityQueueNode[HuffmanNode, int]{}

	// Header
	for {
		b, err := scanner.ReadByte()
		if err != nil {
			if errors.Is(err, io.EOF) {
				break
			}
			return nil, 0, err
		}

		// End of tree
		if b == 0 {
			break
		}

		// Internal Node
		if b == 1 {
			s = append(s, &PriorityQueueNode[HuffmanNode, int]{})
			continue
		}

		if err := scanner.UnreadByte(); err != nil {
			return nil, 0, err
		}

		// Leaf Node
		r, _, err := scanner.ReadRune()
		if err != nil {
			if errors.Is(err, io.EOF) {
				break
			}
			return nil, 0, err
		}

		s = append(s, &PriorityQueueNode[HuffmanNode, int]{
			Data: HuffmanNode{
				Rune: r,
			},
		})
	}

	textLength, err := binary.ReadUvarint(scanner)
	if err != nil {
		return nil, 0, err
	}

	stack := &Stack[*PriorityQueueNode[HuffmanNode, int]]{
		Data: []*PriorityQueueNode[HuffmanNode, int]{},
	}

	for i := len(s) - 1; i >= 0; i-- {
		stack.Push(s[i])
	}

	return buildHuffmanTreeFromStack(stack), int(textLength), nil
}

func buildHuffmanTreeFromStack(stack *Stack[*PriorityQueueNode[HuffmanNode, int]]) *PriorityQueueNode[HuffmanNode, int] {
	if stack.Len() < 1 {
		return nil
	}
	node := stack.Pop()
	if node.Data.Rune != 0 {
		return node
	}
	node.Data.Left = buildHuffmanTreeFromStack(stack)
	node.Data.Right = buildHuffmanTreeFromStack(stack)
	return node
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
