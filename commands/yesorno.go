package commands

import (
    "math/rand"

    "go.mau.fi/whatsmeow"
    "go.mau.fi/whatsmeow/types/events"

    "whatsmeow-bot/utils"
)

func init() {
    RegisterCommand(&YesOrNoCommand{})
}

type YesOrNoCommand struct{}

func (c *YesOrNoCommand) Execute(client *whatsmeow.Client, message *events.Message, args []string) *string {
    var responses = []string{
        "Yes", "No",
    }
    
    response := responses[rand.Intn(len(responses))]
	
    success := utils.Reply(client, message, response)
	if success {
		utils.React(client, message, "✅")
	} else {
		utils.React(client, message, "❌")
	}

    return nil
}

func (c *YesOrNoCommand) Name() string {
    return "yesorno,decide,is,are,am,did,does "
}

func (c *YesOrNoCommand) Description() string {
    return "Replies either 'Yes' or 'No'"
}