package wrapper

import (
	"context"
	"fmt"
	"github.com/confluentinc/confluent-kafka-go/kafka"
	"log"

	//"github.com/confluentinc/confluent-kafka-go/kafka"
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

	if err := exec.CommandContext(ctx, "bash", "-c", "cd ./../compose;docker-compose up -d").Run(); err != nil {
		fmt.Printf("Error starting compose: %v\n", err)
	}

}

func stopKafka() {
	ctx, cancel := context.WithTimeout(context.Background(), 5000*time.Millisecond)
	defer cancel()

	if err := exec.CommandContext(ctx, "bash", "-c", "cd ./../compose;docker-compose down").Run(); err != nil {
		log.Printf("Error stopping compose: %v\n", err)
	}

}

type Msg struct {
	sleepTime int
}

func (m *Msg) Send(msgbyte chan []byte) {

	s := [][]byte{[]byte("test:0"), []byte("test:1"),
		[]byte("test:2"), []byte("test:3")}
	for _, v := range s {
		msgbyte <- v
		time.Sleep(time.Duration(m.sleepTime) * time.Millisecond)
	}
	close(msgbyte)

}

func TestPS_Write(t *testing.T) {
	topic := "topic2"
	broker := "localhost:29099"
	numparts := 1

	ps, err := Create(topic, broker, numparts)
	if err != nil {
		t.Fatalf("Error Create: %v\n", err)
	}

	msg := &Msg{500}

	ctx, cancel := context.WithTimeout(context.Background(), 5000*time.Millisecond)
	defer cancel()

	statusKafka, term, err := ps.Write(ctx, msg)
	if err != nil {
		t.Fatalf("ps.Write error: %s\n", err.Error())
	}

	// Gets status of messages
	run := true
	for run {
		select {
		case val := <-statusKafka:
			fmt.Println("statusKafka:", string(val.Value))
		case val := <-term:
			fmt.Println("done: ", val)
			run = false
		default:
		}
	}

	// Read
	msgchan := make(chan kafka.Message, 1)
	errorchan := make(chan error, 1)
	ctx2, cancel2 := context.WithTimeout(context.Background(), 50000*time.Millisecond)
	defer cancel2()
	Read(ctx2, topic, broker, "group", msgchan, errorchan)

	msgResults := []string{}
	run2 := true
	count := 0
	for run2 {
		select {
		case err := <-errorchan:
			run2 = false
			t.Fatalf("ErrorChan: %v", err)
		case res := <-msgchan:
			fmt.Println(res.Value)
			msgResults = append(msgResults, string(res.Value))

			// Quick exit
			count += 1
			if count >= 3 {
				run2 = false
			}
		case <-time.After(12 * time.Second):
			t.Logf("timeout 12 seconds")
			run2 = false
		}
	}

	for i, v := range msgResults {
		expected := fmt.Sprintf("test:%d", i)
		if v != expected {
			t.Fatalf("Expected: %s, Got: %s", expected, v)
		}
	}

	err = ps.Delete()
	if err != nil {
		t.Fatalf("err: %v\n", err)
	}
}
