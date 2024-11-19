package event

import (
	"log"

	amqp "github.com/rabbitmq/amqp091-go"
)

type Emmitter struct {
	connection *amqp.Connection
}

func (e *Emmitter) setup() error {
	channel, err := e.connection.Channel()
	if err != nil {
		return err
	}

	defer channel.Close()

	if err := declareExchange(channel); err != nil {
		return err
	}

	return nil
}

func (e *Emmitter) Push(event string, severity string) error {
	channel, err := e.connection.Channel()
	if err != nil {
		return err
	}

	defer channel.Close()

	log.Println("Pushing event", event, "with severity", severity)

	if err := channel.Publish(
		"logs_topic", // exchange
		severity,     // routing key
		false,        // mandatory
		false,        // immediate
		amqp.Publishing{
			ContentType: "text/plain",
			Body:        []byte(event),
		},
	); err != nil {
		return err
	}

	return nil
}

func NewEventEmmitter(conn *amqp.Connection) (*Emmitter, error) {
	emmitter := &Emmitter{
		connection: conn,
	}

	if err := emmitter.setup(); err != nil {
		return nil, err
	}

	return emmitter, nil
}
