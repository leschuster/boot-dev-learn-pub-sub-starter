package main

import (
	"fmt"
	"log"
	"strconv"
	"time"

	"github.com/bootdotdev/learn-pub-sub-starter/internal/gamelogic"
	"github.com/bootdotdev/learn-pub-sub-starter/internal/pubsub"
	"github.com/bootdotdev/learn-pub-sub-starter/internal/routing"
	amqp "github.com/rabbitmq/amqp091-go"
)

func main() {
	fmt.Println("Starting Peril client...")
	const rabbitConnString = "amqp://guest:guest@localhost:5672/"

	conn, err := amqp.Dial(rabbitConnString)
	if err != nil {
		log.Fatalf("Cannot open connection to RabbitMQ: %v", err)
	}
	defer conn.Close()
	fmt.Println("Connection to RabbitMQ successfully established.")

	channel, err := conn.Channel()
	if err != nil {
		log.Fatalf("Cannot open channel: %v", err)
	}

	username, err := gamelogic.ClientWelcome()
	if err != nil {
		log.Fatalf("Please provide a valid username\n")
	}

	gamestate := gamelogic.NewGameState(username)

	err = pubsub.SubscribeJSON(
		conn,
		routing.ExchangePerilDirect,
		routing.PauseKey+"."+username,
		routing.PauseKey,
		pubsub.QueueTransient,
		handlerPause(gamestate),
	)
	if err != nil {
		log.Fatalf("Could not subscribe to queue: %v\n", err)
	}

	err = pubsub.SubscribeJSON(
		conn,
		routing.ExchangePerilTopic,
		routing.ArmyMovesPrefix+"."+username,
		routing.ArmyMovesPrefix+".*",
		pubsub.QueueTransient,
		handlerMove(channel, gamestate),
	)
	if err != nil {
		log.Fatalf("Could not subscribe to queue: %v\n", err)
	}

	err = pubsub.SubscribeJSON(
		conn,
		routing.ExchangePerilTopic,
		routing.WarRecognitionsPrefix,
		routing.WarRecognitionsPrefix+".*",
		pubsub.QueueDurable,
		handlerWar(channel, gamestate),
	)
	if err != nil {
		log.Fatalf("Could not subscribe to queue: %v\n", err)
	}

loop0:
	for {
		input := gamelogic.GetInput()
		if len(input) == 0 {
			continue
		}

		switch input[0] {
		case "spawn":
			err := gamestate.CommandSpawn(input)
			if err != nil {
				fmt.Printf("Error: %v\n", err)
			}
		case "move":
			mv, err := gamestate.CommandMove(input)
			if err != nil {
				fmt.Printf("Error: %v\n", err)
				continue
			}

			err = pubsub.PublishJSON(
				channel,
				routing.ExchangePerilTopic,
				routing.ArmyMovesPrefix+"."+username,
				mv,
			)
			if err != nil {
				fmt.Printf("Error: %v\n", err)
				continue
			}
			fmt.Println("Move was published successfully.")
		case "status":
			gamestate.CommandStatus()
		case "help":
			gamelogic.PrintClientHelp()
		case "spam":
			if len(input) != 2 {
				fmt.Println("Please provide a number of spam messages: spam <number>")
				continue
			}
			n, err := strconv.Atoi(input[1])
			if err != nil {
				fmt.Println("Could not parse int. Please provide the following format: spam <number>")
				continue
			}
			for i := 0; i < n; i += 1 {
				msg := gamelogic.GetMaliciousLog()
				err = publishGameLog(channel, username, msg)
				if err != nil {
					fmt.Printf("error publishing malicious log: %s\n", err)
				}
			}
			fmt.Printf("Published %v malicious logs\n", n)
		case "quit":
			break loop0
		default:
			fmt.Println("Command not recognized. Please try again.")
		}
	}

	fmt.Println("Gracefully shutting down")
}

func publishGameLog(ch *amqp.Channel, username, msg string) error {
	return pubsub.PublishGob(
		ch,
		routing.ExchangePerilTopic,
		routing.GameLogSlug+"."+username,
		routing.GameLog{
			Username:    username,
			CurrentTime: time.Now(),
			Message:     msg,
		},
	)
}
