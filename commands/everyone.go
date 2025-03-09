package commands

import (
    "fmt"
	"strings"
    "go.mau.fi/whatsmeow"
    "go.mau.fi/whatsmeow/types/events"
    "whatsmeow-bot/utils"
)

func init() {
    RegisterCommand(&PingAllCommand{})
}

type PingAllCommand struct{}

func (c *PingAllCommand) Execute(client *whatsmeow.Client, message *events.Message, args []string) {
    if !message.Info.IsGroup {
        utils.Reply(client, message, "This command can only be used in a group!")
        return
    }

	appendMessage := strings.Join(args, " ")

    groupJid := message.Info.Chat
    groupInfo, err := client.GetGroupInfo(groupJid)
    if err != nil {
        fmt.Println("Error fetching group info:", err)
        utils.Reply(client, message, "Failed to retrieve group information.")
        return
    }
    
    //myJID := client.Store.ID.ToNonAD().User

    var mentions []string
    for _, participant := range groupInfo.Participants {
        userJID := participant.JID.User
        //if userJID != myJID {
            mentions = append(mentions, userJID)
        //}
    }
    
    if len(mentions) == 0 {
        utils.Reply(client, message, "No members to mention!")
        return
    }

    messageText := "@everyone"
	var formattedMessage string
	if appendMessage == "" {
		formattedMessage = messageText
	} else {
		formattedMessage = messageText + " | " + appendMessage
	}
    utils.SendMessageWithMentions(client, message.Info.Chat, formattedMessage, mentions)
    utils.React(client, message, "✅")
}

func (c *PingAllCommand) Name() string {
    return "everyone"
}

func (c *PingAllCommand) Description() string {
    return "Pings all members in the current group"
}