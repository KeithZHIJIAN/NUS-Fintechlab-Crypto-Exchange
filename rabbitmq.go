package main

import (
	"log"

	amqp "github.com/rabbitmq/amqp091-go"
)

var ch = NewChanel()

func NewChanel() *amqp.Channel {
	conn, err := amqp.Dial("amqp://guest:guest@localhost:5672/")
	failOnError(err, "Failed to connect to RabbitMQ")

	ch, err := conn.Channel()
	failOnError(err, "Failed to open a channel")
	return ch
}

func failOnError(err error, msg string) {
	if err != nil {
		log.Panicf("%s: %s", msg, err)
	}
}

func Listen(ob *OrderBook) {
	q, err := ch.QueueDeclare(
		ob.symbol, // name
		false,     // durable
		false,     // delete when unused
		false,     // exclusive
		false,     // no-wait
		nil,       // arguments
	)
	failOnError(err, "Failed to declare a queue")

	msgs, err := ch.Consume(
		q.Name, // queue
		"",     // consumer
		true,   // auto-ack
		false,  // exclusive
		false,  // no-local
		false,  // no-wait
		nil,    // args
	)
	failOnError(err, "Failed to register a consumer")

	var forever chan struct{}

	go func() {
		for d := range msgs {
			log.Printf("Received a message: %s", d.Body)
			ob.Apply(string(d.Body[:]))
		}
	}()

	log.Printf(" [*] %s waiting for messages. To exit press CTRL+C", ob.symbol)
	<-forever
}

func UpdateAskOrder(ob *OrderBook) {
	exchange := "UPDATE_ASK_ORDER_" + ob.symbol + ".DLQ.Exchange"

	err := ch.ExchangeDeclare(
		exchange, // name
		"fanout", // type
		false,    // durable
		false,    // auto-deleted
		false,    // internal
		false,    // no-wait
		nil,      // arguments
	)
	failOnError(err, "Failed to declare an exchange")

	err = ch.Publish(
		exchange, // exchange
		"",       // routing key
		false,    // mandatory
		false,    // immediate
		amqp.Publishing{
			ContentType: "text/plain",
			Body:        []byte(ob.asks.UpdateString()),
		})
	failOnError(err, "Failed to publish a message")
}

func UpdateBidOrder(ob *OrderBook) {
	exchange := "UPDATE_BID_ORDER_" + ob.symbol + ".DLQ.Exchange"

	err := ch.ExchangeDeclare(
		exchange, // name
		"fanout", // type
		false,    // durable
		false,    // auto-deleted
		false,    // internal
		false,    // no-wait
		nil,      // arguments
	)
	failOnError(err, "Failed to declare an exchange")

	err = ch.Publish(
		exchange, // exchange
		"",       // routing key
		false,    // mandatory
		false,    // immediate
		amqp.Publishing{
			ContentType: "text/plain",
			Body:        []byte(ob.bids.UpdateString()),
		})
	failOnError(err, "Failed to publish a message")
}
