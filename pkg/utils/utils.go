package utils

import (
	"context"
	"fmt"
	"github.com/confluentinc/confluent-kafka-go/kafka"
	"os"
	"os/signal"
	"syscall"
	"time"
)

type KT struct {
	Broker string
}

func NewKT(broker string) *KT {
	kt := &KT{Broker: broker}
	return kt
}

/*

Create("localhost:29092","topic0",4,1)
*/
func (kt *KT) Create(topic string, numParts, replicationFactor int) error {

	// Create a new AdminClient.
	// AdminClient can also be instantiated using an existing
	// Producer or Consumer instance, see NewAdminClientFromProducer and
	// NewAdminClientFromConsumer.
	a, err := kafka.NewAdminClient(&kafka.ConfigMap{"bootstrap.servers": kt.Broker})
	if err != nil {
		fmt.Printf("Failed to create Admin client: %s\n", err)
		return err
	}

	// Contexts are used to abort or limit the amount of time
	// the Admin call blocks waiting for a result.
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Create topics on cluster.
	// Set Admin options to wait for the operation to finish (or at most 60s)
	maxDur, err := time.ParseDuration("60s")
	if err != nil {
		return err
	}
	results, err := a.CreateTopics(
		ctx,
		// Multiple topics can be created simultaneously
		// by providing more TopicSpecification structs here.
		[]kafka.TopicSpecification{{
			Topic:             topic,
			NumPartitions:     numParts,
			ReplicationFactor: replicationFactor}},
		// Admin options
		kafka.SetAdminOperationTimeout(maxDur))
	if err != nil {
		fmt.Printf("Failed to create pkg: %v\n", err)
		return err
	}

	// Print results
	for _, result := range results {
		fmt.Printf("%s\n", result)
	}

	a.Close()
	return nil
}

func (kt *KT) Delete(topics []string) error {

	// Create a new AdminClient.
	// AdminClient can also be instantiated using an existing
	// Producer or Consumer instance, see NewAdminClientFromProducer and
	// NewAdminClientFromConsumer.
	a, err := kafka.NewAdminClient(&kafka.ConfigMap{"bootstrap.servers": kt.Broker})
	if err != nil {
		fmt.Printf("Failed to create Admin client: %s\n", err)
		return err
	}

	// Contexts are used to abort or limit the amount of time
	// the Admin call blocks waiting for a result.
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Delete topics on cluster
	// Set Admin options to wait for the operation to finish (or at most 60s)
	maxDur, err := time.ParseDuration("60s")
	if err != nil {
		return err
	}

	results, err := a.DeleteTopics(ctx, topics, kafka.SetAdminOperationTimeout(maxDur))
	if err != nil {
		fmt.Printf("Failed to delete topics: %v\n", err)
		return err
	}

	// Print results
	for _, result := range results {
		fmt.Printf("%s\n", result)
	}

	a.Close()
	return nil
}

func (kt *KT) Producer(topic string, msg string) error {

	p, err := kafka.NewProducer(&kafka.ConfigMap{"bootstrap.servers": kt.Broker})

	if err != nil {
		fmt.Printf("Failed to create producer: %s\n", err)
		return err
	}

	fmt.Printf("Created Producer %v\n", p)

	doneChan := make(chan bool)

	go func() {
		defer close(doneChan)
		for e := range p.Events() {
			switch ev := e.(type) {
			case *kafka.Message:
				m := ev
				if m.TopicPartition.Error != nil {
					fmt.Printf("Delivery failed: %v\n", m.TopicPartition.Error)
				} else {
					fmt.Printf("Delivered message to pkg %s [%d] at offset %v\n",
						*m.TopicPartition.Topic, m.TopicPartition.Partition, m.TopicPartition.Offset)
				}
				return

			default:
				fmt.Printf("Ignored event: %s\n", ev)
			}
		}
	}()

	p.ProduceChannel() <- &kafka.Message{TopicPartition: kafka.TopicPartition{Topic: &topic, Partition: kafka.PartitionAny}, Value: []byte(msg)}

	// wait for delivery report goroutine to finish
	_ = <-doneChan

	p.Close()
	return nil

}

func (kt *KT) Consumer(ctx context.Context, topic, group string,
	msgchan chan kafka.Message, errorchan chan error) {
	broker := kt.Broker
	//topics := []string{topic}

	sigchan := make(chan os.Signal, 1)
	signal.Notify(sigchan, syscall.SIGINT, syscall.SIGTERM)

	c, err := kafka.NewConsumer(&kafka.ConfigMap{
		"bootstrap.servers":               broker,
		"group.id":                        group,
		"session.timeout.ms":              6000,
		"go.events.channel.enable":        true,
		"go.application.rebalance.enable": true,
		// Enable generation of PartitionEOF when the
		// end of a partition is reached.
		"enable.partition.eof": true,
		"auto.offset.reset":    "earliest"})

	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to create consumer: %s\n", err)
		errorchan <- err
		return
	}

	fmt.Printf("Created Consumer %v\n", c)

	// err = c.SubscribeTopics([]string{topic, "^aRegex.*[Tt]opic"}, nil)
	err = c.SubscribeTopics([]string{topic}, nil)


	if err != nil {
		fmt.Printf("ERROR SubTopic")
		errorchan <- err
		return
	}

	run := true
	for run == true {
		select {

		case <-ctx.Done():
			fmt.Printf("Caught ctx\n")
			run = false

		case sig := <-sigchan:
			fmt.Printf("Caught signal %v: terminating\n", sig)
			run = false

		case ev := <-c.Events():
			switch e := ev.(type) {
			case kafka.AssignedPartitions:
				fmt.Fprintf(os.Stderr, "****** %% %v\n", e)
				c.Assign(e.Partitions)
			case kafka.RevokedPartitions:
				fmt.Fprintf(os.Stderr, "%% %v\n", e)
				c.Unassign()
			case *kafka.Message:
				msgchan <- *e
				//fmt.Printf("%% (c) Message on %s:\n%s\n",
				//	e.TopicPartition, string(e.Value))
			case kafka.PartitionEOF:
				fmt.Printf("%% Reached %v\n", e)
			case kafka.Error:
				// Errors should generally be considered as informational, the client will try to automatically recover
				fmt.Fprintf(os.Stderr, "%% Error: %v\n", e)
			}
		}
	}

	fmt.Printf("Closing consumer\n")
	c.Close()

}

