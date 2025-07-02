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
		constants.AUDIT_QUEUE_NAME, // queue
		"",                         // consumer
		false,                      // auto-ack
		false,                      // exclusive
		false,                      // no-local
		false,                      // no-wait
		nil,                        // args
	)
	if err != nil {
		log.Fatal("Failed to register a consumer: %s", zap.Error(err))
		return
	}

	var forever chan struct{}
	var data dtoschema.RunnerAuditSchema

	go func() {
		for d := range msgs {
			log.Info("Received a message", zap.String("message", string(d.Body)))
			if err := json.Unmarshal(d.Body, &data); err != nil {
				log.Error("Failed to unmarshal message", zap.Error(err))
				continue
			}

			res, err := mongoClient.Database(constants.RUNNER_DATABASE).Collection(constants.RUNNER_AUDIT_COLLECTION).InsertOne(ctx, data)
			if err != nil {
				log.Error("Failed to insert document", zap.Error(err))
				continue
			}
			log.Info("Inserted document", zap.Any("id", res.InsertedID))
		}
	}()

	log.Info("Waiting for messages. To exit press CTRL+C")
	<-forever
}
