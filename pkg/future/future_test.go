package future

import (
	"fmt"
	"testing"
)

func Test_Future_ReadyReturnsFalseWithoutResult(t *testing.T) {
	f := New[int]()
	ok := f.Ready()
	if ok {
		t.Fatalf("f.Ready() returned %t, expected %t", ok, false)
	}
}

func Test_Future_ReadyReturnsTrueWithResult(t *testing.T) {
	f := New[int]()

	exp := 1

	f.SetValue(exp)

	ok := f.Ready()
	if !ok {
		t.Fatalf("f.Ready() returned %t, expected %t", ok, true)
	}

	v, _ := f.Result()

	if v != exp {
		t.Fatalf("f.Peek() returned %d, expected %d", v, exp)
	}
}

func Test_Future_SetValueDoesNotBlock(t *testing.T) {
	f := New[int]()

	v0 := 1
	f.SetValue(v0)

	v1, _ := f.Result()

	if v0 != v1 {
		t.Fatalf("f.Result() returned %d, expected %d", v1, v0)
	}
}

func Test_Future_SetErrorDoesNotBlock(t *testing.T) {
	f := New[int]()

	errOrig := fmt.Errorf("error")
	f.SetError(errOrig)

	_, err := f.Result()

	if err != errOrig {
		t.Fatalf("f.Result() returned %d, expected %d", err, errOrig)
	}
}

func Test_Future_SetValueTwicePanics(t *testing.T) {
	f := New[int]()

	f.SetValue(1)

	defer func() {
		if v := recover(); v == nil {
			t.Fatal("Future value already set once. It must not allow setting it multiple times.")
		}
	}()

	f.SetValue(2)
}

func Test_Future_ResultTwiceSucceeds(t *testing.T) {
	f := New[int]()

	f.SetValue(1)

	_, _ = f.Result()
	v, _ := f.Result()

	if v != 1 {
		t.Fatalf("second f.Result() returned %d, expected %d", v, 1)
	}
}
