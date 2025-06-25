package auth_middleware

import (
	"io"
	"net/http"
	"net/url"
	"time"

	jwtmiddleware "github.com/auth0/go-jwt-middleware/v2"
	"github.com/auth0/go-jwt-middleware/v2/jwks"
	"github.com/auth0/go-jwt-middleware/v2/validator"
	"github.com/gin-gonic/gin"
	adapter "github.com/gwatts/gin-adapter"
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

func UserInfoMiddleware(log *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		r := c.Request
		token, err := jwtmiddleware.AuthHeaderTokenExtractor(r)
		if err != nil {
			log.Fatal("Failed to extract the token", zap.Error(err))
			c.AbortWithError(http.StatusInternalServerError, err)
		}

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