func (kt *KT) ProducerIdem(ctx context.Context, topic string, c <-chan []byte,
	status chan *kafka.Message, term chan<- string) {

	broker := kt.Broker
	run := true

	p, err := kafka.NewProducer(&kafka.ConfigMap{
		"bootstrap.servers": broker,
		// Enable the Idempotent Producer
		"enable.idempotence": true})

	if err != nil {
		fmt.Printf("Failed to create producer: %s\n", err)
		os.Exit(1)
	}

	// For signalling termination from main to go-routine
	termChan := make(chan bool, 1)
	// For signalling that termination is done from go-routine to main
	doneChan := make(chan bool)

	// Go routine for serving the events channel for delivery reports and error events.
	go func() {
		doTerm := false
		for !doTerm {
			select {
			case e := <-p.Events():
				switch ev := e.(type) {
				case *kafka.Message:
					// Message delivery report
					m := ev
					status <- ev
					if m.TopicPartition.Error != nil {
						fmt.Printf("Delivery failed: %v\n", m.TopicPartition.Error)
					} else {
						fmt.Printf("Delivered message to pkg %s [%d] at offset %v\n",
							*m.TopicPartition.Topic, m.TopicPartition.Partition, m.TopicPartition.Offset)

					}

				case kafka.Error:
					// Generic client instance-level errors, such as
					// broker connection failures, authentication issues, etc.
					//
					// These errors should generally be considered informational
					// as the underlying client will automatically try to
					// recover from any errors encountered, the application
					// does not need to take action on them.
					//
					// But with idempotence enabled, truly fatal errors can
					// be raised when the idempotence guarantees can't be
					// satisfied, these errors are identified by
					// `e.IsFatal()`.

					e := ev
					if e.IsFatal() {
						// Fatal error handling.
						//
						// When a fatal error is detected by the producer
						// instance, it will emit kafka.Error event (with
						// IsFatal()) set on the Events channel.
						//
						// Note:
						//   After a fatal error has been raised, any
						//   subsequent Produce*() calls will fail with
						//   the original error code.
						fmt.Printf("FATAL ERROR: %v: terminating\n", e)
						run = false
						term <- fmt.Sprintf("FATAL ERROR: %v: terminating\n", e)
					} else {
						fmt.Printf("Error: %v\n", e)
					}

				default:
					fmt.Printf("Ignored event: %s\n", ev)
				}

			case <-termChan:
				doTerm = true
			}
		}

		close(doneChan)
	}()

	msgcnt := 0

	for run == true {

		select {
		case <-ctx.Done():
			fmt.Printf("Caught ctx\n")
			run = false
			term <- fmt.Sprintf("ctx.Done()")
			continue
		default:

		}

		// Produce message.
		// This is an asynchronous call, on success it will only
		// enqueue the message on the internal producer queue.
		// The actual delivery attempts to the broker are handled
		// by background threads.
		// Per-message delivery reports are emitted on the Events() channel,
		// see the go-routine above.

		v, ok := <-c
		if ok != true {
			run = false
			term <- fmt.Sprintf("c channel closed")
			continue
		}

		err = p.Produce(&kafka.Message{
			TopicPartition: kafka.TopicPartition{Topic: &topic, Partition: kafka.PartitionAny},
			Value:          v,
		}, nil)

		if err != nil {
			fmt.Printf("Failed to produce message: %v\n", err)
		}

		msgcnt++

		// Since fatal errors can't be triggered in practice,
		// use the test API to trigger a fabricated error after some time.
		//if msgcnt == 13 {
		//	p.TestFatalError(kafka.ErrOutOfOrderSequenceNumber, "Testing fatal errors")
		//}

		//time.Sleep(500 * time.Millisecond)

	}

	// Clean termination to get delivery results
	// for all outstanding/in-transit/queued messages.
	fmt.Printf("Flushing outstanding messages\n")
	p.Flush(15 * 1000)

	// signal termination to go-routine
	termChan <- true
	// wait for go-routine to terminate
	<-doneChan

	fatalErr := p.GetFatalError()

	p.Close()

	// Exit application with an error (1) if there was a fatal error.
	if fatalErr != nil {
		fmt.Printf("fatalErr in ProducerIdem")
	} else {
		//os.Exit(0)
	}
}
