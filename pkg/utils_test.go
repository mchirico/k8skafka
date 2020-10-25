package pkg

import (
	"context"
	"fmt"
	"github.com/confluentinc/confluent-kafka-go/kafka"
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
		// This will fail after 100 milliseconds. The 5 second sleep
		// will be interrupted.
		fmt.Printf("Error starting compose: %v\n", err)
	}

}

func stopKafka() {
	ctx, cancel := context.WithTimeout(context.Background(), 9*time.Second)
	defer cancel()

	if err := exec.CommandContext(ctx, "bash", "-c", "cd ./../compose;docker-compose down").Run(); err != nil {

		fmt.Printf("Error stopping compose: %v\n", err)
	}

}

func TestCreateProduceConsumeDelete(t *testing.T) {

	topic := "topic0"
	broker := "localhost:29099"

	msgchan := make(chan kafka.Message, 1)
	errorchan := make(chan error, 1)

	msgs := []kafka.Message{}

	kt := NewKT(broker)

	err := kt.Create(topic, 1, 1)
	if err != nil {
		kt.Delete([]string{topic})
		t.Fatalf("Can't create topic: %s\n", topic)
	}

	for i := 0; i < 4; i++ {
		err = kt.Producer(topic, fmt.Sprintf("test:%d", i))
		if err != nil {
			t.Fatalf("Producer error")
		}
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5000*time.Millisecond)
	defer cancel()

	go kt.Consumer(ctx, topic, "group", msgchan, errorchan)

	run := true
	for run {
		select {
		case err := <-errorchan:
			run = false
			t.Fatalf("ErrorChan: %v", err)
		case res := <-msgchan:
			fmt.Println(res.Value)
			msgs = append(msgs, res)
		case <-time.After(13 * time.Second):
			t.Logf("timeout 13 seconds")
			run = false
		}
	}

	fmt.Println(msgs)
	for i, v := range msgs {
		t.Logf("msg: %d, %v\n", i, string(v.Value))
		expected := fmt.Sprintf("test:%d", i)
		if string(v.Value) != expected {
			t.Fatalf("Expected:%s Got:%s", expected, string(v.Value))
		}
	}
	if len(msgs) != 4 {
		t.Fatalf("Expected: %s, got: %d", "4 messages", len(msgs))
	}

	err = kt.Delete([]string{topic})
	if err != nil {
		t.Fatalf("Can't delete pkg: %s\n", topic)
	}

}

func CreateKTandTopic(topic, broker string) (*KT, error) {

	kt := NewKT(broker)

	err := kt.Create(topic, 1, 1)
	if err != nil {
		kt.Delete([]string{topic})
		return kt, err
	}
	return kt, nil
}

func ReadT(t *testing.T, topic, broker, group string) []kafka.Message {
	msgs := []kafka.Message{}
	msgchan := make(chan kafka.Message, 1)
	errorchan := make(chan error, 1)

	kt := NewKT(broker)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	go kt.Consumer(ctx, topic, group, msgchan, errorchan)

	run := true
	for run {
		select {
		case err := <-errorchan:
			run = false
			t.Fatalf("ErrorChan: %v", err)
		case res := <-msgchan:
			fmt.Println(res.Value)
			msgs = append(msgs, res)
		case <-time.After(7 * time.Second):
			t.Logf("timeout 13 seconds")
			run = false
		}
	}

	fmt.Println(msgs)
	for i, v := range msgs {
		t.Logf("***msg: %d, %v\n", i, string(v.Value))
	}
	return msgs

}

func TestKT_ProducerIdem(t *testing.T) {
	topic := "topic1"
	broker := "localhost:29099"

	result := []string{}

	kt, err := CreateKTandTopic(topic, broker)
	if err != nil {
		kt.Delete([]string{topic})
		t.Fatalf("Can't create topic: %s\n", topic)
	}

	// Setup
	ctx, cancel := context.WithTimeout(context.Background(), 5000*time.Millisecond)
	defer cancel()

	msgByte := make(chan []byte)
	statusKafka := make(chan *kafka.Message)
	term := make(chan string)
	go kt.ProducerIdem(ctx, topic, msgByte, statusKafka, term)

	// This sends messages
	go func() {
		s := [][]byte{[]byte("one"), []byte("two"),
			[]byte("three"), []byte("four")}
		for _, v := range s {
			msgByte <- v
			time.Sleep(1000 * time.Millisecond)
		}
		close(msgByte)
	}()

	// Gets status of messages
	run := true
	for run {
		select {
		case val := <-statusKafka:
			fmt.Println("statusKafka:", string(val.Value))
			result = append(result, string(val.Value))
		case val := <-term:
			fmt.Println("done: ", val)
			run = false
		default:
		}
	}

	// Kafka preserves the order of messages within a partition.
	msgs := ReadT(t, topic, broker, "group")
	if string(msgs[3].Value) != "four" {
		t.Fatalf("Expected: %s, got: ->%s<-", "four", msgs[3].Value)
	}

	// Close down topic
	err = kt.Delete([]string{topic})
	if err != nil {
		t.Fatalf("Can't delete pkg: %s\n", topic)
	}

}
