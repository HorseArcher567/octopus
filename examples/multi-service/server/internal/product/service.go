package product

import "context"

type Service struct{ repo Repository }

func NewService(repo Repository) *Service { return &Service{repo: repo} }
func (s *Service) GetByID(ctx context.Context, id int64) (*Product, error) { return s.repo.GetByID(ctx, id) }
func (s *Service) List(ctx context.Context, page, pageSize int32) ([]*Product, int64, error) { return s.repo.List(ctx, page, pageSize) }
