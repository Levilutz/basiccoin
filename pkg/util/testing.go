package util

import "testing"

func Assert(t *testing.T, condition bool, msg string, v ...interface{}) {
	if !condition {
		t.Fatalf(msg, v...)
	}
}

func AssertNoErr(t *testing.T, err error) {
	errStr := ""
	if err != nil {
		errStr = err.Error()
	}
	Assert(t, err == nil, "unexpected error: %s", errStr)
}
