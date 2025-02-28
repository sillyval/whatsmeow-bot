package commands

import (
    "context"
    "fmt"
    "strings"

    "go.mau.fi/whatsmeow"
	"go.mau.fi/whatsmeow/types"
    "go.mau.fi/whatsmeow/types/events"
    waProto "go.mau.fi/whatsmeow/binary/proto"
    "google.golang.org/protobuf/proto"
)

// Register the StatusCommand
func init() {
    RegisterCommand(&StatusCommand{})
}

type StatusCommand struct{}

func (c *StatusCommand) Execute(client *whatsmeow.Client, message *events.Message, args []string) {

    fmt.Println("--- SENDING STATUS UPDATE ---")

    if len(args) == 0 {
        fmt.Println("No text provided for status update.")
        return
    }

    if !message.Info.IsFromMe {
        // Prevent processing own messages for status updates
        return
    }

    // Join the arguments to form the status text
    statusText := strings.Join(args, " ")

	// Ensure the message is a text message
	if message.Info.Type != "text" {
		return
	}

	var reaction *waProto.Message

	// Construct the status update message
	statusMessage := &waProto.Message{
		ExtendedTextMessage: &waProto.ExtendedTextMessage{
			Text:           proto.String(statusText),
			BackgroundArgb: proto.Uint32(0xFF000000), // Example ARGB color (black)
			TextArgb:       proto.Uint32(0xFFFFFFFF), // Example ARGB color (white)
			Font:           waProto.ExtendedTextMessage_SYSTEM.Enum(), // Example font
		},
	}

	// Use types.StatusBroadcastJID for posting status updates
	_, err := client.SendMessage(context.Background(), types.StatusBroadcastJID, statusMessage)
	if err != nil {
		fmt.Printf("Failed to post status update: %v\n", err)
		reaction = client.BuildReaction(message.Info.Chat, message.Info.Sender, message.Info.ID, "❌")
	} else {
		fmt.Println("Status update posted successfully.")
		reaction = client.BuildReaction(message.Info.Chat, message.Info.Sender, message.Info.ID, "✅")
	}

	// Send the reaction
	if reaction != nil {
		_, err := client.SendMessage(context.Background(), message.Info.Chat, reaction)
		if err != nil {
			fmt.Printf("Failed to send reaction: %v\n", err)
		}
	}

}

func (c *StatusCommand) Name() string {
    return "status"
}