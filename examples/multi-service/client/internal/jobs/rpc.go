package jobs

import (
	"context"
	"fmt"
	"time"

	"github.com/HorseArcher567/octopus/examples/multi-service/proto/pb"
	"github.com/HorseArcher567/octopus/pkg/assemble"
	"github.com/HorseArcher567/octopus/pkg/job"
	"github.com/HorseArcher567/octopus/pkg/rpc"
	"github.com/HorseArcher567/octopus/pkg/xlog"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func registerRPCJobs(ctx *assemble.DomainContext, target string) error {
	baseLog := ctx.Logger()

	jobs := map[string]job.Func{
		"rpc.user_flow": func(runCtx *job.Context) error {
			return runRPCUserFlow(runCtx.Context(), preferJobLog(runCtx.Logger(), baseLog), target)
		},
		"rpc.order_flow": func(runCtx *job.Context) error {
			return runRPCOrderFlow(runCtx.Context(), preferJobLog(runCtx.Logger(), baseLog), target)
		},
		"rpc.product_flow": func(runCtx *job.Context) error {
			return runRPCProductFlow(runCtx.Context(), preferJobLog(runCtx.Logger(), baseLog), target)
		},
	}

	for name, fn := range jobs {
		if err := ctx.RegisterJob(name, fn); err != nil {
			return err
		}
	}
	return nil
}

func runRPCUserFlow(ctx context.Context, log *xlog.Logger, target string) error {
	conn, err := newRPCClient(log, target)
	if err != nil {
		return fmt.Errorf("new rpc client: %w", err)
	}
	defer conn.Close()

	userClient := pb.NewUserClient(conn)
	username, email := uniqueUser("rpc_user")

	createUserResp, err := userClient.CreateUser(ctx, &pb.CreateUserRequest{Username: username, Email: email})
	if err != nil {
		return fmt.Errorf("CreateUser: %w", err)
	}
	log.Info("rpc user flow create ok", "user_id", createUserResp.UserId)

	if _, err := userClient.GetUser(ctx, &pb.GetUserRequest{UserId: createUserResp.UserId}); err != nil {
		return fmt.Errorf("GetUser: %w", err)
	}
	log.Info("rpc user flow get ok", "user_id", createUserResp.UserId)
	return nil
}

func runRPCOrderFlow(ctx context.Context, log *xlog.Logger, target string) error {
	conn, err := newRPCClient(log, target)
	if err != nil {
		return fmt.Errorf("new rpc client: %w", err)
	}
	defer conn.Close()

	userClient := pb.NewUserClient(conn)
	orderClient := pb.NewOrderClient(conn)
	username, email := uniqueUser("rpc_order_user")

	createUserResp, err := userClient.CreateUser(ctx, &pb.CreateUserRequest{Username: username, Email: email})
	if err != nil {
		return fmt.Errorf("CreateUser for order flow: %w", err)
	}
	if _, err := orderClient.CreateOrder(ctx, &pb.CreateOrderRequest{UserId: createUserResp.UserId, ProductName: "demo-product", Amount: 99.9}); err != nil {
		return fmt.Errorf("CreateOrder: %w", err)
	}
	log.Info("rpc order flow ok", "user_id", createUserResp.UserId)
	return nil
}

func runRPCProductFlow(ctx context.Context, log *xlog.Logger, target string) error {
	conn, err := newRPCClient(log, target)
	if err != nil {
		return fmt.Errorf("new rpc client: %w", err)
	}
	defer conn.Close()

	productClient := pb.NewProductClient(conn)
	if _, err := productClient.ListProducts(ctx, &pb.ListProductsRequest{Page: 1, PageSize: 10}); err != nil {
		return fmt.Errorf("ListProducts: %w", err)
	}
	log.Info("rpc product flow ok")
	return nil
}

func newRPCClient(log *xlog.Logger, target string) (*grpc.ClientConn, error) {
	_ = log
	return rpc.NewClient(target, grpc.WithTransportCredentials(insecure.NewCredentials()))
}

func uniqueUser(prefix string) (string, string) {
	suffix := time.Now().UnixNano()
	username := fmt.Sprintf("%s_%d", prefix, suffix)
	email := fmt.Sprintf("%s_%d@example.com", prefix, suffix)
	return username, email
}

func preferJobLog(jobLog, fallback *xlog.Logger) *xlog.Logger {
	if jobLog != nil {
		return jobLog
	}
	return fallback
}
