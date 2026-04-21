package order

import (
	"context"
	"errors"
	"testing"
)

type stubRepo struct {
	getFn    func(ctx context.Context, orderID int64) (*Order, error)
	createFn func(ctx context.Context, order *Order) (int64, error)
}

func (s *stubRepo) GetByID(ctx context.Context, orderID int64) (*Order, error) {
	if s.getFn != nil { return s.getFn(ctx, orderID) }
	return nil, nil
}
func (s *stubRepo) Create(ctx context.Context, order *Order) (int64, error) {
	if s.createFn != nil { return s.createFn(ctx, order) }
	return 0, nil
}

func TestServiceCreateValidation(t *testing.T) {
	svc := NewService(&stubRepo{})
	cases := []struct { uid int64; prod string; amt float64 }{{0, "p", 1}, {1, "", 1}, {1, "p", 0}}
	for _, tc := range cases {
		_, err := svc.Create(context.Background(), tc.uid, tc.prod, tc.amt)
		if !errors.Is(err, ErrInvalidArgument) { t.Fatalf("expected ErrInvalidArgument, got: %v", err) }
	}
}
