package service

import (
	"context"
	"errors"
	"testing"

	"github.com/HorseArcher567/octopus/examples/multi-service/server/internal/domain"
)

type stubUserRepo struct {
	getFn    func(ctx context.Context, userID int64) (*domain.User, error)
	createFn func(ctx context.Context, user *domain.User) (int64, error)
}

func (s *stubUserRepo) GetByID(ctx context.Context, userID int64) (*domain.User, error) {
	if s.getFn != nil {
		return s.getFn(ctx, userID)
	}
	return nil, nil
}

func (s *stubUserRepo) Create(ctx context.Context, user *domain.User) (int64, error) {
	if s.createFn != nil {
		return s.createFn(ctx, user)
	}
	return 0, nil
}

type stubOrderRepo struct {
	getFn    func(ctx context.Context, orderID int64) (*domain.Order, error)
	createFn func(ctx context.Context, order *domain.Order) (int64, error)
}

func (s *stubOrderRepo) GetByID(ctx context.Context, orderID int64) (*domain.Order, error) {
	if s.getFn != nil {
		return s.getFn(ctx, orderID)
	}
	return nil, nil
}

func (s *stubOrderRepo) Create(ctx context.Context, order *domain.Order) (int64, error) {
	if s.createFn != nil {
		return s.createFn(ctx, order)
	}
	return 0, nil
}

type stubProductRepo struct {
	getFn  func(ctx context.Context, productID int64) (*domain.Product, error)
	listFn func(ctx context.Context, page, pageSize int32) ([]*domain.Product, int64, error)
}

func (s *stubProductRepo) GetByID(ctx context.Context, productID int64) (*domain.Product, error) {
	if s.getFn != nil {
		return s.getFn(ctx, productID)
	}
	return nil, nil
}

func (s *stubProductRepo) List(ctx context.Context, page, pageSize int32) ([]*domain.Product, int64, error) {
	if s.listFn != nil {
		return s.listFn(ctx, page, pageSize)
	}
	return nil, 0, nil
}

func TestUserServiceCreateValidation(t *testing.T) {
	svc := NewUserService(&stubUserRepo{})

	if _, err := svc.Create(context.Background(), "", "a@b.com"); !errors.Is(err, domain.ErrInvalidArgument) {
		t.Fatalf("expected ErrInvalidArgument for empty username, got: %v", err)
	}
	if _, err := svc.Create(context.Background(), "alice", ""); !errors.Is(err, domain.ErrInvalidArgument) {
		t.Fatalf("expected ErrInvalidArgument for empty email, got: %v", err)
	}
}

func TestUserServiceCreateSuccess(t *testing.T) {
	called := false
	svc := NewUserService(&stubUserRepo{createFn: func(_ context.Context, user *domain.User) (int64, error) {
		called = true
		if user.Username != "alice" || user.Email != "alice@example.com" {
			t.Fatalf("unexpected user payload: %+v", user)
		}
		return 42, nil
	}})

	id, err := svc.Create(context.Background(), "alice", "alice@example.com")
	if err != nil {
		t.Fatalf("Create returned error: %v", err)
	}
	if !called || id != 42 {
		t.Fatalf("expected repo called with id=42, called=%v id=%d", called, id)
	}
}

func TestOrderServiceCreateValidation(t *testing.T) {
	svc := NewOrderService(&stubOrderRepo{})

	cases := []struct {
		name string
		uid  int64
		prod string
		amt  float64
	}{
		{name: "invalid user id", uid: 0, prod: "p", amt: 1},
		{name: "empty product", uid: 1, prod: "", amt: 1},
		{name: "invalid amount", uid: 1, prod: "p", amt: 0},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := svc.Create(context.Background(), tc.uid, tc.prod, tc.amt)
			if !errors.Is(err, domain.ErrInvalidArgument) {
				t.Fatalf("expected ErrInvalidArgument, got: %v", err)
			}
		})
	}
}

func TestProductServiceListDelegates(t *testing.T) {
	expected := []*domain.Product{{ProductID: 1, Name: "book"}}
	svc := NewProductService(&stubProductRepo{listFn: func(_ context.Context, page, size int32) ([]*domain.Product, int64, error) {
		if page != 2 || size != 5 {
			t.Fatalf("unexpected paging: page=%d size=%d", page, size)
		}
		return expected, 10, nil
	}})

	items, total, err := svc.List(context.Background(), 2, 5)
	if err != nil {
		t.Fatalf("List returned error: %v", err)
	}
	if total != 10 || len(items) != 1 || items[0].ProductID != 1 {
		t.Fatalf("unexpected list result: total=%d items=%+v", total, items)
	}
}
