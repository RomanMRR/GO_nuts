package main

import (
	"io/ioutil"
	"log"

	"github.com/nats-io/stan.go"
)

func main() {

	file := []string{"model.json"}

	sc, err := stan.Connect("test-cluster", "publisher-id")

	if err != nil {
		log.Fatalln(err)
	}

	defer sc.Close()

	for _, file := range filenames {

		fContent, err := ioutil.ReadFile(filename)
		if err != nil {
			log.Fatalln(err)
		}
		sc.Publish("book", fContent)
	}
}
