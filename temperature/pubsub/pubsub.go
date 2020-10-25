package pubsub

import (
	"context"
	"fmt"
	"github.com/confluentinc/confluent-kafka-go/kafka"
	"github.com/mchirico/goKafka/pkg/wrapper"
	"github.com/mchirico/goKafka/temperature/readings"
	"log"
	"time"
)

type Msg struct {
	hostname  string
	sleepTime int
}

func (m *Msg) Send(msgbyte chan []byte) {

	var count int64
	for {

		b, err := readings.GenerateTemperatureJSON(count, "test")
		if err != nil {
			log.Printf("Send JSON: %s\n", err)
			close(msgbyte)
		}
		msgbyte <- b
		count += 1
		time.Sleep(time.Duration(m.sleepTime) * time.Second)
	}
	close(msgbyte)

}

func Pub(timeOutSeconds int,broker string) error {
	topic := "celcius-readings"
	numparts := 1

	ps, err := wrapper.Create(topic, broker, numparts)
	if err != nil {
		return err
	}

	msg := &Msg{"pubTest", 1}

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(timeOutSeconds)*time.Second)
	defer cancel()

	statusKafka, term, err := ps.Write(ctx, msg)
	if err != nil {
		return err
	}

	// Gets status of messages
	run := true
	for run {
		select {
		case val := <-statusKafka:
			fmt.Println("status Pub celcius-readings:", string(val.Value))
		case val := <-term:
			fmt.Println("done: ", val)
			run = false
		default:
		}
	}

	return nil
}

func Sub(ctx context.Context,broker string, timeOutSeconds int) {
	topic := "celcius-readings"

	// Read
	msgchan := make(chan kafka.Message, 1)
	errorchan := make(chan error, 1)

	wrapper.Read(ctx, topic, broker, "group", msgchan, errorchan)

	msgResults := []string{}
	run := true
	for run {
		select {
		case err := <-errorchan:
			run = false
			log.Printf("ErrorChan: %v", err)
		case res := <-msgchan:
			result, err := readings.ConvertTempToStruct(res.Value)
			if err != nil {
				log.Printf("Sub read err: %s\n", err)
			}
			fmt.Printf("Message received\n"+
				"%v\n", result)
			msgResults = append(msgResults, string(res.Value))

		case <-time.After(time.Duration(timeOutSeconds) * time.Second):
			log.Printf("Timeout, case statement. End run.\n")
			run = false
		}
	}

}
