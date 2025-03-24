package commands

import (
	"fmt"
	"whatsmeow-bot/cats"
	"whatsmeow-bot/utils"

	"go.mau.fi/whatsmeow"
	"go.mau.fi/whatsmeow/types/events"
)

type CatIndexCommand struct{}

func init() {
    RegisterCommand(&CatIndexCommand{})
}


func (c *CatIndexCommand) Execute(client *whatsmeow.Client, message *events.Message, args []string) *string {

    currentIndex := cats.CurrentIndex()
    nextIndex := cats.NextIndex(currentIndex)
    totalImages := cats.TotalImages()

    if utils.IsNewsletter(message) {
        utils.NewsletterMessage(client, message.Info.Chat, fmt.Sprintf("Current index: %v/%v\nNext index: %v/%v", currentIndex, totalImages, nextIndex, totalImages))
        utils.NewsletterReact(client, message, "✅")
    } else {
        utils.Reply(client, message, fmt.Sprintf("Current index: %v/%v\nNext index: %v/%v", currentIndex, totalImages, nextIndex, totalImages))
        utils.React(client, message, "✅")
    }

    return nil
}

func (c *CatIndexCommand) Name() string {
    return "catindex"
}

func (c *CatIndexCommand) Description() string {
    return ""
}