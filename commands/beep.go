package commands

import (
    "go.mau.fi/whatsmeow"
    "go.mau.fi/whatsmeow/types/events"

    "whatsmeow-bot/utils"
)

func init() {
    RegisterCommand(&BeepCommand{})
}

type BeepCommand struct{}

func (c *BeepCommand) Execute(client *whatsmeow.Client, message *events.Message, args []string) {
    utils.Reply(client, message, "boop!")
}

func (c *BeepCommand) Name() string {
    return "beep"
}