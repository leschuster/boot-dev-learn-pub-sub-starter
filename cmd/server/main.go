package main

import (
	"fmt"
	"log"
	"os"

	"github.com/bootdotdev/learn-pub-sub-starter/internal/gamelogic"
	"github.com/bootdotdev/learn-pub-sub-starter/internal/pubsub"
	"github.com/bootdotdev/learn-pub-sub-starter/internal/routing"
	amqp "github.com/rabbitmq/amqp091-go"
)

func main() {
	fmt.Println("Starting Peril server...")

	cred := "amqp://guest:guest@localhost:5672/"

	conn, err := amqp.Dial(cred)
	if err != nil {
		fmt.Printf("Cannot open connection to RabbitMQ\n")
		os.Exit(1)
	}
	defer conn.Close()

	fmt.Println("Connection to RabbitMQ successfully established.")

	channel, err := conn.Channel()
	if err != nil {
		log.Fatalf("Cannot open channel: %v", err)
	}

	gamelogic.PrintServerHelp()

	err = pubsub.SubscribeGob(
		conn,
		routing.ExchangePerilTopic,
		routing.GameLogSlug,
		routing.GameLogSlug+".*",
		pubsub.QueueDurable,
		handleLog,
	)
	if err != nil {
		log.Fatalf("Cannot subscribe to game logs: %v", err)
	}

loop:
	for {
		input := gamelogic.GetInput()
		if len(input) == 0 {
			continue
		}

		switch input[0] {
		case "pause":
			fmt.Println("Sending a pause message.")
			pubsub.PublishJSON(
				channel,
				routing.ExchangePerilDirect,
				routing.PauseKey,
				routing.PlayingState{IsPaused: true},
			)
		case "resume":
			fmt.Println("Sending a resume message.")
			pubsub.PublishJSON(
				channel,
				routing.ExchangePerilDirect,
				routing.PauseKey,
				routing.PlayingState{IsPaused: false},
			)
		case "quit":
			fmt.Println("Exiting application.")
			break loop
		default:
			fmt.Println("Command not recognized. Please try again.")
		}
	}

	fmt.Println("Gracefully shutting down")
}
