package runner_api

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/nihal-ramaswamy/RunnerIO/internal/constants"
	"github.com/nihal-ramaswamy/RunnerIO/internal/dto"
	mongo_schema "github.com/nihal-ramaswamy/RunnerIO/internal/dto/mongodb_schema"
	interfaces "github.com/nihal-ramaswamy/RunnerIO/internal/interface"
	"github.com/nihal-ramaswamy/RunnerIO/internal/utils"
	"go.mongodb.org/mongo-driver/mongo"
	"go.uber.org/zap"
)

// @Summary Create Runner Group
// @Description Create Runner Group
// @Tags Runner Group
// @Produce json
// @Success 200 {object} {message: "Group created successfully"}
// @Failure 400 {object} dto.ErrorResponse
// @Router /runner/createRunnerGroup/ [post]
// @Body {object} {group: string}
type CreateRunnerGroupHandler struct {
	ctx         context.Context
	mongoClient *mongo.Client
	log         *zap.Logger

	interfaces.HandlerInterface
}

func NewCreateRunnerGroupHandler(ctx context.Context, mongoClient *mongo.Client, log *zap.Logger) *CreateRunnerGroupHandler {
	return &CreateRunnerGroupHandler{
		mongoClient: mongoClient,
		log:         log,
		ctx:         ctx,
	}
}

func (*CreateRunnerGroupHandler) Pattern() string {
	return "/createRunnerGroup"
}

func (h *CreateRunnerGroupHandler) Handler() gin.HandlerFunc {
	return func(c *gin.Context) {
		var requestData struct {
			Group string `json:"group"`
		}

		userDataStr := c.GetHeader("userData")
		var userData dto.UserData
		if err := json.Unmarshal([]byte(userDataStr), &userData); err != nil {
			c.AbortWithError(http.StatusInternalServerError, err)
			return
		}

		if err := c.ShouldBindJSON(&requestData); err != nil {
			c.AbortWithError(http.StatusBadRequest, err)
			return
		}

		collection := constants.RUNNER_GROUP_COLLECTION
		name := requestData.Group
		members := []string{userData.Sub}
		timeStamp := time.Now().UnixMilli()
		code := utils.CreateCode(h.mongoClient, h.log)

		data := mongo_schema.RunnerGroupSchema{
			GroupName: name,
			GroupCode: code,
			CreatedAt: timeStamp,
			Members:   members,
			Owner:     userData.Sub,
		}

		_, err := h.mongoClient.Database(constants.RUNNER_DATABASE).Collection(collection).InsertOne(
			h.ctx,
			data,
		)

		if err != nil {
			c.AbortWithError(http.StatusInternalServerError, err)
			return
		}

		c.JSON(http.StatusCreated, gin.H{
			"message": "Group created successfully",
		})
	}
}

func (*CreateRunnerGroupHandler) RequestMethod() string {
	return http.MethodPost
}

func (h *CreateRunnerGroupHandler) Middlewares() []gin.HandlerFunc {
	return []gin.HandlerFunc{}
}
