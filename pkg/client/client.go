package client

import (
	"context"
	"math/rand"
	"time"
    "encoding/json"

	amqp "github.com/rabbitmq/amqp091-go"
)

type ChatAppClient struct{
    conn *amqp.Connection
    queue amqp.Queue
}

type ChatMessage struct {
    Message  string `json:"message"`
    Username string `json:"username"`
}

type ChatResponse struct {
    Message  string `json:"message"`
    Username string `json:"username"`
}


func randomString(l int) string {
	bytes := make([]byte, l)
	for i := 0; i < l; i++ {
		bytes[i] = byte(randInt(65, 90))
	}
	return string(bytes)
}

func randInt(min int, max int) int {
	return min + rand.Intn(max-min)
}

func (client ChatAppClient) SendMessageRPC(message ChatMessage) (err error) {
	ch, err := client.conn.Channel()
    if err != nil {
        return err
    }
    defer ch.Close()
    msg, err := json.Marshal(message)
    if err != nil {
        return err
    }

	corrId := randomString(32)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err = ch.PublishWithContext(ctx,
		"",          // exchange
		"rpc_queue", // routing key
		true,        // mandatory
		false,       // immediate
		amqp.Publishing{
			ContentType:   "text/plain",
			CorrelationId: corrId,
			ReplyTo:       client.queue.Name,
			Body:          msg,
		})
    return err
}

func (client ChatAppClient) ListenForMessage() (ChatResponse, error) {
	ch, err := client.conn.Channel()
    if err != nil {
        return ChatResponse{}, err
    }
	defer ch.Close()

    msgs, err := ch.Consume(
		client.queue.Name, // queue
		"",     // consumer
		true,   // auto-ack
		false,  // exclusive
		false,  // no-local
		false,  // no-wait
		nil,    // args
	)
    if err != nil {
        return ChatResponse{}, err
    }

	for d := range msgs {
        resp := ChatResponse{}
        err = json.Unmarshal(d.Body, &resp)
        if err != nil {
            return ChatResponse{}, err
        }
        return resp, err
	}
    return ChatResponse{Username: "BOT", Message: "FROM ANOTHER PACKAGE!!!"}, err
}

func NewClient() (ChatAppClient, error) {
    client := ChatAppClient{}
    conn, err := amqp.Dial("amqp://user:password@localhost:5672/")
    if err != nil {
        return ChatAppClient{}, err
    }
    client.conn = conn

	ch, err := conn.Channel()
    if err != nil {
        conn.Close()
        return ChatAppClient{}, err
    }

	q, err := ch.QueueDeclare(
		"",    // name
		false, // durable
		false,  // delete when unused
		false,  // exclusive
		false, // noWait
		nil,   // arguments
	)
    if err != nil {
        ch.Close()
        conn.Close()
        return ChatAppClient{}, err
    }
    client.queue = q

    err = ch.QueueBind(
        q.Name, // queue name
        "",     // routing key
        "chat_app", // exchange name
        false,  // no-wait
        nil,    // arguments
    )
    if err != nil {
        ch.Close()
        conn.Close()
        return ChatAppClient{}, err
    }

    return client, nil
}

func (client ChatAppClient) Close() error {
    return client.conn.Close()
}

