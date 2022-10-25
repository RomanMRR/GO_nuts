package main

import (
	"fmt"
	"io/ioutil"
	"log"

	stan "github.com/nats-io/stan.go"
)

func main() {

	sc, _ := stan.Connect("prod", "simple-pub", stan.NatsURL(stan.DefaultNatsURL),
		stan.SetConnectionLostHandler(func(_ stan.Conn, err error) {
			log.Printf("Connection lost %v", err)
		}))
	defer sc.Close()

	files, err := ioutil.ReadDir("./models/")
	if err != nil {
		log.Fatal(err)
	}

	for _, file := range files {
		body, err := ioutil.ReadFile("./models/" + file.Name())
		if err != nil {
			log.Fatalf("unable to read file: %v", err)
		}
		sc.Publish("Order", body)
		fmt.Println("Sending completed successfully")
	}

}
