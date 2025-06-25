package auth_middleware

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/url"
	"time"

	jwtmiddleware "github.com/auth0/go-jwt-middleware/v2"
	"github.com/auth0/go-jwt-middleware/v2/jwks"
	"github.com/auth0/go-jwt-middleware/v2/validator"
	"github.com/gin-gonic/gin"
	adapter "github.com/gwatts/gin-adapter"
	"github.com/nihal-ramaswamy/RunnerIO/internal/dto"
	"github.com/nihal-ramaswamy/RunnerIO/internal/utils"
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
	mgmtClientId := utils.GetDotEnvVariable("MGMT_AUTH0_CLIENT_ID")
	mgmtClientSecret := utils.GetDotEnvVariable("MGMT_AUTH0_CLIENT_SECRET")
	mgmtAudience := utils.GetDotEnvVariable("MGMT_AUTH0_AUDIENCE")
	mgmtTokenUrl := utils.GetDotEnvVariable("MGMT_AUTH0_TOKEN_URL")
	mgmtPermissionUrl := utils.GetDotEnvVariable("MGMT_AUTH0_PERMISSIONS")

	return func(c *gin.Context) {
		userDataStr := c.GetHeader("userData")

		var userData dto.UserData
		if err := json.Unmarshal([]byte(userDataStr), &userData); err != nil {
			log.Error("Failed to unmarshal the user data", zap.Error(err))
			c.AbortWithError(http.StatusInternalServerError, err)
		}

		client := &http.Client{}
		method := http.MethodPost
		mgmtBody := dto.NewMgmtPostRequest(mgmtClientId, mgmtClientSecret, mgmtAudience)
		bbody, err := json.Marshal(mgmtBody)
		if err != nil {
			log.Fatal("Failed to marshal the body", zap.Error(err))
			c.AbortWithError(http.StatusInternalServerError, err)
		}

		req, err := http.NewRequest(method, mgmtTokenUrl, bytes.NewBuffer(bbody))
		req.Header.Add("Content-Type", "application/json")
		res, err := client.Do(req)

		if err != nil {
			log.Fatal("Failed to make the request", zap.Error(err))
			c.AbortWithError(http.StatusInternalServerError, err)
		}
		defer res.Body.Close()

		body, err := io.ReadAll(res.Body)
		if err != nil {
			log.Fatal("Failed to read the response body", zap.Error(err))
			c.AbortWithError(http.StatusInternalServerError, err)
		}
		log.Info("body", zap.String("body", string(body)))

		var mgmtPostResponse dto.MgmtPostResponse
		if err := json.Unmarshal([]byte(string(body)), &mgmtPostResponse); err != nil {
			log.Error("Failed to unmarshal the user data", zap.Error(err))
			c.AbortWithError(http.StatusInternalServerError, err)
		}

		sub := userData.Sub
		access_token := mgmtPostResponse.AccessToken
		mgmtPermissionUrl, err = url.JoinPath(mgmtPermissionUrl, sub, "/permissions")

		if err != nil {
			log.Fatal("Failed to join the user info url", zap.Error(err))
			c.AbortWithError(http.StatusInternalServerError, err)
		}

		log.Info("permissions url", zap.String("url", mgmtPermissionUrl))

		client = &http.Client{}
		method = http.MethodGet

		req, err = http.NewRequest(method, mgmtPermissionUrl, nil)

		if err != nil {
			log.Fatal("Failed to create the request", zap.Error(err))
			c.AbortWithError(http.StatusInternalServerError, err)
		}

		req.Header.Add("Accept", "application/json")
		req.Header.Add("authorization", "Bearer "+access_token)
		res, err = client.Do(req)

		if err != nil {
			log.Fatal("Failed to make the request", zap.Error(err))
			c.AbortWithError(http.StatusInternalServerError, err)
		}
		defer res.Body.Close()

		body, err = io.ReadAll(res.Body)
		if err != nil {
			log.Fatal("Failed to read the response body", zap.Error(err))
			c.AbortWithError(http.StatusInternalServerError, err)
		}

		log.Info("body", zap.Any("body", string(body)))

		c.Next()
	}
}

func UserInfoMiddleware(log *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		r := c.Request
		token, err := jwtmiddleware.AuthHeaderTokenExtractor(r)
		if err != nil {
			log.Fatal("Failed to extract the token", zap.Error(err))
			c.AbortWithError(http.StatusInternalServerError, err)
		}
		log.Info("token", zap.String("token", token))

		authority := utils.GetDotEnvVariable("AUTH0_AUTHORITY")

		issuerURL, err := url.Parse(authority)
		if err != nil {
			log.Fatal("Failed to parse the issuer url", zap.Error(err))
			c.AbortWithError(http.StatusInternalServerError, err)
		}

		userInfoUrl, err := url.JoinPath(issuerURL.String(), "/userinfo")
		if err != nil {
			log.Fatal("Failed to join the user info url", zap.Error(err))
			c.AbortWithError(http.StatusInternalServerError, err)
		}

		client := &http.Client{}
		method := http.MethodGet

		req, err := http.NewRequest(method, userInfoUrl, nil)

		if err != nil {
			log.Fatal("Failed to create the request", zap.Error(err))
			c.AbortWithError(http.StatusInternalServerError, err)
		}

		req.Header.Add("Accept", "application/json")
		req.Header.Add("authorization", "Bearer "+token)
		res, err := client.Do(req)

		if err != nil {
			log.Fatal("Failed to make the request", zap.Error(err))
			c.AbortWithError(http.StatusInternalServerError, err)
		}
		defer res.Body.Close()

		body, err := io.ReadAll(res.Body)
		if err != nil {
			log.Fatal("Failed to read the response body", zap.Error(err))
			c.AbortWithError(http.StatusInternalServerError, err)
		}

		c.Request.Header.Add("userData", string(body))
		c.Next()
	}
}
