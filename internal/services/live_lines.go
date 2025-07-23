package services

import (
	"context"
	"encoding/json"
	"sync"

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

	forever := make(chan struct{})
	var data dtoschema.LiveLinesData

	var wg sync.WaitGroup

	go func() {
		for d := range msgs {
			log.Info("Received a message", zap.String("message", string(d.Body)))
			if err := json.Unmarshal(d.Body, &data); err != nil {
				log.Error("Failed to unmarshal message", zap.Error(err))
				continue
			}

			wg.Add(1)
			// Persist to Live Lines Db
			go func() {
				defer wg.Done()
				res, err := mongoClient.Database(constants.RUNNER_DATABASE).Collection(constants.RUNNER_LIVE_LINES_COLLECTION).InsertOne(ctx, data)
				if err != nil {
					log.Error("Failed to insert document", zap.Error(err))
					return
				}
				log.Info("Inserted document", zap.Any("id", res.InsertedID))
			}()

			wg.Add(1)
			// Run Polygon processor
			go func() {
				defer wg.Done()
				RunProcessorOnGroup(ctx, data.GroupCode, mongoClient, amqpconfig, log)
			}()

			wg.Wait()
			err := d.Ack(false)
			if err != nil {
				log.Error("Failed to ack message", zap.Error(err))
			}

		}
	}()

	log.Info("Waiting for messages. To exit press CTRL+C")
	<-forever
}
