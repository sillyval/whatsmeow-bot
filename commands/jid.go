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

    quotedMessage := utils.GetQuotedMessage(message)
    if quotedMessage != nil {
        // Get JID of a quoted message

        quotedMessage := utils.GetQuotedMessage(message)
        if quotedMessage != nil {
            contextInfo := utils.GetMessageContextInfo(message.Message)
            if contextInfo != nil {

                conversationJID := contextInfo.StanzaID

                utils.Reply(client, message, *conversationJID)
                utils.React(client, message, "✅") 
                return nil    
            }
        }
    }

	currentJID := message.Info.Chat.String()
    text := fmt.Sprintf("Current chat JID: %s", currentJID)

    if utils.IsNewsletter(message) {
        utils.NewsletterMessage(client, message.Info.Chat, text)
        utils.NewsletterReact(client, message, "✅")
    } else {
        utils.Reply(client, message, text)
        utils.React(client, message, "✅")
    }
    return nil
}

func (c *JIDCommand) Name() string {
    return "jid"
}

func (c *JIDCommand) Description() string {
    return "Get the JID of the current chat"
}