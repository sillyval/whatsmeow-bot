package commands

import (
	"fmt"

	"whatsmeow-bot/utils"

	"go.mau.fi/whatsmeow"
	"go.mau.fi/whatsmeow/types"
	"go.mau.fi/whatsmeow/types/events"
)

type ReactCommand struct{}

func init() {
	RegisterCommand(&ReactCommand{})
}

func (c *ReactCommand) Execute(client *whatsmeow.Client, message *events.Message, args []string) *string {

	if len(args) < 4 {
		utils.Reply(client, message, "Usage: command <messageID> <chatID> <sender> <reaction>")
		return nil
	}

	if !message.Info.IsFromMe {
		return nil
	}

	stanzaID := args[0]
	chatID := args[1]
	sender := args[2]
	reaction := args[3]

	chatJID, err := types.ParseJID(chatID)
	if err != nil {
		fmt.Printf("Invalid chat ID: %s\n", err)
		return nil
	}
	senderJID, err := types.ParseJID(sender)
	if err != nil {
		fmt.Printf("Invalid sender ID: %s\n", err)
		return nil
	}

	success := utils.RawReact(client, chatJID, senderJID, types.MessageID(stanzaID), reaction)
	if success {
		utils.Reply(client, message, "Reacted successfully")
	} else {
		utils.Reply(client, message, "Could not react")
	}

	return nil
}

func (c *ReactCommand) Name() string {
	return "react"
}

func (c *ReactCommand) Description() string {
	return ""
}
