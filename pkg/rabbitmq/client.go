// Package rabbitmq implements means for communication with RabbitMQ.
package rabbitmq

import (
	"context"
	"os"
	"time"

	"github.com/eroshiva/cloudtalk/pkg/logger"
	amqp "github.com/rabbitmq/amqp091-go"
)

const (
	envAMQPURL       = "AMQP_URL"
	defaultAMQPURL   = "amqp://guest:guest@localhost:5672/"
	envQueueName     = "QUEUE_NAME"
	defaultQueueName = "review_events"
	defaultTimeout   = 5 * time.Second
)

var (
	queueName = os.Getenv(envQueueName)
	zlog      = logger.NewLogger("rabbitmq")
)

// Connect establishes connection with RabbitMQ.
func Connect() (<-chan amqp.Delivery, *amqp.Connection, *amqp.Channel, error) {
	// setting input session parameters
	amqpURL := os.Getenv(envAMQPURL)
	if amqpURL == "" {
		amqpURL = defaultAMQPURL
	}
	if queueName == "" {
		queueName = defaultQueueName
	}
	zlog.Info().Msgf("Connecting to RabbitMQ on %s with topic %s", amqpURL, queueName)

	// dialing to RabbitMQ
	conn, err := amqp.Dial(amqpURL) // connection would be gracefully closed in the main function
	if err != nil {
		zlog.Error().Err(err).Msgf("Failed to connect to RabbitMQ on %s", amqpURL)
		return nil, nil, nil, err
	}

	// establishing channel
	ch, err := conn.Channel()
	if err != nil {
		zlog.Error().Err(err).Msg("Failed to open a channel")
		return nil, nil, nil, err
	}

	// declaring/fetching the queue
	q, err := ch.QueueDeclare(
		queueName, // name
		false,     // durable
		false,     // delete when unused
		false,     // exclusive
		false,     // no-wait
		nil,       // arguments
	)
	if err != nil {
		zlog.Error().Err(err).Msgf("Failed to declare a queue %s", queueName)
		return nil, nil, nil, err
	}

	// starting to consume messages (in case we want to receive something)
	msgs, err := ch.Consume(
		q.Name, // queue
		"",     // consumer
		true,   // auto-ack
		false,  // exclusive
		false,  // no-local
		false,  // no-wait
		nil,    // args
	)
	if err != nil {
		zlog.Error().Err(err).Msgf("Failed to register a consumer for queue (%s)", q.Name)
		return nil, nil, nil, err
	}
	return msgs, conn, ch, nil
}

// CloseChannel closes the open channel with RabbitMQ.
func CloseChannel(ch *amqp.Channel) {
	zlog.Info().Msgf("Closing channel for RabbitMQ's queue (%s)", queueName)
	err := ch.Close()
	if err != nil {
		zlog.Error().Err(err).Msg("Failed to close channel")
	}
}

// CloseConnection shuts down connection with RabbitMQ message bus.
func CloseConnection(conn *amqp.Connection) {
	zlog.Info().Msgf("Closing connection for RabbitMQ on %s", conn.RemoteAddr())
	err := conn.Close()
	if err != nil {
		zlog.Error().Err(err).Msgf("Failed to close RabbitMQ connection")
	}
}

// PublishMessage publishes message to the RabbitMQ's channel.
// For the sake of simplicity, only simple text messages are transmitted.
func PublishMessage(ctx context.Context, ch *amqp.Channel, text string) error {
	zlog.Info().Msgf("Publishing message to RabbitMQ: '%s'", text)
	// send out message to RabbitMQ
	err := ch.PublishWithContext(ctx,
		"",        // exchange
		queueName, // routing key
		false,     // mandatory
		false,     // immediate
		amqp.Publishing{
			ContentType: "text/plain",
			Body:        []byte(text),
		})
	if err != nil {
		zlog.Error().Err(err).Msgf("Failed to publish a message")
		return err
	}
	return nil
}
