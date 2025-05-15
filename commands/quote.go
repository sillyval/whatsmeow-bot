package commands

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"strings"

	"github.com/google/uuid"
	"go.mau.fi/whatsmeow"
	"go.mau.fi/whatsmeow/types"
	"go.mau.fi/whatsmeow/types/events"

	"whatsmeow-bot/utils"
)

func init() {
	RegisterCommand(&QuoteCommand{})
}

type QuoteCommand struct{}

func (c *QuoteCommand) Execute(client *whatsmeow.Client, message *events.Message, args []string) *string {
	if message.Message.ExtendedTextMessage == nil {
		utils.Reply(client, message, "You need to reply to a message to use this command.")
		utils.React(client, message, "❌")
		return nil
	}

	contextInfo := message.Message.ExtendedTextMessage.GetContextInfo()
	if contextInfo == nil || contextInfo.QuotedMessage == nil {
		utils.Reply(client, message, "You need to reply to a message to use this command.")
		utils.React(client, message, "❌")
		return nil
	}

	quotedMsg := contextInfo.QuotedMessage

	quotedText := utils.GetMessageBody(quotedMsg)

	if quotedText == "" {
		utils.Reply(client, message, "Quoted message must contain text.")
		utils.React(client, message, "❌")
		return nil
	}

	isGPT := utils.IsGPTMessage(quotedMsg);

	if false {//isGPT {
		quotedText = strings.Join(strings.Split(quotedText, " ")[1:], " ")
	}

	if !isGPT && utils.IsSystemMessage(quotedMsg) {
		utils.Reply(client, message, "You can't quote system messages")
		utils.React(client, message, "❌")
		return nil
	}

	var err error
	var senderName string

	if false {//if isGPT {
		senderName = "ChatGPT"
	} else {
		senderJID := contextInfo.GetParticipant()
		parsedJID, err := types.ParseJID(senderJID)
		if err != nil {
			fmt.Println("Error parsing JID")
			utils.React(client, message, "❌")
			return nil
		}

		contact, err := client.Store.Contacts.GetContact(parsedJID)
		
		if err == nil {
			if contact.PushName != "" {
				senderName = contact.PushName
			} else if contact.FullName != "" {
				senderName = contact.FullName
			} else {
				senderName = senderJID
			}
		} else {
			senderName = senderJID
		}
	}

	var filename string
	if false {//if isGPT {
		filename = "chatgpt"
	} else {
		filename = uuid.New().String()
	}
	savePath := fmt.Sprintf("/tmp/%s.jpg", filename)

	if false {//if isGPT {
		savePath = "quote/chatgpt.png"
	} else {

		senderJID := contextInfo.GetParticipant()
		parsedJID, err := types.ParseJID(senderJID)
		if err != nil {
			fmt.Println("Error parsing JID")
			utils.React(client, message, "❌")
			return nil
		}

		profilePic, err := client.GetProfilePictureInfo(parsedJID, &whatsmeow.GetProfilePictureParams{Preview: false})
		if err != nil || profilePic.URL == "" {
			fmt.Println("No profile picture found, using default black image.")
			err = ioutil.WriteFile(savePath, make([]byte, 600*600*3), 0644) // Empty black image
			if err != nil {
				utils.Reply(client, message, "Failed to create profile picture placeholder.")
				utils.React(client, message, "❌")
				return nil
			}
		} else {
			profilePicData, _, err := utils.DownloadMediaFromURL(profilePic.URL)
			if err != nil {
				utils.Reply(client, message, "Failed to download profile picture.")
				utils.React(client, message, "❌")
				return nil
			}
	
			err = ioutil.WriteFile(savePath, profilePicData, 0644)
			if err != nil {
				utils.Reply(client, message, "Failed to save profile picture.")
				utils.React(client, message, "❌")
				return nil
			}
		}
	}

	outputPath := "quote/" + filename + "_output.png"
	cmd := exec.Command("venv/bin/python", "quote/quote.py", senderName, quotedText, savePath)
	fmt.Println(cmd.String())
	err = cmd.Run()
	if err != nil {
		fmt.Println("Failed to execute quote.py:", err)
		utils.Reply(client, message, "Failed to generate quote image.")
		utils.React(client, message, "❌")
		return nil
	}

	imageData, err := ioutil.ReadFile(outputPath)
	if err != nil {
		utils.Reply(client, message, "Failed to read generated image.")
		utils.React(client, message, "❌")
		return nil
	}

	success := utils.ReplyImageToQuoted(client, message, *contextInfo.StanzaID, imageData, "image/png", "")
	if success {
		utils.React(client, message, "✅")
	} else {
		utils.React(client, message, "❌")
	}

	if !isGPT {
		os.Remove(savePath)
	}
	os.Remove(outputPath)

	return nil
}

func (c *QuoteCommand) Name() string {
	return "quote"
}

func (c *QuoteCommand) Description() string {
	return "Creates a quote image from a replied message."
}
