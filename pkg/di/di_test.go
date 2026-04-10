package di

import (
	"errors"
	"testing"
)

type sample struct{ value string }

type greeter interface{ Greet() string }

type greeterImpl struct{ msg string }

func (g greeterImpl) Greet() string { return g.msg }

func TestContainerProvideAndResolveNamed(t *testing.T) {
	c := New()
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
	c := New()
	if err := c.Provide(greeterImpl{msg: "a"}); err != nil {
		t.Fatalf("provide #1: %v", err)
	}
	if err := c.Provide(greeterImpl{msg: "b"}); err != nil {
		t.Fatalf("provide #2: %v", err)
	}

	var g greeter
	err := c.Resolve(&g)
	if !errors.Is(err, ErrAmbiguous) {
		t.Fatalf("expected ambiguity error, got %v", err)
	}
}

func TestContainerResolveAll(t *testing.T) {
	c := New()
	if err := c.Provide(greeterImpl{msg: "a"}); err != nil {
		t.Fatalf("provide #1: %v", err)
	}
	if err := c.Provide(greeterImpl{msg: "b"}); err != nil {
		t.Fatalf("provide #2: %v", err)
	}

	var values []greeter
	if err := c.ResolveAll(&values); err != nil {
		t.Fatalf("resolve all: %v", err)
	}
	if len(values) != 2 {
		t.Fatalf("expected 2 values, got %d", len(values))
	}
	if values[0].Greet() != "a" || values[1].Greet() != "b" {
		t.Fatalf("unexpected order: %q, %q", values[0].Greet(), values[1].Greet())
	}
}

func TestContainerResolveAllNamed(t *testing.T) {
	c := New()
	if err := c.ProvideNamed("workers", greeterImpl{msg: "a"}); err != nil {
		t.Fatalf("provide named #1: %v", err)
	}
	if err := c.ProvideNamed("workers", greeterImpl{msg: "b"}); err != nil {
		t.Fatalf("provide named #2: %v", err)
	}

	var values []greeter
	if err := c.ResolveAllNamed("workers", &values); err != nil {
		t.Fatalf("resolve all named: %v", err)
	}
	if len(values) != 2 {
		t.Fatalf("expected 2 values, got %d", len(values))
	}
	if values[0].Greet() != "a" || values[1].Greet() != "b" {
		t.Fatalf("unexpected order: %q, %q", values[0].Greet(), values[1].Greet())
	}
}

func TestContainerResolveExactMatchPreferred(t *testing.T) {
	c := New()
	exact := greeterImpl{msg: "exact"}
	if err := c.Provide(exact); err != nil {
		t.Fatalf("provide exact: %v", err)
	}
	if err := c.Provide(&sample{value: "irrelevant"}); err != nil {
		t.Fatalf("provide unrelated: %v", err)
	}

	var out greeterImpl
	if err := c.Resolve(&out); err != nil {
		t.Fatalf("resolve exact: %v", err)
	}
	if out.msg != "exact" {
		t.Fatalf("expected exact match, got %q", out.msg)
	}
}

func TestContainerResolveNotFound(t *testing.T) {
	c := New()
	var out *sample
	err := c.Resolve(&out)
	if !errors.Is(err, ErrNotFound) {
		t.Fatalf("expected not found error, got %v", err)
	}
}

func TestContainerResolveInvalidTarget(t *testing.T) {
	c := New()
	if err := c.Resolve(nil); !errors.Is(err, ErrInvalidTarget) {
		t.Fatalf("expected invalid target for nil, got %v", err)
	}
	var out sample
	if err := c.Resolve(out); !errors.Is(err, ErrInvalidTarget) {
		t.Fatalf("expected invalid target for non-pointer, got %v", err)
	}
	if err := c.ResolveAll(&out); !errors.Is(err, ErrInvalidTarget) {
		t.Fatalf("expected invalid target for non-slice pointer, got %v", err)
	}
}

func TestContainerInvoke(t *testing.T) {
	c := New()
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
	c := New()
	want := errors.New("boom")
	if err := c.Invoke(func() error { return want }); !errors.Is(err, want) {
		t.Fatalf("expected %v, got %v", want, err)
	}
}
