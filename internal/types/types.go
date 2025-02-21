package types

type Pair[A any, B any] struct {
	Fst A
	Snd B
}

type Set[T comparable] map[T]struct{}

func (set Set[T]) GetFirst() T {
	var nothing T
	for e := range set {
		delete(set, e)
		return e
	}
	return nothing
}
