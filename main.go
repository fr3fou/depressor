package main

import (
	"bufio"
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"math/bits"
	"os"
	"strings"
)

func main() {
	text := "asdfasdfъ🔥"
	if len(os.Args) > 1 && os.Args[1] == "decompress" {
		decompress(text)
	} else {
		compress(text)
	}
}

func compress(text string) {
	huffman := BuildHuffmanTree(text)
	dictionary := BuildHuffmanDictionary(&huffman)

	file, err := os.Create("text.maikati")
	if err != nil {
		panic(err)
	}
	defer file.Close()

	fmt.Println(Render(&huffman))

	writer := bytes.NewBuffer([]byte{})
	buf := make([]byte, 8)

	// Header
	if err := EncodeHuffmanTree(writer, &huffman); err != nil {
		panic(err)
	}

	if err := binary.Write(writer, binary.BigEndian, byte(0)); err != nil {
		panic(err)
	}

	binary.BigEndian.PutUint64(buf, uint64(len(text)))
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

	for _, b := range writer.Bytes() {
		fmt.Printf("%08b ", b)
	}
	fmt.Println()

	if err := binary.Write(file, binary.BigEndian, writer.Bytes()); err != nil {
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
