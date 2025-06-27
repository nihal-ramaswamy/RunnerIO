package auth_middleware

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/url"
	"time"

	jwtmiddleware "github.com/auth0/go-jwt-middleware/v2"
	"github.com/auth0/go-jwt-middleware/v2/jwks"
	"github.com/auth0/go-jwt-middleware/v2/validator"
	"github.com/gin-gonic/gin"
	adapter "github.com/gwatts/gin-adapter"
	"github.com/nihal-ramaswamy/RunnerIO/internal/constants"
	"github.com/nihal-ramaswamy/RunnerIO/internal/dto"
	"github.com/nihal-ramaswamy/RunnerIO/internal/utils"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

func AuthMiddleware(log *zap.Logger) gin.HandlerFunc {
	audience := utils.GetDotEnvVariable("AUTH0_AUDIENCE")
	authority := utils.GetDotEnvVariable("AUTH0_AUTHORITY")

	issuerURL, err := url.Parse(authority)
	if err != nil {
		log.Fatal("Failed to parse the issuer url", zap.Error(err))
	}

	provider := jwks.NewCachingProvider(issuerURL, 5*time.Minute)

	// Set up the validator.
	jwtValidator, err := validator.New(
		provider.KeyFunc,
		validator.RS256,
		authority,
		[]string{audience},
	)
	if err != nil {
		log.Fatal("failed to set up the validator", zap.Error(err))
	}

	// Set up the middleware.
	middleware := jwtmiddleware.New(jwtValidator.ValidateToken)

	return adapter.Wrap(middleware.CheckJWT)
}

func ValidatePermissions(log *zap.Logger, expectedClaims []string) gin.HandlerFunc {
	return func(c *gin.Context) {
		userPermissions := c.GetHeader("userPermissions")
		var userPermissionsDto []dto.UserPermission

		if err := json.Unmarshal([]byte(userPermissions), &userPermissionsDto); err != nil {
			log.Error("Failed to unmarshal the user permissions", zap.Error(err))
			c.AbortWithError(http.StatusInternalServerError, err)
			return
		}

		foundPermission := make([]bool, len(expectedClaims))
		for i := range expectedClaims {
			foundPermission[i] = false
		}

		for _, permission := range userPermissionsDto {
			for i := range expectedClaims {
				if permission.PermissionName == expectedClaims[i] {
					foundPermission[i] = true
				}
			}
		}

		ok := true

		for _, permission := range foundPermission {
			if !permission {
				ok = false
				break
			}
		}

		if !ok {
			log.Error("User does not have the required permissions", zap.Any("expectedClaims", expectedClaims))
			c.AbortWithError(http.StatusUnauthorized, errors.New("User does not have the required permissions"))
			return
		}

		c.Next()
	}
}

func UserInfoMiddleware(
	ctx context.Context,
	redisClient *redis.Client,
	log *zap.Logger) gin.HandlerFunc {
	mgmtAudience := utils.GetDotEnvVariable("MGMT_AUTH0_AUDIENCE")
	mgmtClientId := utils.GetDotEnvVariable("MGMT_AUTH0_CLIENT_ID")
	mgmtClientSecret := utils.GetDotEnvVariable("MGMT_AUTH0_CLIENT_SECRET")
	authority := utils.GetDotEnvVariable("AUTH0_AUTHORITY")

	return func(c *gin.Context) {

		mgmtAudienceApi, err := url.JoinPath(mgmtAudience, "/api/v2/")
		if err != nil {
			log.Fatal("Failed to join the mgmt audience api", zap.Error(err))
			c.AbortWithError(http.StatusInternalServerError, err)
			return
		}

		mgmtTokenUrl, err := url.JoinPath(mgmtAudience, "/oauth/token")
		if err != nil {
			log.Fatal("Failed to join the mgmt token url", zap.Error(err))
			c.AbortWithError(http.StatusInternalServerError, err)
			return
		}

		mgmtPermissionUrl, err := url.JoinPath(mgmtAudience, "/api/v2/users/")
		if err != nil {
			log.Fatal("Failed to join the mgmt permission url", zap.Error(err))
			c.AbortWithError(http.StatusInternalServerError, err)
			return
		}

		r := c.Request
		token, err := jwtmiddleware.AuthHeaderTokenExtractor(r)
		if err != nil {
			log.Fatal("Failed to extract the token", zap.Error(err))
			c.AbortWithError(http.StatusInternalServerError, err)
		}

		// Fetch the user info
		issuerURL, err := url.Parse(authority)
		if err != nil {
			log.Fatal("Failed to parse the issuer url", zap.Error(err))
			c.AbortWithError(http.StatusInternalServerError, err)
		}

		userData, err := utils.GetUserInfo(issuerURL, token, log)
		if err != nil {
			log.Fatal("Failed to get the user info", zap.Error(err))
			c.AbortWithError(http.StatusInternalServerError, err)
			return
		}

		userDataMarshal, err := json.Marshal(userData)
		if err != nil {
			log.Fatal("Failed to marshal the user data", zap.Error(err))
			c.AbortWithError(http.StatusInternalServerError, err)
			return
		}

		// Fetch the user permissions
		access_token := ""

		cachedToken, err := redisClient.Get(ctx, "mgmtAccessToken").Result()

		if err == nil {
			access_token = cachedToken
			log.Info("Reading from cache")
		} else {
			log.Info("Calling auth0 to fetch mgmt access token")
			mgmtPostResponse, err := utils.GetMgmtPostResponse(mgmtClientId, mgmtClientSecret, mgmtAudienceApi, mgmtTokenUrl, log)
			if err != nil {
				log.Fatal("Failed to get the mgmt post response", zap.Error(err))
				c.AbortWithError(http.StatusInternalServerError, err)
				return
			}
			access_token = mgmtPostResponse.AccessToken
			redisClient.Set(ctx, "mgmtAccessToken", access_token, constants.TOKEN_EXPIRY_TIME)
		}

		sub := userData.Sub
		mgmtPermissionUrl, err = url.JoinPath(mgmtPermissionUrl, sub, "/permissions")
		if err != nil {
			log.Fatal("Failed to join the user info url", zap.Error(err))
			c.AbortWithError(http.StatusInternalServerError, err)
		}

		userPermissions, err := utils.GetUserPermissions(mgmtPermissionUrl, mgmtTokenUrl, mgmtClientId, mgmtClientSecret, mgmtAudienceApi, access_token, log)
		if err != nil {
			log.Fatal("Failed to get the user permissions", zap.Error(err))
			c.AbortWithError(http.StatusInternalServerError, err)
			return
		}

		userPermissionsMarshal, err := json.Marshal(userPermissions)
		if err != nil {
			log.Fatal("Failed to marshal the user permissions", zap.Error(err))
			c.AbortWithError(http.StatusInternalServerError, err)
			return
		}

		c.Request.Header.Add("userData", string(userDataMarshal))
		c.Request.Header.Add("userPermissions", string(userPermissionsMarshal))
		c.Next()
	}
}
