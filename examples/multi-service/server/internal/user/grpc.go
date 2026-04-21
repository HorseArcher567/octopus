package user

import (
	"context"
	"errors"

	"github.com/HorseArcher567/octopus/examples/multi-service/proto/pb"
	"github.com/HorseArcher567/octopus/pkg/xlog"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type GRPCHandler struct {
	pb.UnimplementedUserServer
	svc *Service
	log *xlog.Logger
}

func NewGRPCHandler(svc *Service, log *xlog.Logger) *GRPCHandler {
	return &GRPCHandler{svc: svc, log: log}
}

func RegisterGRPC(s *grpc.Server, h *GRPCHandler) {
	pb.RegisterUserServer(s, h)
}

func (h *GRPCHandler) GetUser(ctx context.Context, req *pb.GetUserRequest) (*pb.GetUserResponse, error) {
	log := xlog.GetOr(ctx, h.log).With("user_id", req.UserId)
	log.Info("get user")

	user, err := h.svc.GetByID(ctx, req.UserId)
	if err != nil {
		log.Error("get user failed", "error", err)
		return nil, mapGRPCError(err, "user not found")
	}
	return &pb.GetUserResponse{UserId: user.ID, Username: user.Username, Email: user.Email}, nil
}

func (h *GRPCHandler) CreateUser(ctx context.Context, req *pb.CreateUserRequest) (*pb.CreateUserResponse, error) {
	log := xlog.GetOr(ctx, h.log).With("username", req.Username, "email", req.Email)
	log.Info("create user")

	id, err := h.svc.Create(ctx, req.Username, req.Email)
	if err != nil {
		log.Error("create user failed", "error", err)
		return nil, mapGRPCError(err, "failed to create user")
	}
	return &pb.CreateUserResponse{UserId: id, Message: "User created successfully"}, nil
}

func mapGRPCError(err error, notFoundMsg string) error {
	if err == nil {
		return nil
	}
	if errors.Is(err, ErrNotFound) {
		return status.Error(codes.NotFound, notFoundMsg)
	}
	if errors.Is(err, ErrInvalidArgument) {
		return status.Error(codes.InvalidArgument, err.Error())
	}
	return status.Error(codes.Internal, err.Error())
}
