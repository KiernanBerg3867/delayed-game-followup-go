package main

import (
	"fmt"
	"log"
)

func main() {
	client, err := NewClient()
	if err != nil {
		log.Fatal(err)
	}
	if err := client.CreateQueue("game-followups-dlq", "", 0); err != nil {
		log.Fatal(err)
	}
	if err := client.CreateQueue("game-followups", "game-followups-dlq", 3); err != nil {
		log.Fatal(err)
	}
	asset := Asset{PlayerID: "player-17", AssetID: "skin-204", Published: true}
	event := LiveEvent{EventID: "event-204", AssetID: asset.AssetID, FollowUp: "grant-review-reward"}
	if err := scheduleFollowUp(client, asset, event); err != nil {
		log.Fatal(err)
	}
	items, err := client.Consume("game-followups")
	if err != nil {
		log.Fatal(err)
	}
	for _, item := range items {
		if err := runFollowUp(client, item); err != nil {
			log.Fatal(err)
		}
	}
	fmt.Println("follow-up queued and processed")
}
