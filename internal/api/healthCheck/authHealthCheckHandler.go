package healthcheck_api

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/nihal-ramaswamy/RunnerIO/internal/dto"
	interfaces "github.com/nihal-ramaswamy/RunnerIO/internal/interface"
	auth_middleware "github.com/nihal-ramaswamy/RunnerIO/internal/middlewares/auth"
	"go.uber.org/zap"
)

type AuthHealthCheckHandler struct {
	middlewares []gin.HandlerFunc
	log         *zap.Logger

	interfaces.HandlerInterface
}

func NewAuthHealthCheckHandler(ctx context.Context, log *zap.Logger) *AuthHealthCheckHandler {
	return &AuthHealthCheckHandler{
		log: log,
		middlewares: []gin.HandlerFunc{
			auth_middleware.AuthMiddleware(log),
			auth_middleware.UserInfoMiddleware(ctx, log),
			auth_middleware.ValidatePermissions(log, []string{"read:all"}),
		},
	}
}

func (*AuthHealthCheckHandler) Pattern() string {
	return "/authHealthcheck"
}

func (a *AuthHealthCheckHandler) Handler() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		userDataStr := ctx.GetHeader("userData")

		var userData dto.UserData
		if err := json.Unmarshal([]byte(userDataStr), &userData); err != nil {
			a.log.Error("Failed to unmarshal the user data", zap.Error(err))
			ctx.AbortWithError(http.StatusInternalServerError, err)
		}

		ctx.JSON(http.StatusOK, userData)
	}
}

func (*AuthHealthCheckHandler) RequestMethod() string {
	return http.MethodGet
}

func (h *AuthHealthCheckHandler) Middlewares() []gin.HandlerFunc {
	return h.middlewares
}
