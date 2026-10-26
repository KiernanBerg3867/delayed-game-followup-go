package main

import "encoding/json"

type Asset struct {
	PlayerID  string
	AssetID   string
	Published bool
}
type LiveEvent struct {
	EventID  string
	AssetID  string
	FollowUp string
}

func scheduleFollowUp(client *Client, asset Asset, event LiveEvent) error {
	if !asset.Published || asset.PlayerID == "" || event.AssetID != asset.AssetID {
		return nil
	}
	return client.Publish("game-followups", event)
}

func runFollowUp(client *Client, msg message) error {
	var event LiveEvent
	if err := json.Unmarshal(msg.Payload, &event); err != nil {
		return client.CaptureFailure(map[string]any{"exception": err.Error()})
	}
	if event.FollowUp == "" {
		return client.CaptureFailure(map[string]any{"exception": "follow-up action is empty"})
	}
	return client.Ack("game-followups", msg.MessageID)
}
