package user

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/HorseArcher567/octopus/pkg/api"
	"github.com/HorseArcher567/octopus/pkg/xlog"
	"github.com/gin-gonic/gin"
)

type HTTPHandler struct {
	svc *Service
	log *xlog.Logger
}

type createUserRequest struct {
	Username string `json:"username"`
	Email    string `json:"email"`
}

type getUserResponse struct {
	UserID   int64  `json:"user_id"`
	Username string `json:"username"`
	Email    string `json:"email"`
}

type createUserResponse struct {
	UserID  int64  `json:"user_id"`
	Message string `json:"message"`
}

type errorResponse struct {
	Error string `json:"error"`
}

func NewHTTPHandler(svc *Service, log *xlog.Logger) *HTTPHandler {
	return &HTTPHandler{svc: svc, log: log}
}

func RegisterHTTP(engine *api.Engine, h *HTTPHandler) {
	engine.GET("/users/:id", h.GetUser)
	engine.POST("/users", h.CreateUser)
}

func (h *HTTPHandler) GetUser(c *gin.Context) {
	userID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, errorResponse{Error: "invalid user id"})
		return
	}

	ctx := c.Request.Context()
	log := xlog.GetOr(ctx, h.log).With("user_id", userID)
	log.Info("get user")

	user, err := h.svc.GetByID(ctx, userID)
	if err != nil {
		log.Error("get user failed", "error", err)
		writeHTTPError(c, err, "user not found")
		return
	}

	c.JSON(http.StatusOK, getUserResponse{UserID: user.ID, Username: user.Username, Email: user.Email})
}

func (h *HTTPHandler) CreateUser(c *gin.Context) {
	var req createUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, errorResponse{Error: "invalid request body"})
		return
	}

	ctx := c.Request.Context()
	log := xlog.GetOr(ctx, h.log).With("username", req.Username, "email", req.Email)
	log.Info("create user")

	id, err := h.svc.Create(ctx, req.Username, req.Email)
	if err != nil {
		log.Error("create user failed", "error", err)
		writeHTTPError(c, err, "failed to create user")
		return
	}

	c.JSON(http.StatusOK, createUserResponse{UserID: id, Message: "User created successfully"})
}

func writeHTTPError(c *gin.Context, err error, notFoundMsg string) {
	if err == nil {
		return
	}
	switch {
	case errors.Is(err, ErrNotFound):
		c.JSON(http.StatusNotFound, errorResponse{Error: notFoundMsg})
	case errors.Is(err, ErrInvalidArgument):
		c.JSON(http.StatusBadRequest, errorResponse{Error: err.Error()})
	default:
		c.JSON(http.StatusInternalServerError, errorResponse{Error: err.Error()})
	}
}
