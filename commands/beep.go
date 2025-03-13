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

func (c *BeepCommand) Execute(client *whatsmeow.Client, message *events.Message, args []string) *string {
    utils.Reply(client, message, "boop!")
    return nil
}

func (c *BeepCommand) Name() string {
    return "beep"
}

func (c *BeepCommand) Description() string {
    return "Replies 'boop!'"
}