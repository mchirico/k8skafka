package wrapper

import (
	"context"
	"github.com/confluentinc/confluent-kafka-go/kafka"
	"github.com/mchirico/goKafka/pkg"
)

type PS struct {
	Topic    string
	Broker   string
	NumParts int
	kt       *pkg.KT
}

type Sender interface {
	Send(chan []byte)
}

func Create(topic, broker string, numparts int) (*PS, error) {

	ps := &PS{topic, broker, numparts, nil}
	ps.kt = pkg.NewKT(ps.Broker)

	err := ps.kt.Create(ps.Topic, ps.NumParts, 1)
	if err != nil {
		return ps, err
	}
	return ps, nil
}

func (ps *PS) Write(ctx context.Context, s Sender) (chan *kafka.Message, chan string, error) {

	msgByte := make(chan []byte)
	statusKafka := make(chan *kafka.Message)
	term := make(chan string)
	go ps.kt.ProducerIdem(ctx, ps.Topic, msgByte, statusKafka, term)

	go s.Send(msgByte)
	return statusKafka, term, nil

}

func Read(ctx context.Context, topic, broker, group string,
	msgchan chan kafka.Message, errorchan chan error) {
	kt := pkg.NewKT(broker)
	go kt.Consumer(ctx, topic, group, msgchan, errorchan)

}

func (ps *PS) Delete() error {

	err := ps.kt.Delete([]string{ps.Topic})
	if err != nil {
		return err
	}

	return nil
}
