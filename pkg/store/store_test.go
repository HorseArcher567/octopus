package store

import (
	"errors"
	"testing"
)

type sample struct{ value string }

func TestContainerSetAndGetNamed(t *testing.T) {
	c := New()
	v := &sample{value: "ok"}
	if err := c.SetNamed("primary", v); err != nil {
		t.Fatalf("set named: %v", err)
	}
	out, err := GetNamed[*sample](c, "primary")
	if err != nil {
		t.Fatalf("get named: %v", err)
	}
	if out != v {
		t.Fatalf("expected same pointer")
	}
}

func TestContainerSetAndGet(t *testing.T) {
	c := New()
	v := &sample{value: "ok"}
	if err := c.Set(v); err != nil {
		t.Fatalf("set: %v", err)
	}
	out, err := Get[*sample](c)
	if err != nil {
		t.Fatalf("get: %v", err)
	}
	if out != v {
		t.Fatalf("expected same pointer")
	}
}

func TestContainerDuplicate(t *testing.T) {
	c := New()
	if err := c.SetNamed("primary", &sample{value: "a"}); err != nil {
		t.Fatalf("set first: %v", err)
	}
	err := c.SetNamed("primary", &sample{value: "b"})
	if !errors.Is(err, ErrDuplicate) {
		t.Fatalf("expected duplicate error, got %v", err)
	}
}

func TestContainerNotFound(t *testing.T) {
	c := New()
	_, err := Get[*sample](c)
	if !errors.Is(err, ErrNotFound) {
		t.Fatalf("expected not found error, got %v", err)
	}
}

func TestContainerInvalidValue(t *testing.T) {
	c := New()
	if err := c.Set(nil); !errors.Is(err, ErrInvalidValue) {
		t.Fatalf("expected invalid value for nil, got %v", err)
	}
	var s *sample
	if err := c.Set(s); !errors.Is(err, ErrInvalidValue) {
		t.Fatalf("expected invalid value for typed nil, got %v", err)
	}
}

func TestContainerClose(t *testing.T) {
	c := New()
	closed := false
	want := errors.New("boom")
	if err := c.SetNamed("primary", &sample{value: "ok"}, WithClose(func() error {
		closed = true
		return want
	})); err != nil {
		t.Fatalf("set named: %v", err)
	}
	if err := c.Close(); !errors.Is(err, want) {
		t.Fatalf("expected close error %v, got %v", want, err)
	}
	if !closed {
		t.Fatal("expected close fn to run")
	}
}
