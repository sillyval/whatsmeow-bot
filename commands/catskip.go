package commands

import (
    "go.mau.fi/whatsmeow"
    "go.mau.fi/whatsmeow/types/events"
    "whatsmeow-bot/utils"
    "whatsmeow-bot/cats"
)

type CatSkipCommand struct{}

func init() {
    RegisterCommand(&CatSkipCommand{})
}


func (c *CatSkipCommand) Execute(client *whatsmeow.Client, message *events.Message, args []string) *string {

    cats.Skip(client)

    if utils.IsNewsletter(message) {
        utils.NewsletterReact(client, message, "✅")
    } else {
        utils.React(client, message, "✅")
    }

    return nil
}

func (c *CatSkipCommand) Name() string {
    return "catskip"
}

func (c *CatSkipCommand) Description() string {
    return ""
}