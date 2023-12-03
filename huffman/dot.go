package huffman

import (
	"fmt"
	"math/rand"

	"github.com/emicklei/dot"
	"github.com/fr3fou/depressor/pq"
)

func Render(huffman pq.Node[Node, int]) string {
	graph := dot.NewGraph(dot.Directed)
	// This preserves the order in which they were added
	// Otherwise, left & right _could_ be swapped
	graph.Attr("compound", "true")
	graph.Attr("ordering", "out")
	root := graph.Node("Root")
	root.Label(fmt.Sprintf("%d", huffman.Value))
	render(huffman, graph, root)
	return graph.String()
}

func render(huffman pq.Node[Node, int], graph *dot.Graph, node dot.Node) {
	if huffman.Data.Left != nil {
		r := huffman.Data.Left.Data.Rune
		leftNode := graph.Node(randomID(5))
		label := fmt.Sprintf("%d", huffman.Data.Left.Value)
		if r != 0 {
			label = fmt.Sprintf(`'%s', (%s)`, string(r), label)
		}
		leftNode.Label(label)
		render(*huffman.Data.Left, graph, leftNode)
		node.Edge(leftNode)
	}

	if huffman.Data.Right != nil {
		r := huffman.Data.Right.Data.Rune
		rightNode := graph.Node(randomID(5))
		label := fmt.Sprintf("%d", huffman.Data.Right.Value)
		if r != 0 {
			label = fmt.Sprintf(`'%s' (%s)`, string(r), label)
		}
		rightNode.Label(label)
		render(*huffman.Data.Right, graph, rightNode)
		node.Edge(rightNode)
	}
}

const letterBytes = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"

func randomID(n int) string {
	b := make([]byte, n)
	for i := range b {
		b[i] = letterBytes[rand.Int63()%int64(len(letterBytes))]
	}
	return string(b)
}
