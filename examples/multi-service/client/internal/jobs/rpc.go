package jobs

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/HorseArcher567/octopus/examples/multi-service/proto/pb"
	"github.com/HorseArcher567/octopus/pkg/assemble"
	"github.com/HorseArcher567/octopus/pkg/discovery"
	"github.com/HorseArcher567/octopus/pkg/job"
	"github.com/HorseArcher567/octopus/pkg/rpc"
	"github.com/HorseArcher567/octopus/pkg/store"
	"github.com/HorseArcher567/octopus/pkg/xlog"
	"google.golang.org/grpc"
)

func registerRPCJobs(ctx *assemble.Context, target string) error {
	st := ctx.Store()
	baseLog := ctx.Logger()

	jobs := map[string]job.Func{
		"rpc.user_flow": func(runCtx context.Context, log *xlog.Logger) error {
			return runRPCUserFlow(runCtx, preferJobLog(log, baseLog), st, target)
		},
		"rpc.order_flow": func(runCtx context.Context, log *xlog.Logger) error {
			return runRPCOrderFlow(runCtx, preferJobLog(log, baseLog), st, target)
		},
		"rpc.product_flow": func(runCtx context.Context, log *xlog.Logger) error {
			return runRPCProductFlow(runCtx, preferJobLog(log, baseLog), st, target)
		},
	}

	for name, fn := range jobs {
		if err := ctx.RegisterJob(name, fn); err != nil {
			return err
		}
	}
	return nil
}

func runRPCUserFlow(ctx context.Context, log *xlog.Logger, st store.Store, target string) error {
	conn, err := newRPCClient(log, st, target)
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

func runRPCOrderFlow(ctx context.Context, log *xlog.Logger, st store.Store, target string) error {
	conn, err := newRPCClient(log, st, target)
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

func runRPCProductFlow(ctx context.Context, log *xlog.Logger, st store.Store, target string) error {
	conn, err := newRPCClient(log, st, target)
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

func newRPCClient(log *xlog.Logger, st store.Store, target string) (*grpc.ClientConn, error) {
	dialOpts := []grpc.DialOption{}

	switch {
	case strings.HasPrefix(target, "direct:///"):
		dialOpts = append(dialOpts, grpc.WithResolvers(discovery.NewDirectResolver(log).Builder()))
	}

	clientOptions, err := store.GetNamed[*rpc.ClientOptions](st, "default")
	if err == nil && clientOptions != nil {
		dialOpts = append(dialOpts, clientOptions.BuildDialOptions()...)
	}

	return rpc.NewClient(target, dialOpts...)
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
