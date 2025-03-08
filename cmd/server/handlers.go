package main

import (
	"fmt"

	"github.com/bootdotdev/learn-pub-sub-starter/internal/gamelogic"
	"github.com/bootdotdev/learn-pub-sub-starter/internal/pubsub"
	"github.com/bootdotdev/learn-pub-sub-starter/internal/routing"
)

func handleLog(l routing.GameLog) pubsub.AckType {
	defer fmt.Print("> ")

	err := gamelogic.WriteLog(l)
	if err != nil {
		return pubsub.NackDiscard
	}

	return pubsub.Ack
}
