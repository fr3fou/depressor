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
	text, err := os.ReadFile("lorem.txt")
	if err != nil {
		panic(err)
	}
	if len(os.Args) > 1 && os.Args[1] == "decompress" {
		decompress()
	} else {
		compress(string(text))
	}
}

func compress(text string) {
	huffman := BuildHuffmanTree(text)
	dictionary := BuildHuffmanDictionary(&huffman)

	fmt.Println(Render(&huffman))
	file, err := os.Create("text.maikati")
	if err != nil {
		panic(err)
	}
	defer file.Close()

	writer := bufio.NewWriter(file)

	// Header
	if err := EncodeHuffmanTree(writer, &huffman); err != nil {
		panic(err)
	}

	if err := binary.Write(writer, binary.BigEndian, byte(0)); err != nil {
		panic(err)
	}

	buf := []byte{}
	buf = binary.AppendUvarint(buf, uint64(len(text)))
	_, err = writer.Write(buf)
	if err != nil {
		panic(err)
	}

	// Body
	var overflow int8
	buf = []byte{}
	for _, r := range text {
		code := dictionary[r]
		buf, overflow = encodeVarint(buf, overflow, code)
	}

	_, err = writer.Write(buf)
	if err != nil {
		panic(err)
	}

	if err := writer.Flush(); err != nil {
		panic(err)
	}
}

func decompress() {
	file, err := os.Open("text.maikati")
	if err != nil {
		panic(err)
	}

	br := bufio.NewReader(file)

	s := []*PriorityQueueNode[HuffmanNode, int]{}

	// Header
	for {
		b, err := br.ReadByte()
		if err != nil {
			if errors.Is(err, io.EOF) {
				break
			}
			panic(err)
		}

		fmt.Printf("%08b ", b)

		if b == 0 {
			break
		}

		if b == 1 {
			s = append(s, &PriorityQueueNode[HuffmanNode, int]{})
			continue
		}

		if err := br.UnreadByte(); err != nil {
			panic(err)
		}

		// Leaf Node
		r, _, err := br.ReadRune()
		if err != nil {
			if errors.Is(err, io.EOF) {
				break
			}
			panic(err)
		}

		s = append(s, &PriorityQueueNode[HuffmanNode, int]{
			Data: HuffmanNode{
				Rune: r,
			},
		})
	}
	fmt.Println()

	stack := Stack[*PriorityQueueNode[HuffmanNode, int]]{
		Data: []*PriorityQueueNode[HuffmanNode, int]{
			s[len(s)-1],
		},
	}

	for i := len(s) - 2; i >= 0; i-- {
		n := &PriorityQueueNode[HuffmanNode, int]{
			Data: HuffmanNode{
				Left:  left,
				Right: right,
			},
		}
		fmt.Println("size", stack.Len())
		if stack.Len() > 2 {
			right := stack.Pop()
			left := stack.Pop()
			stack.Push()
		} else {
			stack.Push(&PriorityQueueNode[HuffmanNode, int]{
				Data: HuffmanNode{
					Rune: s[i].Data.Rune,
				},
			})
		}
	}
	fmt.Printf("%+v\n", stack)

	textLength, err := binary.ReadUvarint(br)
	if err != nil {
		panic(err)
	}

	// Body
	root := stack.Pop()
	fmt.Println(Render(root))

	// todo: flush / buffer to handle larger files
	sb := &strings.Builder{}
	node := root
	for sb.Len() < int(textLength) {
		b, err := br.ReadByte()
		if err != nil && !errors.Is(err, io.EOF) {
			fmt.Println(err)
			break
		}

		p := 7
		for p >= 0 && sb.Len() < int(textLength) {
			if (b & (1 << p)) == 0 {
				node = node.Data.Left
			} else {
				node = node.Data.Right
			}

			if node.Data.Rune != 0 {
				sb.WriteRune(node.Data.Rune)
				node = root
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

// shoutout tsetso
func encodeVarint(buf []byte, overflow int8, code uint32) ([]byte, int8) {
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
