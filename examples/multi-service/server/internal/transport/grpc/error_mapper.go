package grpc

import (
	"errors"

	"github.com/HorseArcher567/octopus/examples/multi-service/server/internal/domain"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func mapError(err error, notFoundMsg string) error {
	if err == nil {
		return nil
	}
	if errors.Is(err, domain.ErrNotFound) {
		return status.Error(codes.NotFound, notFoundMsg)
	}
	if errors.Is(err, domain.ErrInvalidArgument) {
		return status.Error(codes.InvalidArgument, err.Error())
	}
	return status.Error(codes.Internal, err.Error())
}
