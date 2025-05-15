package commands

import (
	"fmt"
	"go.mau.fi/whatsmeow"
	"go.mau.fi/whatsmeow/types"
	"go.mau.fi/whatsmeow/types/events"

	waProto "go.mau.fi/whatsmeow/binary/proto"
	"google.golang.org/protobuf/proto"

	"whatsmeow-bot/utils"
)

func init() {
    RegisterCommand(&RawReplyCommand{})
}

type RawReplyCommand struct{}

func (c *RawReplyCommand) Execute(client *whatsmeow.Client, message *events.Message, args []string) *string {
    if len(args) < 5 {
        fmt.Println("Usage: command <chatID> <messageID> <sender> <fakeMessage> <replyMessage>")
        return nil
    }
	
	if !message.Info.IsFromMe {
		return nil
	}

    chatID := args[0]
    stanzaID := args[1]
    sender := args[2]
    fakeMessageText := args[3]
    replyMessageText := args[4]

    // Parse chat ID
    chatJID, err := types.ParseJID(chatID)
    if err != nil {
        fmt.Printf("Invalid chat ID: %s\n", err)
        return nil
    }

    // Construct a fake previous message
    fakeMessage := &waProto.Message{
        Conversation: proto.String(fakeMessageText),
    }

	if message.Message.ImageMessage != nil {
        fakeMessage = &waProto.Message{ImageMessage: &waProto.ImageMessage{
            URL:                message.Message.ImageMessage.URL,
            Caption:            proto.String(fakeMessageText),
            DirectPath:         message.Message.ImageMessage.DirectPath,
            MediaKey:           message.Message.ImageMessage.MediaKey,
            MediaKeyTimestamp:  message.Message.ImageMessage.MediaKeyTimestamp,
            JPEGThumbnail:      message.Message.ImageMessage.JPEGThumbnail,
			FileSHA256:         message.Message.ImageMessage.FileSHA256,
			FileEncSHA256:      message.Message.ImageMessage.FileEncSHA256,
        }}
    }

    contextInfo := &waProto.ContextInfo{
        StanzaID:      proto.String(stanzaID),
        Participant:   proto.String(sender),
        QuotedMessage: fakeMessage,
    }

    // Send the reply
    success := utils.ReplyMessage(client, chatJID, fakeMessage, contextInfo, replyMessageText)
    if success {
        fmt.Println("Reply sent successfully!")
		utils.React(client, message, "✅")
    } else {
        fmt.Println("Failed to send the reply.")
		utils.React(client, message, "❌")
    }
    
    return nil
}

func (c *RawReplyCommand) Name() string {
    return "rawreply"
}

func (c *RawReplyCommand) Description() string {
    return ""
}