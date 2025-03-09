package commands

import (
	"fmt"
	"go.mau.fi/whatsmeow"
	"go.mau.fi/whatsmeow/types"
	"go.mau.fi/whatsmeow/types/events"
    "os"
	"os/exec"

    "io/ioutil"

	"github.com/google/uuid"

	waProto "go.mau.fi/whatsmeow/binary/proto"
	"google.golang.org/protobuf/proto"

	"whatsmeow-bot/utils"
)

func init() {
    RegisterCommand(&FakeQuoteCommand{})
}

type FakeQuoteCommand struct{}

func (c *FakeQuoteCommand) Execute(client *whatsmeow.Client, message *events.Message, args []string) {
    if len(args) < 4 {
        fmt.Println("Usage: command <chatID> <sender> <fakeMessage> <replyMessage>")
        return
    }
	
	if !message.Info.IsFromMe {
		return
	}

    chatID := args[0]
    sender := args[1]
    fakeMessageText := args[2]
    replyMessageText := args[3]
	var stanzaID string
	if len(args) == 5 {
		stanzaID = args[4]
	} else {
		stanzaID = "made-by-sillyval:3"
	}

	fmt.Println(replyMessageText)

    chatJID, err := types.ParseJID(chatID)
    if err != nil {
        fmt.Printf("Invalid chat ID: %s\n", err)
        return
    }
    senderJID, err := types.ParseJID(sender)
    if err != nil {
        fmt.Printf("Invalid sender ID: %s\n", err)
        return
    }

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




    contact, err := client.Store.Contacts.GetContact(senderJID)
	var senderName string
	if err == nil {
		if contact.PushName != "" {
			senderName = contact.PushName
		} else if contact.FullName != "" {
			senderName = contact.FullName
		} else {
			senderName = sender
		}
	} else {
		senderName = sender
	}

	filename := uuid.New().String()
	savePath := fmt.Sprintf("/tmp/%s.jpg", filename)

	profilePic, err := client.GetProfilePictureInfo(senderJID, &whatsmeow.GetProfilePictureParams{Preview: false})
	if err != nil || profilePic.URL == "" {
		fmt.Println("No profile picture found, using default black image.")
		err = ioutil.WriteFile(savePath, make([]byte, 600*600*3), 0644) // Empty black image
		if err != nil {
			utils.Reply(client, message, "Failed to create profile picture placeholder.")
			utils.React(client, message, "❌")
			return
		}
	} else {
		profilePicData, _, err := utils.DownloadMediaFromURL(profilePic.URL)
		if err != nil {
			utils.Reply(client, message, "Failed to download profile picture.")
			utils.React(client, message, "❌")
			return
		}

		err = ioutil.WriteFile(savePath, profilePicData, 0644)
		if err != nil {
			utils.Reply(client, message, "Failed to save profile picture.")
			utils.React(client, message, "❌")
			return
		}
	}

	outputPath := "quote/" + filename + "_output.png"
	cmd := exec.Command("venv/bin/python", "quote/quote.py", senderName, fakeMessageText, savePath)
	fmt.Println(cmd.String())
	err = cmd.Run()
	if err != nil {
		fmt.Println("Failed to execute quote.py:", err)
		utils.Reply(client, message, "Failed to generate quote image.")
		utils.React(client, message, "❌")
		return
	}

	imageData, err := ioutil.ReadFile(outputPath)
	if err != nil {
		utils.Reply(client, message, "Failed to read generated image.")
		utils.React(client, message, "❌")
		return
	}

	success = utils.ReplyImageToChat(client, message, chatJID, *contextInfo.StanzaID, imageData, "image/png", "")
	if success {
		utils.React(client, message, "✅")
	} else {
		utils.React(client, message, "❌")
	}

	os.Remove(savePath)
	os.Remove(outputPath)
}

func (c *FakeQuoteCommand) Name() string {
    return "fakequote"
}

func (c *FakeQuoteCommand) Description() string {
    return ""
}