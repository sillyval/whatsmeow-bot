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
    nextIndex := cats.NextIndex(cats.CurrentIndex())

    var indexToPreview string
    var caption string
    if len(args) < 1 {
        indexToPreview = strconv.Itoa(nextIndex)
    } else {
        indexToPreview = args[0]
    }

    index, err := strconv.Atoi(indexToPreview)
    if err != nil || index > totalIndexes || index < 1 {
        utils.Reply(client, message, fmt.Sprintf("Invalid index, must be between 1-%v, got %v", totalIndexes, index))
        utils.React(client, message, "❌")
    }

    if index == nextIndex {
        caption = fmt.Sprintf("Preview for next image (%v/%v)\nCurrent caption: `%v`\nUse `caption: ...` to change.", index, totalIndexes, cats.Caption())
    } else {
        caption = fmt.Sprintf("Preview for %v/%v", index, totalIndexes)
    }

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