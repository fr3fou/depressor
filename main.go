package main

import (
	"bufio"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"math/bits"
	"os"
	"strings"
)

func main() {
	text := "maikati"
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

	buf := make([]byte, 8)

	// Header
	binary.BigEndian.PutUint64(buf, uint64(len(text)))

	// Body
	var overflow int8
	for _, r := range text {
		code := dictionary[r]
		buf, overflow = encodePrefix(buf, overflow, code)
	}

	for _, b := range buf {
		fmt.Printf("%08b ", b)
	}
	fmt.Println()

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

	br := bufio.NewReader(file)

	// Header
	buf := make([]byte, 8)
	_, err = br.Read(buf)
	if err != nil {
		panic(err)
	}

	textLength := binary.BigEndian.Uint64(buf)

	// Body
	sb := &strings.Builder{}
	for sb.Len() < int(textLength) {
		b, err := br.ReadByte()
		if err != nil && !errors.Is(err, io.EOF) {
			fmt.Println(err)
			break
		}

		p := 8
		node := huffman
		for p >= 0 && sb.Len() < int(textLength) {
			if b&(1<<p) == 0 {
				node = *node.Data.Left
			} else {
				node = *node.Data.Right
			}

			if node.Data.Rune != 0 {
				sb.WriteRune(node.Data.Rune)
				node = huffman
			}
			p--
		}

		if err != nil {
			// end of file
			break
		}
	}
	fmt.Println(sb.String())
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
		// Set 0 for left
		BuildHuffmanDictionary(huffman.Data.Left, dictionary, depth+1, state)
	}

	if huffman.Data.Right != nil {
		// Set 1 for right
		BuildHuffmanDictionary(huffman.Data.Right, dictionary, depth+1, state|(1<<depth))
	}
}

// shoutout tsetso
func encodePrefix(buf []byte, overflow int8, code uint32) ([]byte, int8) {
	acc := buf
	l := bits.Len32(code)

	// Keep track of two "pointers":
	// `read` is the current bit we're reading from `code` -> [0..l-2]
	// `write` is the current bit we're writing into `acc` -> [7..0]
	read := 0
	write := overflow - 1

	// Helper for appending a new byte and resetting the write pointer
	newByte := func() {
		acc = append(acc, 0)
		write = 7
	}

	// If there's no space left over from a previous serialization, start by
	// appending a new byte.
	if overflow == 0 {
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
