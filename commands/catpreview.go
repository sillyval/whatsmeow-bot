package commands

import (
    "fmt"
	"strconv"
	"whatsmeow-bot/cats"
	"whatsmeow-bot/utils"

	"go.mau.fi/whatsmeow"
	"go.mau.fi/whatsmeow/types/events"
)

type CatPreviewCommand struct{}

func init() {
    RegisterCommand(&CatPreviewCommand{})
}


func (c *CatPreviewCommand) Execute(client *whatsmeow.Client, message *events.Message, args []string) *string {

    totalIndexes := cats.TotalImages()

    indexToPreview := args[0]
    index, err := strconv.Atoi(indexToPreview)
    if err != nil || index > totalIndexes || index < 1 {
        utils.Reply(client, message, fmt.Sprintf("Invalid index, must be between 1-%v, got %v", totalIndexes, index))
        utils.React(client, message, "❌")
    }

    caption := fmt.Sprintf("Preview for %v", index)
    cats.PreviewNextUpload(client, index-1, totalIndexes, &caption)

    if utils.IsNewsletter(message) {
        utils.NewsletterReact(client, message, "✅")
    } else {
        utils.React(client, message, "✅")
    }

    return nil
}

func (c *CatPreviewCommand) Name() string {
    return "catpreview"
}

func (c *CatPreviewCommand) Description() string {
    return ""
}