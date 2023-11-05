package main

import "cmp"

// PriorityQueue is a min heap - we have the smallest value at the top
// Node children contain values bigger than their parent.
//
//		     5
//		   /   \
//		  6     8
//		 /
//	    9
//
// Pushing appends a new item to the end of the array and then sifts it up.
// Sifting up means moving the newly added element, such that the heap property still applies.
// By constantly swapping nodes with their parents, such that the parents
// are the smaller number, we will eventually have a valid heap.
//
// Popping returns the topmost element and puts the bottommost element to the top and sifts it down
// until it finds its place. Sifting down means swapping the current node with the smaller of its 2 children.
// We repeat this until the node is smaller than its 2 children.
type Node[T any, U cmp.Ordered] struct {
	Data  T
	Value U
}

type PriorityQueue[T any, U cmp.Ordered] struct {
	Heap []Node[T, U]
}

func NewPriorityQueue[T any, U cmp.Ordered]() *PriorityQueue[T, U] {
	return &PriorityQueue[T, U]{
		Heap: []Node[T, U]{},
	}
}
func (pq *PriorityQueue[T, U]) Empty() bool {
	return len(pq.Heap) == 0
}

func (pq *PriorityQueue[T, U]) Pop() (n Node[T, U], ok bool) {
	if pq.Empty() {
		ok = false
		return
	}

	top := pq.Heap[0]
	pq.Heap[0] = pq.Heap[len(pq.Heap)-1]
	pq.Heap = pq.Heap[:len(pq.Heap)-1]
	i := 0
	for {
		left := i*2 + 1
		if left >= len(pq.Heap) {
			break
		}

		right := i*2 + 2
		target := left

		if right < len(pq.Heap) && pq.Heap[left].Value >= pq.Heap[right].Value {
			target = right
		}

		if pq.Heap[i].Value < pq.Heap[target].Value {
			break
		}

		pq.Heap[i], pq.Heap[target] = pq.Heap[target], pq.Heap[i]
		i = target
	}
	return top, true
}

func (pq *PriorityQueue[T, U]) Push(item Node[T, U]) {
	pq.Heap = append(pq.Heap, item)
	i := len(pq.Heap) - 1
	for i > 0 {
		parent := (i - 1) / 2
		if pq.Heap[parent].Value <= pq.Heap[i].Value {
			break
		}
		pq.Heap[parent], pq.Heap[i] = pq.Heap[i], pq.Heap[parent]
		i = parent
	}
}
