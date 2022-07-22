package main

import (
	"fmt"
	"log"

	amqp "github.com/rabbitmq/amqp091-go"
)

var SYMBOLS = [...]string{"BTCUSD", "ETHUSD", "XRPUSD"}

func failOnError(err error, msg string) {
	if err != nil {
		log.Panicf("%s: %s", msg, err)
	}
}

func initForSymbol(symbol string) {
	conn, err := amqp.Dial("amqp://guest:guest@localhost:5672/")
	failOnError(err, "Failed to connect to RabbitMQ")
	defer conn.Close()

	ch, err := conn.Channel()
	failOnError(err, "Failed to open a channel")
	defer ch.Close()

	q, err := ch.QueueDeclare(
		symbol, // name
		false,  // durable
		false,  // delete when unused
		false,  // exclusive
		false,  // no-wait
		nil,    // arguments
	)
	failOnError(err, "Failed to declare a queue")

	for price := 50; price < 110; price += 10 {
		body := fmt.Sprintf("add %s limit bid 1 %d 1 1", symbol, price)
		err = ch.Publish(
			"",     // exchange
			q.Name, // routing key
			false,  // mandatory
			false,  // immediate
			amqp.Publishing{
				ContentType: "text/plain",
				Body:        []byte(body),
			})
		failOnError(err, "Failed to publish a message")

	}
	for price := 110; price < 160; price += 10 {
		body := fmt.Sprintf("add %s limit ask 1 %d 3 3", symbol, price)
		err = ch.Publish(
			"",
			q.Name,
			false, // mandatory
			false, // immediate
			amqp.Publishing{
				ContentType: "text/plain",
				Body:        []byte(body),
			})
		failOnError(err, "Failed to publish a message")

	}

	body := fmt.Sprintf("add %s limit ask 1 100 2 2", symbol)
	err = ch.Publish(
		"", // exchange

		q.Name, // routing key
		false,  // mandatory
		false,  // immediate
		amqp.Publishing{
			ContentType: "text/plain",
			Body:        []byte(body),
		})
	failOnError(err, "Failed to publish a message")
	fmt.Println(symbol, " Order Book initialized")
}

func main() {
	for _, symbol := range SYMBOLS {
		initForSymbol(symbol)
	}
}
