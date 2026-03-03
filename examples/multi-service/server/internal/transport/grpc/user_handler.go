package grpc

import (
	"context"

	"github.com/HorseArcher567/octopus/examples/multi-service/proto/pb"
	"github.com/HorseArcher567/octopus/examples/multi-service/server/internal/service"
	"github.com/HorseArcher567/octopus/pkg/xlog"
)

type UserHandler struct {
	pb.UnimplementedUserServer
	svc *service.UserService
	log *xlog.Logger
}

func NewUserHandler(svc *service.UserService, log *xlog.Logger) *UserHandler {
	return &UserHandler{svc: svc, log: log}
}

func (h *UserHandler) GetUser(ctx context.Context, req *pb.GetUserRequest) (*pb.GetUserResponse, error) {
	log := xlog.GetOr(ctx, h.log).With("user_id", req.UserId)
	log.Info("get user")

	user, err := h.svc.GetByID(ctx, req.UserId)
	if err != nil {
		log.Error("get user failed", "error", err)
		return nil, mapError(err, "user not found")
	}

	return &pb.GetUserResponse{UserId: user.ID, Username: user.Username, Email: user.Email}, nil
}

func (h *UserHandler) CreateUser(ctx context.Context, req *pb.CreateUserRequest) (*pb.CreateUserResponse, error) {
	log := xlog.GetOr(ctx, h.log).With("username", req.Username, "email", req.Email)
	log.Info("create user")

	id, err := h.svc.Create(ctx, req.Username, req.Email)
	if err != nil {
		log.Error("create user failed", "error", err)
		return nil, mapError(err, "failed to create user")
	}

	return &pb.CreateUserResponse{UserId: id, Message: "User created successfully"}, nil
}
