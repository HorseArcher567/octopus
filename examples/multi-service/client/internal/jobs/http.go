package jobs

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/HorseArcher567/octopus/pkg/assemble"
	"github.com/HorseArcher567/octopus/pkg/job"
	"github.com/HorseArcher567/octopus/pkg/xlog"
)

func registerHTTPJobs(ctx *assemble.DomainContext, apiURL string) error {
	baseLog := ctx.Logger()

	jobs := map[string]job.Func{
		"http.user_flow": func(runCtx *job.Context) error {
			return runHTTPUserFlow(runCtx.Context(), preferJobLog(runCtx.Logger(), baseLog), apiURL)
		},
		"http.order_flow": func(runCtx *job.Context) error {
			return runHTTPOrderFlow(runCtx.Context(), preferJobLog(runCtx.Logger(), baseLog), apiURL)
		},
		"http.product_flow": func(runCtx *job.Context) error {
			return runHTTPProductFlow(runCtx.Context(), preferJobLog(runCtx.Logger(), baseLog), apiURL)
		},
	}

	for name, fn := range jobs {
		if err := ctx.RegisterJob(name, fn); err != nil {
			return err
		}
	}
	return nil
}

func runHTTPUserFlow(ctx context.Context, log *xlog.Logger, apiURL string) error {
	baseURL, err := trimAPIBaseURL(apiURL)
	if err != nil {
		return fmt.Errorf("invalid api url: %w", err)
	}

	username, email := uniqueUser("http_user")
	createUserResp := struct {
		UserID int64 `json:"user_id"`
	}{}
	if err := doJSON(ctx, http.MethodPost, baseURL+"/users", map[string]any{"username": username, "email": email}, &createUserResp); err != nil {
		return fmt.Errorf("http CreateUser: %w", err)
	}
	if err := doJSON(ctx, http.MethodGet, fmt.Sprintf("%s/users/%d", baseURL, createUserResp.UserID), nil, &struct {
		UserID int64 `json:"user_id"`
	}{}); err != nil {
		return fmt.Errorf("http GetUser: %w", err)
	}
	log.Info("http user flow ok", "user_id", createUserResp.UserID)
	return nil
}

func runHTTPOrderFlow(ctx context.Context, log *xlog.Logger, apiURL string) error {
	baseURL, err := trimAPIBaseURL(apiURL)
	if err != nil {
		return fmt.Errorf("invalid api url: %w", err)
	}

	username, email := uniqueUser("http_order_user")
	createUserResp := struct {
		UserID int64 `json:"user_id"`
	}{}
	if err := doJSON(ctx, http.MethodPost, baseURL+"/users", map[string]any{"username": username, "email": email}, &createUserResp); err != nil {
		return fmt.Errorf("http CreateUser for order flow: %w", err)
	}
	if err := doJSON(ctx, http.MethodPost, baseURL+"/orders", map[string]any{
		"user_id":      createUserResp.UserID,
		"product_name": "http-demo-product",
		"amount":       88.8,
	}, &struct {
		OrderID int64 `json:"order_id"`
	}{}); err != nil {
		return fmt.Errorf("http CreateOrder: %w", err)
	}
	log.Info("http order flow ok", "user_id", createUserResp.UserID)
	return nil
}

func runHTTPProductFlow(ctx context.Context, log *xlog.Logger, apiURL string) error {
	baseURL, err := trimAPIBaseURL(apiURL)
	if err != nil {
		return fmt.Errorf("invalid api url: %w", err)
	}

	if err := doJSON(ctx, http.MethodGet, baseURL+"/products?page=1&page_size=10", nil, &struct {
		Total int32 `json:"total"`
	}{}); err != nil {
		return fmt.Errorf("http ListProducts: %w", err)
	}
	log.Info("http product flow ok")
	return nil
}

func doJSON(ctx context.Context, method, url string, body any, target any) error {
	client := &http.Client{Timeout: 5 * time.Second}

	var reader io.Reader
	if body != nil {
		payload, err := json.Marshal(body)
		if err != nil {
			return err
		}
		reader = bytes.NewReader(payload)
	}

	req, err := http.NewRequestWithContext(ctx, method, url, reader)
	if err != nil {
		return err
	}
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		payload, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("status %d: %s", resp.StatusCode, strings.TrimSpace(string(payload)))
	}
	if target == nil {
		return nil
	}
	return json.NewDecoder(resp.Body).Decode(target)
}

func trimAPIBaseURL(apiURL string) (string, error) {
	trimmed := strings.TrimRight(apiURL, "/")
	if trimmed == "" {
		return "", fmt.Errorf("empty api url")
	}
	return trimmed, nil
}
