package util

// Get keys from map.
func MapKeys[K comparable, V any](in map[K]V) []K {
	out := make([]K, len(in))
	i := 0
	for k := range in {
		out[i] = k
		i++
	}
	return out
}

// Prepend into slice.
func Prepend[K any](ls []K, items ...K) []K {
	for _, item := range items {
		if len(ls) == 0 {
			ls = []K{item}
		} else {
			ls = append(ls, item)
			copy(ls[1:], ls)
			ls[0] = item
		}
	}
	return ls
}

// Reverse a slice. Does not modify in-place, returns a new one.
func Reverse[K any](ls []K) []K {
	out := make([]K, len(ls))
	for i, item := range ls {
		out[len(ls)-i-1] = item
	}
	return out
}

// Shallow copy a map.
func CopyMap[K comparable, V any](m map[K]V) map[K]V {
	out := make(map[K]V, len(m))
	for k, v := range m {
		out[k] = v
	}
	return out
}

// Shallow copy a list.
func CopyList[K any](ls []K) []K {
	out := make([]K, len(ls))
	copy(out, ls)
	return out
}

// Flatten a double list.
func FlattenLists[K comparable](in [][]K) []K {
	out := make([]K, 0)
	for _, inL := range in {
		out = append(out, inL...)
	}
	return out
}

// Write to a channel, don't care if it fails.
func WriteChIfPossible[K any](ch chan<- K, val K) {
	go func() {
		defer func() { recover() }()
		ch <- val
		close(ch)
	}()
}
