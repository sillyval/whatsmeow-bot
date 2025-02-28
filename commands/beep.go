package commands

import (
    "context"
    "fmt"

    "github.com/golang/protobuf/proto"
    "go.mau.fi/whatsmeow"
    "go.mau.fi/whatsmeow/types/events"
    waProto "go.mau.fi/whatsmeow/binary/proto"
)

// Register the BeepCommand
func init() {
    RegisterCommand(&BeepCommand{})
}

type BeepCommand struct{}

func (c *BeepCommand) Execute(client *whatsmeow.Client, message *events.Message, args []string) {
    // if message.Info.IsFromMe {
    //     return // Avoid replying to own messages
    // }

    textMessage := &waProto.Message{
        Conversation: proto.String("boop!"),
    }
    if client != nil {
        _, err := client.SendMessage(context.Background(), message.Info.Chat, textMessage)
        if err != nil {
            fmt.Println("Failed to send 'boop!':", err)
        }
    } else {
        fmt.Println("Client is not initialized")
    }
}

func (c *BeepCommand) Name() string {
    return "beep"
}

// Placeholder function - Implement to obtain your client instance if needed
func GetClient() *whatsmeow.Client {
    // Implementation to return the current WhatsApp client instance
    return nil
}