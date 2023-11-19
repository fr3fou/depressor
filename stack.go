package main

type Stack[T any] struct {
	Data []T
}

func (s *Stack[T]) Push(value T) {
	s.Data = append(s.Data, value)
}

func (s *Stack[T]) Peek() T {
	return s.Data[len(s.Data)-1]
}

func (s *Stack[T]) Pop() T {
	value := s.Peek()
	s.Data = s.Data[:len(s.Data)-1]
	return value
}

func (s *Stack[T]) Len() int {
	return len(s.Data)
}
