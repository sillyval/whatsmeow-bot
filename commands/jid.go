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

func (c *JIDCommand) Execute(client *whatsmeow.Client, message *events.Message, args []string) *string {
	currentJID := message.Info.Chat.String()
    text := fmt.Sprintf("Current chat JID: %s", currentJID)

    if utils.IsNewsletter(message) {
        utils.NewsletterMessage(client, message.Info.Chat, text)
    } else {
        utils.Reply(client, message, text)
    }
    return nil
}

func (c *JIDCommand) Name() string {
    return "jid"
}

func (c *JIDCommand) Description() string {
    return "Get the JID of the current chat"
}