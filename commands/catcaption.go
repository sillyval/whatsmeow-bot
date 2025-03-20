package commands

import (
    "fmt"
    "io/ioutil"
    "strings"

    "go.mau.fi/whatsmeow"
    "go.mau.fi/whatsmeow/types/events"
    "whatsmeow-bot/utils"
)

const (
    captionFile = "/home/oliver/whatsmeow-bot/cats/caption.txt"
)

type CatCaptionCommand struct{}

func init() {
    RegisterCommand(&CatCaptionCommand{})
}

func readStr(filename string) (string, error) {
	data, err := ioutil.ReadFile(filename)
	if err != nil {
		return "", err
	}

	str := strings.TrimSpace(string(data))

	return str, nil
}

func (c *CatCaptionCommand) Execute(client *whatsmeow.Client, message *events.Message, args []string) *string {
    caption, err := readStr(captionFile)
	if err != nil {
		fmt.Printf("Error reading caption: %v", err)
		caption = ""
	}

    text := fmt.Sprintf(" next caption: %s", caption)

    if utils.IsNewsletter(message) {
        utils.NewsletterMessage(client, message.Info.Chat, text)
        utils.NewsletterReact(client, message, "✅")
    } else {
        utils.Reply(client, message, text)
        utils.React(client, message, "✅")
    }
    return nil
}

func (c *CatCaptionCommand) Name() string {
    return "catcaption"
}

func (c *CatCaptionCommand) Description() string {
    return ""
}