package main

import (
	"bufio"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"log"
	"math/bits"
	"os"
	"strings"

	"github.com/fr3fou/depressor/huffman"
)

func main() {
	if err := run(); err != nil {
		log.Fatalln(err)
	}
}

func usage() {
	fmt.Println("usage: ")
	fmt.Println("	depressor -c file.txt")
	fmt.Println("	depressor -d file.txt.mkti")
	fmt.Println("	depressor -help")
}

func run() error {
	if len(os.Args) < 3 {
		usage()
		return nil
	}

	cmd := os.Args[1]
	fileName := os.Args[2]
	input, err := os.Open(fileName)
	if err != nil {
		return fmt.Errorf("failed opening file for reading: %w", err)
	}
	defer input.Close()

	if cmd == "-c" {
		return compress(fileName, input)
	}

	if cmd == "-d" {
		return decompress(fileName, input)
	}

	usage()
	return nil
}

func compress(fileName string, reader io.Reader) error {
	sb := &strings.Builder{}
	io.Copy(sb, reader)

	text := sb.String()
	tree := huffman.BuildTree(text)
	dictionary := huffman.BuildDictionary(tree)

	output, err := os.Create(fmt.Sprintf("%s.mkti", fileName))
	if err != nil {
		return err
	}
	defer output.Close()

	writer := bufio.NewWriter(output)

	if err := huffman.EncodeTree(writer, &tree); err != nil {
		return nil
	}

	if err := binary.Write(writer, binary.BigEndian, byte(0)); err != nil {
		return nil
	}

	buf := []byte{}
	buf = binary.AppendUvarint(buf, uint64(len(text)))
	_, err = writer.Write(buf)
	if err != nil {
		return err
	}

	var overflow int8
	buf = []byte{}
	for _, r := range text {
		code := dictionary[r]
		buf, overflow = encode(buf, overflow, code)
	}

	_, err = writer.Write(buf)
	if err != nil {
		return nil
	}

	if err := writer.Flush(); err != nil {
		return nil
	}
	return nil
}

func decompress(fileName string, reader io.Reader) error {
	br := bufio.NewReader(reader)
	root, err := huffman.Decode(br)
	if err != nil {
		return err
	}

	textLength, err := binary.ReadUvarint(br)
	if err != nil {
		return err
	}

	output, err := os.Create(strings.ReplaceAll(fileName, ".mkti", ""))
	if err != nil {
		return fmt.Errorf("failed opening file for writing: %w", err)
	}
	defer output.Close()

	sb := &strings.Builder{}
	node := root
	for sb.Len() < int(textLength) {
		b, err := br.ReadByte()
		if err != nil && !errors.Is(err, io.EOF) {
			return err
		}

		p := 7
		for p >= 0 && sb.Len() < int(textLength) {
			if (b & (1 << p)) == 0 {
				node = *node.Data.Left
			} else {
				node = *node.Data.Right
			}

			if node.Data.Rune != 0 {
				sb.WriteRune(node.Data.Rune)
				node = root
			}
			p--
		}

		if err != nil {
			break
		}
	}

	_, err = output.WriteString(sb.String())
	return err
}

// shoutout tsetso
func encode(buf []byte, overflow int8, code uint32) ([]byte, int8) {
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
