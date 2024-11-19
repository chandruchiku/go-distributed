package event

import (
	"bytes"
	"encoding/json"
	"errors"
	"log"
	"net/http"

	amqp "github.com/rabbitmq/amqp091-go"
)

type Consumer struct {
	conn      *amqp.Connection
	queueName string
}

func NewConsumer(conn *amqp.Connection) (*Consumer, error) {
	consumer := &Consumer{
		conn: conn,
	}

	err := consumer.setup()
	if err != nil {
		return nil, err
	}

	return consumer, nil
}

func (c *Consumer) setup() error {
	// create a channel
	ch, err := c.conn.Channel()
	if err != nil {
		return err
	}

	return declareExchange(ch)
}

type Payload struct {
	Name string `json:"name"`
	Data string `json:"data"`
}

func (c *Consumer) Listen(topics []string) error {
	ch, err := c.conn.Channel()
	if err != nil {
		return err
	}
	defer ch.Close()

	q, err := declareRandomQueue(ch)
	if err != nil {
		return err
	}

	for _, topic := range topics {
		if err := ch.QueueBind(q.Name, topic, "logs_topic", false, nil); err != nil {
			return err
		}
	}

	msgs, err := ch.Consume(q.Name, "", true, false, false, false, nil)
	if err != nil {
		return err
	}

	forever := make(chan bool)
	go func() {
		for d := range msgs {
			log.Printf("Received a message: %s", d.Body)
			var payload Payload
			_ = json.Unmarshal(d.Body, &payload)

			go handlePayload(payload)
		}
	}()

	log.Printf(" [*] Waiting for logs [Exchange, Queue] = [%s, %s]", "logs_topic", q.Name)
	<-forever

	return nil
}

func handlePayload(payload Payload) {
	switch payload.Name {
	case "log", "event":
		err := logEvent(payload)
		if err != nil {
			log.Printf("Error logging event: %s", err)
		}
	case "mail":
		log.Printf("Mail message: %s", payload.Data)
	default:
		err := logEvent(payload)
		if err != nil {
			log.Printf("Error logging event: %s", err)
		}
	}
}

func logEvent(entry Payload) error {
	jsonData, _ := json.MarshalIndent(entry, "", "\t")
	log.Printf("Logging: %s", jsonData)

	request, err := http.NewRequest("POST", "http://logger-service:8080/log", bytes.NewBuffer(jsonData))
	if err != nil {
		return err
	}

	request.Header.Set("Content-Type", "application/json")
	client := &http.Client{}
	response, err := client.Do(request)
	if err != nil {
		return err
	}

	defer response.Body.Close()

	if response.StatusCode != http.StatusAccepted {
		return errors.New("error logging")
	}

	return nil
}
