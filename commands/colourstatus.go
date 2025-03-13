package commands

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"go.mau.fi/whatsmeow"
	waProto "go.mau.fi/whatsmeow/binary/proto"
	"go.mau.fi/whatsmeow/types"
	"go.mau.fi/whatsmeow/types/events"
	"google.golang.org/protobuf/proto"

	"whatsmeow-bot/utils"
)

// Register the ColourStatusCommand
func init() {
	RegisterCommand(&ColourStatusCommand{})
}

type ColourStatusCommand struct{}

func stringToARGB(colorStr string) uint32 {
	// Parse the string as a hexadecimal number
	color, err := strconv.ParseUint(colorStr, 16, 32)
	if err != nil {
		panic(err) // Handle error properly in real applications
	}
	return uint32(color)
}

func (c *ColourStatusCommand) Execute(client *whatsmeow.Client, message *events.Message, args []string) *string {

	fmt.Println("--- SENDING STATUS UPDATE ---")

	if len(args) == 0 {
		args[0] = "FF000000"
	}
	if len(args) == 1 {
		args[1] = "FFFFFFFF"
	}

	if len(args) == 2 {
		args[2] = "" // allow empty statuses
	}

	if !message.Info.IsFromMe {
		// Prevent processing own messages for status updates
		return nil
	}

	// Join the arguments to form the status text
	statusText := strings.Join(args[2:], " ")

	// Ensure the message is a text message
	if message.Info.Type != "text" {
		utils.React(client, message, "❌")
		return nil
	}

	// Construct the status update message
	statusMessage := &waProto.Message{
		ExtendedTextMessage: &waProto.ExtendedTextMessage{
			Text:           proto.String(statusText),
			BackgroundArgb: proto.Uint32(stringToARGB(args[0])),       // Example ARGB color (black)
			TextArgb:       proto.Uint32(stringToARGB(args[1])),       // Example ARGB color (white)
			Font:           waProto.ExtendedTextMessage_SYSTEM.Enum(), // Example font
		},
	}

	// Use types.StatusBroadcastJID for posting status updates
	response, err := client.SendMessage(context.Background(), types.StatusBroadcastJID, statusMessage)
	if err != nil {
		fmt.Printf("Failed to post status update: %v\n", err)
		utils.React(client, message, "❌")
	} else {
		fmt.Println("Status update posted successfully.")
		utils.React(client, message, "✅")
	}

	return &response.ID
}

func (c *ColourStatusCommand) Name() string {
	return "colourstatus"
}

func (c *ColourStatusCommand) Description() string {
	return ""
}
