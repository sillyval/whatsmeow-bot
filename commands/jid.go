package commands

import (
    "fmt"

    "go.mau.fi/whatsmeow"
    "go.mau.fi/whatsmeow/types/events"
    "whatsmeow-bot/utils"
)

type JIDCommand struct{}

func init() {
    RegisterCommand(&JIDCommand{})
}

func (c *JIDCommand) Execute(client *whatsmeow.Client, message *events.Message, args []string) {
	currentJID := message.Info.Chat.String()
	utils.Reply(client, message, fmt.Sprintf("Current chat JID: %s", currentJID))
}

func (c *JIDCommand) Name() string {
    return "jid"
}

func (c *JIDCommand) Description() string {
    return "Get the JID of the current chat"
}