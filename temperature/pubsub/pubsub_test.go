package pubsub

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/exec"
	"testing"
	"time"
)

func TestMain(m *testing.M) {
	startKafka()
	code := m.Run()
	stopKafka()
	os.Exit(code)

}

func startKafka() {
	ctx, cancel := context.WithTimeout(context.Background(), 10000*time.Millisecond)
	defer cancel()

	if err := exec.CommandContext(ctx, "bash", "-c", "cd ./../../compose;docker-compose up -d").Run(); err != nil {
		fmt.Printf("Error starting compose: %v\n", err)
	}

}

func stopKafka() {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := exec.CommandContext(ctx, "bash", "-c", "cd ./../../compose;docker-compose down").Run(); err != nil {
		log.Printf("Error stopping compose: %v\n", err)
	}
	log.Printf("Kafka stopped wrapper\n")
}

func TestPub(t *testing.T) {
	timeOuts := 10

	broker :=  "localhost:29099"
	go Pub(timeOuts,broker)

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(timeOuts)*time.Second)
	defer cancel()
	Sub(ctx,broker, timeOuts)
}
