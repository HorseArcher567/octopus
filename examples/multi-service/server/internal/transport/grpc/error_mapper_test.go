package grpc

import (
	"errors"
	"fmt"
	"testing"

	"github.com/HorseArcher567/octopus/examples/multi-service/server/internal/domain"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestMapErrorNil(t *testing.T) {
	if got := mapError(nil, "not found"); got != nil {
		t.Fatalf("expected nil, got %v", got)
	}
}

func TestMapErrorNotFound(t *testing.T) {
	err := fmt.Errorf("wrapped: %w", domain.ErrNotFound)
	mapped := mapError(err, "resource missing")

	st := status.Convert(mapped)
	if st.Code() != codes.NotFound || st.Message() != "resource missing" {
		t.Fatalf("unexpected status: code=%s msg=%q", st.Code(), st.Message())
	}
}

func TestMapErrorInvalidArgument(t *testing.T) {
	err := fmt.Errorf("bad input: %w", domain.ErrInvalidArgument)
	mapped := mapError(err, "unused")

	st := status.Convert(mapped)
	if st.Code() != codes.InvalidArgument {
		t.Fatalf("unexpected code: %s", st.Code())
	}
	if st.Message() != err.Error() {
		t.Fatalf("expected message %q, got %q", err.Error(), st.Message())
	}
}

func TestMapErrorInternalFallback(t *testing.T) {
	err := errors.New("db down")
	mapped := mapError(err, "unused")

	st := status.Convert(mapped)
	if st.Code() != codes.Internal || st.Message() != err.Error() {
		t.Fatalf("unexpected status: code=%s msg=%q", st.Code(), st.Message())
	}
}
