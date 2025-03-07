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

	"whatsmeow-bot/utils"
)

// Register the EmptyStatusCommand
func init() {
    RegisterCommand(&EmptyStatusCommand{})
}

type EmptyStatusCommand struct{}

func (c *EmptyStatusCommand) Execute(client *whatsmeow.Client, message *events.Message, args []string) {

    fmt.Println("--- SENDING STATUS UPDATE ---")

    if len(args) == 0 {
	args[0] = ""
    }

    if !message.Info.IsFromMe {
		utils.React(client, message, "❌")
		return
    }

    // Join the arguments to form the status text
    statusText := strings.Join(args, " ")

	// Ensure the message is a text message
	if message.Info.Type != "text" {
		utils.React(client, message, "❌")
		return
	}


	// Construct the status update message
	statusMessage := &waProto.Message{
		ExtendedTextMessage: &waProto.ExtendedTextMessage{
			Text:           proto.String(statusText),
			BackgroundArgb: proto.Uint32(0x00000000),
			TextArgb:       proto.Uint32(0x00FFFFFF),
			Font:           waProto.ExtendedTextMessage_SYSTEM.Enum(),
		},
	}

	// Use types.StatusBroadcastJID for posting status updates
	_, err := client.SendMessage(context.Background(), types.StatusBroadcastJID, statusMessage)
	if err != nil {
		fmt.Printf("Failed to post status update: %v\n", err)
		utils.React(client, message, "❌")
	} else {
		fmt.Println("Status update posted successfully.")
		utils.React(client, message, "✅")
	}
}

func (c *EmptyStatusCommand) Name() string {
    return "emptystatus"
}
