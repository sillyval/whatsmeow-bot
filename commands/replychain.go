package commands

import (
	"fmt"
	"strings"
	"whatsmeow-bot/utils"

	"go.mau.fi/whatsmeow"
	"go.mau.fi/whatsmeow/types/events"
)

type ReplyChainCommand struct{}

func init() {
    RegisterCommand(&ReplyChainCommand{})
}

func (c *ReplyChainCommand) Execute(client *whatsmeow.Client, message *events.Message, args []string) *string {

	fmt.Println(message.Message.GetConversation())
	fmt.Println(message.RawMessage.Conversation)

    chain := utils.GetReplyChain(message)

    if len(chain) > 0 {
        replyContent := strings.Join(chain, " > ")
        utils.Reply(client, message, replyContent)
    } else {
        utils.Reply(client, message, "No prior messages in the chain.")
    }

    return nil
}

func (c *ReplyChainCommand) Name() string {
    return "replychain"
}

func (c *ReplyChainCommand) Description() string {
    return ""
}