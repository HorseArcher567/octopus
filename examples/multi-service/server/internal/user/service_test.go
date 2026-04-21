package user

import (
	"context"
	"errors"
	"testing"
)

type stubRepo struct {
	getFn    func(ctx context.Context, userID int64) (*User, error)
	createFn func(ctx context.Context, user *User) (int64, error)
}

func (s *stubRepo) GetByID(ctx context.Context, userID int64) (*User, error) {
	if s.getFn != nil { return s.getFn(ctx, userID) }
	return nil, nil
}
func (s *stubRepo) Create(ctx context.Context, user *User) (int64, error) {
	if s.createFn != nil { return s.createFn(ctx, user) }
	return 0, nil
}

func TestServiceCreateValidation(t *testing.T) {
	svc := NewService(&stubRepo{})
	if _, err := svc.Create(context.Background(), "", "a@b.com"); !errors.Is(err, ErrInvalidArgument) { t.Fatalf("expected ErrInvalidArgument for empty username, got: %v", err) }
	if _, err := svc.Create(context.Background(), "alice", ""); !errors.Is(err, ErrInvalidArgument) { t.Fatalf("expected ErrInvalidArgument for empty email, got: %v", err) }
}

func TestServiceCreateSuccess(t *testing.T) {
	called := false
	svc := NewService(&stubRepo{createFn: func(_ context.Context, user *User) (int64, error) {
		called = true
		if user.Username != "alice" || user.Email != "alice@example.com" { t.Fatalf("unexpected user payload: %+v", user) }
		return 42, nil
	}})
	id, err := svc.Create(context.Background(), "alice", "alice@example.com")
	if err != nil { t.Fatalf("Create returned error: %v", err) }
	if !called || id != 42 { t.Fatalf("expected repo called with id=42, called=%v id=%d", called, id) }
}
