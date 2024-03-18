package event

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	amqp "github.com/rabbitmq/amqp091-go"
)

type Consumser struct {
	conn      *amqp.Connection
	queueName string
}

func NewConsumer(conn *amqp.Connection) (Consumser, error) {
	consumer := Consumser{
		conn: conn,
	}

	err := consumer.setup()
	if err != nil {
		return Consumser{}, err
	}

	return consumer, nil

}

func (consumer *Consumser) setup() error {
	channel, err := consumer.conn.Channel()
	if err != nil {
		return err
	}
	return declareExchange(channel)
}

type Payload struct {
	Name string `json:"name"`
	Data string `json:"data"`
}

func (consumner *Consumser) Listen(topics []string) error {
	ch, err := consumner.conn.Channel()
	if err != nil {
		return err
	}
	defer ch.Close()

	q, err := declareRandomQueue(ch)
	if err != nil {
		return err
	}

	for _, s := range topics {
		ch.QueueBind(
			q.Name,
			s,
			"log_topics",
			false,
			nil,
		)
	}

	messages, err := ch.Consume(q.Name, "", true, false, false, false, nil)
	if err != nil {
		return err
	}

	forever := make(chan bool)
	go func() {
		for d := range messages {
			var payload Payload
			_ = json.Unmarshal(d.Body, &payload)

			go handlePayload(payload)
		}
	}()

	fmt.Printf("Waiting for message [Exchange, Queue] [log_topics, %s]\n", q.Name)
	<-forever

	return nil
}

func handlePayload(payload Payload) {
	switch payload.Name {
	case "log", "event":
		err := logEvent(payload)
		if err != nil {
			log.Println(err)
		}

	case "auth":
		// do something

	default:
		err := logEvent(payload)
		if err != nil {
			log.Println(err)
		}
	}
}

func logEvent(entry Payload) error {
	jsonData, _ := json.MarshalIndent(entry, "", "\t")

	logServiceUrl := "http://localhost:8040/log"

	request, err := http.NewRequest("POST", logServiceUrl, bytes.NewBuffer(jsonData))
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
		return err
	}

	return nil
}
