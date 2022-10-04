package utils

import (
	"fmt"
	"log"

	amqp "github.com/rabbitmq/amqp091-go"
)

var Ch = NewChanel()

func NewChanel() *amqp.Channel {
	conn, err := amqp.Dial("amqp://guest:guest@localhost:5672/")
	FailOnError(err, "Failed to connect to RabbitMQ")

	ch, err := conn.Channel()
	FailOnError(err, "Failed to open a channel")
	return ch
}

func FailOnError(err error, msg string) {
	if err != nil {
		log.Panicf("%s: %s", msg, err)
	}
}

func RabbitmqSend(exchange, msg string) {
	err := Ch.ExchangeDeclare(
		exchange, // name
		"fanout", // type
		false,    // durable
		false,    // auto-deleted
		false,    // internal
		false,    // no-wait
		nil,      // arguments
	)
	FailOnError(err, "Failed to declare an exchange")

	err = Ch.Publish(
		exchange, // exchange
		"",       // routing key
		false,    // mandatory
		false,    // immediate
		amqp.Publishing{
			ContentType: "text/plain",
			Body:        []byte(msg),
		})
	FailOnError(err, "Failed to publish a message")
}

func RabbitmqConsume(symbol string) <-chan amqp.Delivery {
	q, err := Ch.QueueDeclare(
		symbol, // name
		false,  // durable
		false,  // delete when unused
		false,  // exclusive
		false,  // no-wait
		nil,    // arguments
	)
	FailOnError(err, "Failed to declare a queue")

	msgs, err := Ch.Consume(
		q.Name, // queue
		"",     // consumer
		true,   // auto-ack
		false,  // exclusive
		false,  // no-local
		false,  // no-wait
		nil,    // args
	)
	FailOnError(err, "Failed to register a consumer")
	return msgs
}

func InitQueueForSymbol(symbol string) {
	conn, err := amqp.Dial("amqp://guest:guest@localhost:5672/")
	FailOnError(err, "Failed to connect to RabbitMQ")
	defer conn.Close()

	ch, err := conn.Channel()
	FailOnError(err, "Failed to open a channel")
	defer ch.Close()

	q, err := ch.QueueDeclare(
		symbol, // name
		false,  // durable
		false,  // delete when unused
		false,  // exclusive
		false,  // no-wait
		nil,    // arguments
	)
	FailOnError(err, "Failed to declare a queue")

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
		FailOnError(err, "Failed to publish a message")

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
		FailOnError(err, "Failed to publish a message")

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
	FailOnError(err, "Failed to publish a message")
	fmt.Println(symbol, " Order Book initialized")
}
