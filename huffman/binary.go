package huffman

import (
	"encoding/binary"
	"errors"
	"io"
	"slices"
	"unicode/utf8"

	"github.com/fr3fou/depressor/pq"
	"github.com/fr3fou/depressor/stack"
)

func EncodeTree(w io.Writer, huffman *pq.Node[Node, int]) error {
	if huffman == nil {
		return nil
	}

	if err := encodeNode(w, huffman); err != nil {
		return err
	}

	if err := EncodeTree(w, huffman.Data.Left); err != nil {
		return err
	}

	if err := EncodeTree(w, huffman.Data.Right); err != nil {
		return err
	}

	return nil
}

func encodeNode(w io.Writer, huffman *pq.Node[Node, int]) error {
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

type Scanner interface {
	io.ByteScanner
	io.RuneScanner
}

func Decode(scanner Scanner) (pq.Node[Node, int], error) {
	s := []*pq.Node[Node, int]{}

	for {
		b, err := scanner.ReadByte()
		if err != nil {
			if errors.Is(err, io.EOF) {
				break
			}
			return pq.Node[Node, int]{}, err
		}

		// End of tree
		if b == 0 {
			break
		}

		// Internal node
		if b == 1 {
			s = append(s, &pq.Node[Node, int]{})
			continue
		}

		if err := scanner.UnreadByte(); err != nil {
			return pq.Node[Node, int]{}, err
		}

		// Leaf node
		r, _, err := scanner.ReadRune()
		if err != nil {
			if errors.Is(err, io.EOF) {
				break
			}
			return pq.Node[Node, int]{}, err
		}

		s = append(s, &pq.Node[Node, int]{
			Data: Node{
				Rune: r,
			},
		})
	}

	slices.Reverse(s)
	stack := stack.Stack[*pq.Node[Node, int]]{Data: s}
	return *rebuildTree(&stack), nil
}

func rebuildTree(stack *stack.Stack[*pq.Node[Node, int]]) *pq.Node[Node, int] {
	if stack.Len() < 1 {
		return nil
	}
	node := stack.Pop()
	if node.Data.Rune != 0 {
		return node
	}
	node.Data.Left = rebuildTree(stack)
	node.Data.Right = rebuildTree(stack)
	return node
}
