package commands

import (
    "fmt"
    "strings"

    "go.mau.fi/whatsmeow"
    "go.mau.fi/whatsmeow/types"
    "go.mau.fi/whatsmeow/types/events"
    waProto "go.mau.fi/whatsmeow/binary/proto"
    "google.golang.org/protobuf/proto"

    "whatsmeow-bot/utils"
)

// Register the EveryoneStatusCommand
func init() {
    RegisterCommand(&EveryoneStatusCommand{})
}

type EveryoneStatusCommand struct{}

func (c *EveryoneStatusCommand) Execute(client *whatsmeow.Client, message *events.Message, args []string) *string {
    if !message.Info.IsFromMe {
        utils.React(client, message, "❌")
        return nil
    }

    if len(args) == 0 {
        args = append(args, "") // Allow empty statuses
    }

    // Join the arguments to form the status text
    statusText := strings.Join(args, " ")

    // Ensure the message is a text message
    if message.Info.Type != "text" {
        utils.React(client, message, "❌")
        return nil
    }

    // Fetch JIDs of recipients who will see the status
    recipients, err := client.DangerousInternals().GetStatusBroadcastRecipients()
    if err != nil {
        fmt.Println("❌ Error fetching status recipients:", err)
        utils.React(client, message, "❌")
        return nil
    }

    // Create the mentions list
	var mentions []string
	for _, recipient := range recipients {
		mentions = append(mentions, recipient.User)
	}

	// Log recipients
	fmt.Println("✅ Status will be visible to:")
	for _, mention := range mentions {
		fmt.Println(" -", mention)
	}

    // Construct the status update message
    statusMessage := &waProto.Message{
        ExtendedTextMessage: &waProto.ExtendedTextMessage{
            Text:           proto.String(statusText),
            BackgroundArgb: proto.Uint32(0xFF000000), // Black
            TextArgb:       proto.Uint32(0xFFFFFFFF), // White
            Font:           waProto.ExtendedTextMessage_SYSTEM.Enum(),
        },
    }

    // Use types.StatusBroadcastJID for posting status updates
	
	response := *utils.SendExtendedMessageWithMentions(client, types.StatusBroadcastJID, statusMessage, mentions)
    utils.React(client, message, "✅")

    return &response.ID
}

func (c *EveryoneStatusCommand) Name() string {
    return "everyonestatus"
}

func (c *EveryoneStatusCommand) Description() string {
    return "Posts a status update and logs recipients."
}
