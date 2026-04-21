package user

import (
	"errors"
	"fmt"
	"testing"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestMapGRPCErrorNil(t *testing.T) {
	if got := mapGRPCError(nil, "not found"); got != nil { t.Fatalf("expected nil, got %v", got) }
}
func TestMapGRPCErrorNotFound(t *testing.T) {
	err := fmt.Errorf("wrapped: %w", ErrNotFound)
	st := status.Convert(mapGRPCError(err, "resource missing"))
	if st.Code() != codes.NotFound || st.Message() != "resource missing" { t.Fatalf("unexpected status: code=%s msg=%q", st.Code(), st.Message()) }
}
func TestMapGRPCErrorInvalidArgument(t *testing.T) {
	err := fmt.Errorf("bad input: %w", ErrInvalidArgument)
	st := status.Convert(mapGRPCError(err, "unused"))
	if st.Code() != codes.InvalidArgument || st.Message() != err.Error() { t.Fatalf("unexpected status: code=%s msg=%q", st.Code(), st.Message()) }
}
func TestMapGRPCErrorInternalFallback(t *testing.T) {
	err := errors.New("db down")
	st := status.Convert(mapGRPCError(err, "unused"))
	if st.Code() != codes.Internal || st.Message() != err.Error() { t.Fatalf("unexpected status: code=%s msg=%q", st.Code(), st.Message()) }
}
