package app

import (
	"errors"
	"testing"
)

type sample struct{ value string }

type greeter interface{ Greet() string }

type greeterImpl struct{ msg string }

func (g greeterImpl) Greet() string { return g.msg }

func TestContainerProvideAndResolveNamed(t *testing.T) {
	c := newContainer()
	v := &sample{value: "ok"}
	if err := c.ProvideNamed("primary", v); err != nil {
		t.Fatalf("provide named: %v", err)
	}

	var out *sample
	if err := c.ResolveNamed("primary", &out); err != nil {
		t.Fatalf("resolve named: %v", err)
	}
	if out != v {
		t.Fatalf("expected same pointer")
	}
}

func TestContainerResolveAmbiguous(t *testing.T) {
	c := newContainer()
	if err := c.Provide(greeterImpl{msg: "a"}); err != nil {
		t.Fatalf("provide #1: %v", err)
	}
	if err := c.Provide(greeterImpl{msg: "b"}); err != nil {
		t.Fatalf("provide #2: %v", err)
	}

	var g greeter
	if err := c.Resolve(&g); err == nil {
		t.Fatal("expected ambiguity error")
	}
}

func TestContainerInvoke(t *testing.T) {
	c := newContainer()
	if err := c.Provide(&sample{value: "ok"}); err != nil {
		t.Fatalf("provide: %v", err)
	}

	called := false
	err := c.Invoke(func(s *sample) error {
		called = true
		if s.value != "ok" {
			t.Fatalf("unexpected value: %s", s.value)
		}
		return nil
	})
	if err != nil {
		t.Fatalf("invoke: %v", err)
	}
	if !called {
		t.Fatal("expected function to be called")
	}
}

func TestContainerInvokeReturnsError(t *testing.T) {
	c := newContainer()
	want := errors.New("boom")
	if err := c.Invoke(func() error { return want }); !errors.Is(err, want) {
		t.Fatalf("expected %v, got %v", want, err)
	}
}
