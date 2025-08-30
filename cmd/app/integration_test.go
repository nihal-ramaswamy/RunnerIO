package app_test

import (
	"context"
	"fmt"
	"log"
	"testing"

	"github.com/joho/godotenv"
	"github.com/testcontainers/testcontainers-go/modules/compose"
)

func TestInit(t *testing.T) {
	stack, err := compose.NewDockerCompose("../../docker-compose.yml")
	if err != nil {
		log.Printf("Failed to create stack: %v", err)
		return
	}

	godotenv.Load("../../.env.test")

	ctx := t.Context()

	err = stack.
		WithOsEnv().
		Up(ctx, compose.Wait(true))
	if err != nil {
		log.Printf("Failed to start stack: %v", err)
		t.Fail()
		return
	}
	defer func(ctx context.Context) {
		err = stack.Down(
			ctx,
			compose.RemoveOrphans(true),
			compose.RemoveVolumes(true),
			compose.RemoveImagesLocal,
		)
		if err != nil {
			log.Printf("Failed to stop stack: %v", err)
			t.Fail()
		}
	}(ctx)

	serviceNames := stack.Services()

	fmt.Println(serviceNames)
}
