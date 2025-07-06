package messages

import (
	"context"
	"encoding/json"

	amqpconfig "github.com/nihal-ramaswamy/RunnerIO/internal/config/amqp"
	"github.com/nihal-ramaswamy/RunnerIO/internal/constants"
	dtoschema "github.com/nihal-ramaswamy/RunnerIO/internal/dto/schema"
	"go.mongodb.org/mongo-driver/mongo"
	"go.uber.org/zap"
)

func PersistAuditData(
	ctx context.Context,
	amqpconfig *amqpconfig.AmqpConfig,
	mongoClient *mongo.Client, log *zap.Logger) {
	msgs, err := amqpconfig.Channel.Consume(
		constants.LIVE_LINES_QUEUE_NAME, // queue
		"",                              // consumer
		false,                           // auto-ack
		false,                           // exclusive
		false,                           // no-local
		false,                           // no-wait
		nil,                             // args
	)
	if err != nil {
		log.Fatal("Failed to register a consumer: %s", zap.Error(err))
		return
	}

	var forever chan struct{}
	var data dtoschema.RunnerLiveLinesSchema

	go func() {
		for d := range msgs {
			log.Info("Received a message", zap.String("message", string(d.Body)))
			if err := json.Unmarshal(d.Body, &data); err != nil {
				log.Error("Failed to unmarshal message", zap.Error(err))
				continue
			}

			// Persist to Live Lines Db
			go func() {
				res, err := mongoClient.Database(constants.RUNNER_DATABASE).Collection(constants.RUNNER_LIVE_LINES_COLLECTION).InsertOne(ctx, data)
				if err != nil {
					log.Error("Failed to insert document", zap.Error(err))
					return
				}
				log.Info("Inserted document", zap.Any("id", res.InsertedID))
			}()

			// Eat processor
			go func() {
				RunProcessorOnGroup(ctx, data.GroupCode, mongoClient, amqpconfig, log)
			}()

		}
	}()

	log.Info("Waiting for messages. To exit press CTRL+C")
	<-forever
}
