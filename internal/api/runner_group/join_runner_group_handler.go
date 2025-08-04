package runner_api

import (
	"context"
	"encoding/json"
	"net/http"
	"slices"

	"github.com/gin-gonic/gin"
	"github.com/nihal-ramaswamy/RunnerIO/internal/constants"
	"github.com/nihal-ramaswamy/RunnerIO/internal/dto"
	mongo_schema "github.com/nihal-ramaswamy/RunnerIO/internal/dto/mongodb_schema"
	interfaces "github.com/nihal-ramaswamy/RunnerIO/internal/interface"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.uber.org/zap"
)

// @Summary Join Runner Group
// @Description Join Runner Group
// @Tags Runner Group
// @Produce json
// @Success 200 {object} {message: "Joined group successfully"}
// @Success 200 {object} {message: "You are already a member of this group"}
// @Failure 200 {object} {message: "Group is full"}
// @Router /runner/joinRunnerGroup/{groupCode} [get]
type JoinRunnerGroupHandler struct {
	ctx         context.Context
	mongoClient *mongo.Client
	log         *zap.Logger

	interfaces.HandlerInterface
}

func NewJoinRunnerGroupHandler(ctx context.Context, mongoClient *mongo.Client, log *zap.Logger) *JoinRunnerGroupHandler {
	return &JoinRunnerGroupHandler{
		mongoClient: mongoClient,
		log:         log,
		ctx:         ctx,
	}
}

func (*JoinRunnerGroupHandler) Pattern() string {
	return "/joinRunnerGroup/:groupCode"
}

func (h *JoinRunnerGroupHandler) Handler() gin.HandlerFunc {
	return func(c *gin.Context) {

		code := c.Param("groupCode")

		userDataStr := c.GetHeader("userData")
		var userData dto.UserData
		if err := json.Unmarshal([]byte(userDataStr), &userData); err != nil {
			c.AbortWithError(http.StatusInternalServerError, err)
			return
		}

		collection := constants.RUNNER_GROUP_COLLECTION

		filter := bson.M{"code": code}
		var result mongo_schema.RunnerGroupSchema
		err := h.mongoClient.Database(constants.RUNNER_DATABASE).Collection(collection).FindOne(h.ctx, filter).Decode(&result)
		if err != nil {
			c.AbortWithError(http.StatusInternalServerError, err)
			return
		}

		if slices.Contains(result.Members, userData.Sub) {
			c.JSON(http.StatusOK, gin.H{
				"message": "You are already a member of this group",
			})
			return
		}

		if len(result.Members) == constants.MAX_GROUP_MEMBERS {
			c.JSON(http.StatusOK, gin.H{
				"message": "Group is full",
			})
			return
		}

		result.Members = append(result.Members, userData.Sub)

		_, err = h.mongoClient.Database(constants.RUNNER_DATABASE).Collection(collection).UpdateOne(
			h.ctx,
			filter,
			bson.M{"$set": bson.M{"members": result.Members}},
		)

		if err != nil {
			c.AbortWithError(http.StatusInternalServerError, err)
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"message": "Joined group successfully",
		})
	}
}

func (*JoinRunnerGroupHandler) RequestMethod() string {
	return http.MethodGet
}

func (h *JoinRunnerGroupHandler) Middlewares() []gin.HandlerFunc {
	return []gin.HandlerFunc{}
}
