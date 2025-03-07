package types

type Pair[A any, B any] struct {
	Fst A
	Snd B
}

func MakePair[A any, B any](a A, b B) Pair[A, B] {
	return Pair[A, B]{Fst: a, Snd: b}
}

type Set[T comparable] map[T]struct{}

func (set Set[T]) Pop() T {
	var nothing T
	for e := range set {
		delete(set, e)
		return e
	}
	return nothing
}
