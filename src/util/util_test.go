package util_test

import (
	"testing"

	. "github.com/levilutz/basiccoin/src/util"
)

func TestPrepend(t *testing.T) {
	ls := make([]int, 0)
	ls = Prepend(ls, 4)
	ls = Prepend(ls, 3)
	ls = Prepend(ls, 2)
	ls = Prepend(ls, 1)
	ls = Prepend(ls, 0)
	Assert(t, len(ls) == 5, "incorrect length: %d", len(ls))
	for i := 0; i < 5; i++ {
		Assert(t, ls[i] == i, "mismatch on %d: %d", i, ls[i])
	}
}
