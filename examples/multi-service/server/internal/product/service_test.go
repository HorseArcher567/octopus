package product

import (
	"context"
	"testing"
)

type stubRepo struct {
	getFn  func(ctx context.Context, productID int64) (*Product, error)
	listFn func(ctx context.Context, page, pageSize int32) ([]*Product, int64, error)
}

func (s *stubRepo) GetByID(ctx context.Context, productID int64) (*Product, error) {
	if s.getFn != nil {
		return s.getFn(ctx, productID)
	}
	return nil, nil
}
func (s *stubRepo) List(ctx context.Context, page, pageSize int32) ([]*Product, int64, error) {
	if s.listFn != nil {
		return s.listFn(ctx, page, pageSize)
	}
	return nil, 0, nil
}

func TestServiceListDelegates(t *testing.T) {
	expected := []*Product{{ProductID: 1, Name: "book"}}
	svc := NewService(&stubRepo{listFn: func(_ context.Context, page, size int32) ([]*Product, int64, error) {
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
