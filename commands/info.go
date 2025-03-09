package commands

import (
    "fmt"
	"time"

    "go.mau.fi/whatsmeow"
    "go.mau.fi/whatsmeow/types/events"

    "whatsmeow-bot/utils"
)

func init() {
    RegisterCommand(&InfoCommand{})
}

type InfoCommand struct{}

func FormatDateWithOrdinal(t time.Time) string {
    day := t.Day()
    suffix := "th"
    if day%10 == 1 && day != 11 {
        suffix = "st"
    } else if day%10 == 2 && day != 12 {
        suffix = "nd"
    } else if day%10 == 3 && day != 13 {
        suffix = "rd"
    }

    // Format the output string
    return fmt.Sprintf("%d%s %s %d at %d:%02d%s",
        day, suffix, t.Month().String(), t.Year(), t.Hour()%12, t.Minute(), formatAMPM(t.Hour()))
}

// Helper to get am/pm from hour
func formatAMPM(hour int) string {
    if hour < 12 {
        return "am"
    }
    return "pm"
}

func (c *InfoCommand) Execute(client *whatsmeow.Client, message *events.Message, args []string) {
    if !message.Info.IsGroup {
        utils.Reply(client, message, "This command can only be used in a group!")
        return
    }

    groupJid := message.Info.Chat
    groupInfo, err := client.GetGroupInfo(groupJid)
    if err != nil {
        fmt.Println("Error fetching group info:", err)
        utils.Reply(client, message, "Failed to retrieve group information.")
        return
    }

	fmt.Println(groupInfo)

	groupName := groupInfo.GroupName.Name
	groupNameOwnerJid := groupInfo.GroupName.NameSetBy.User
    ownerJid := groupInfo.OwnerJID.User
    description := groupInfo.Topic

	descOwner := groupInfo.TopicSetBy.User
    descOwnerJid := groupInfo.TopicSetBy.User
    createdAt := groupInfo.GroupCreated
	dateString := FormatDateWithOrdinal(createdAt)

    chatJID := groupInfo.JID.String()

    // Format and send the message
    replyMessage := fmt.Sprintf(`_*Group Details*_

*Name*: %s
*Name Set By*: @%s
*Description*: %s
*Description Set By*: @%s
*Chat JID*: %s
*Created At*: %s
*Created By*: @%s
*Participant Count*: %d`, 
groupName, groupNameOwnerJid, description, descOwner, chatJID, dateString, ownerJid, len(groupInfo.Participants))

    mentions := []string{groupNameOwnerJid, descOwnerJid, ownerJid}
    utils.SendMessageWithMentions(client, message.Info.Chat, replyMessage, mentions)
    utils.React(client, message, "✅")
}

func (c *InfoCommand) Name() string {
    return "info"
}

func (c *InfoCommand) Description() string {
    return "Replies with the group chat info"
}