package main

import "github.com/mchirico/goKafka/pkg"

func main() {
	topic := "topic0"
	broker := "localhost:29099"

	kt := pkg.NewKT(broker)

	err := kt.Create(topic, 4, 1)
	if err != nil {
		return
	}
	err = kt.Delete([]string{topic})
	if err != nil {
		return
	}

}
