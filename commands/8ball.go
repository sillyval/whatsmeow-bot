package commands

import (
    "math/rand"

    "go.mau.fi/whatsmeow"
    "go.mau.fi/whatsmeow/types/events"

    "whatsmeow-bot/utils"
)

func init() {
    RegisterCommand(&EightBallCommand{})
}

type EightBallCommand struct{}

func (c *EightBallCommand) Execute(client *whatsmeow.Client, message *events.Message, args []string) *string {
    var responses = []string{
        "It is certain", "It is decidedly so", "Without a doubt", "Yes definitely", "You may rely on it",
        "As I see it, yes", "Most likely", "Outlook good", "Yes", "Signs point to yes",
        "Reply hazy, try again", "Ask again later", "Better not tell you now", "Cannot predict now", "Concentrate and ask again",
        "Don't count on it", "My reply is no", "My sources say no", "Outlook not so good", "Very doubtful",
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

func (c *EightBallCommand) Name() string {
    return "8ball"
}

func (c *EightBallCommand) Description() string {
    return "Roll the magic 8 ball!"
}