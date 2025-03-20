package discordbot

import (
	"bytes"
	"fmt"
	"mime"
	"strings"
	"whatsmeow-bot/utils"
	"time"

	"github.com/bwmarrin/discordgo"
)

var (
	ForumChannelID = "1349513744082014309" 
	GuildID =        "1347216013891997798"
	session          *discordgo.Session
)

func InitBot(token string) error {
	var err error
	session, err = discordgo.New("Bot " + token)

	if err != nil {
		return fmt.Errorf("error creating Discord session: %v", err)
	}

	shard := [2]int{0, 3}
	session.Identify.Shard = &shard

	session.Identify.Intents = discordgo.IntentsGuildMessages | discordgo.IntentsGuilds | discordgo.IntentsGuildMessageReactions

	err = session.Open()
	if err != nil {
		return fmt.Errorf("error opening Discord session: %v", err)
	}

	fmt.Printf("Shard 1/3 (val) is running...\n",)
	return nil
}

func FindOrCreateForum(jid, contactName, whatsappUsername string) (string, error) {
	channel, err := session.Channel(ForumChannelID)
	if err != nil {
		return "", fmt.Errorf("failed to retrieve channel: %v", err)
	}

	if channel.Type != discordgo.ChannelTypeGuildForum {
		return "", fmt.Errorf("the specified channel is not a forum channel")
	}

	threads, err := session.GuildThreadsActive(GuildID)
	if err != nil {
		fmt.Printf("Error retrieving active threads: %v\n", err)
		return "", fmt.Errorf("failed to get active threads: %v", err)
	}

	for _, thread := range threads.Threads {
		if thread.ParentID == ForumChannelID {
			parts := strings.Split(thread.Name, " / ")
			if len(parts) > 0 && parts[0] == jid {
				return thread.ID, nil
			}
		}
	}

	startingMsg := &discordgo.MessageSend{
		Content: "Status log of "+jid,	
	}

	thread, err := session.ForumThreadStartComplex(ForumChannelID, &discordgo.ThreadStart{
		Name:                fmt.Sprintf("%s / %s / %s", jid, contactName, whatsappUsername),
		AutoArchiveDuration: 7 * 24 * 60, // 7 days, maximum discord will allow (in minutes)
		Type: 				 discordgo.ChannelTypeGuildPublicThread,
	}, startingMsg)

	if err != nil {
		return "", fmt.Errorf("failed to create forum thread: %v", err)
	}

	return thread.ID, nil
}

func JIDToPhoneNumber(jid string) string {
	phoneNumber := strings.Split(jid, "@")[0]
	return "+" + phoneNumber[0:2] + " " + phoneNumber[2:]
}

func formatAMPM(hour int) string {
    if hour < 12 {
        return "am"
    }
    return "pm"
}

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
    return fmt.Sprintf("%d%s %s %d, %d:%02d%s",
        day, suffix, t.Month().String(), t.Year(), t.Hour()%12, t.Minute(), formatAMPM(t.Hour()))
}

type AlreadyExistsError struct{}
func (e *AlreadyExistsError) Error() string{
	return "Message already exists!"
}

func LogStatus(jid, contactName, whatsappUsername, statusText, statusJID string, colour *int, updateTitle bool) error {

	message := GetMessageByJID(jid, contactName, whatsappUsername, statusJID)
	if message != nil {
		return &AlreadyExistsError{};
	}
	
	jid = JIDToPhoneNumber(jid)

	fmt.Printf("\nFinding forum...\n")
	threadID, err := FindOrCreateForum(jid, contactName, whatsappUsername)
	if err != nil {
		return err
	}
	fmt.Printf("\nFound forum!\n")

	if colour == nil {
		colourDefault := 0xffbeef // cotton candy pink :3
		colour = &colourDefault
	}

	currentTime := FormatDateWithOrdinal(time.Now())

	footer := &discordgo.MessageEmbedFooter{
		Text: fmt.Sprintf("%s • %s", currentTime, statusJID),
	}

	embed := &discordgo.MessageEmbed{
		Title:       "",
		Description: statusText,
		Footer:      footer,
		Color:       *colour,
	}

	if updateTitle {
		session.ChannelEdit(threadID, &discordgo.ChannelEdit{
			Name: fmt.Sprintf("%s / %s / %s", jid, contactName, whatsappUsername),
		})
	}

	_, err = session.ChannelMessageSendEmbed(threadID, embed)
	if err != nil {
		return fmt.Errorf("failed to send status message: %v", err)
	}

	//fmt.Printf("\nEmbed sent: %s\n", embed)
	fmt.Printf("\nEmbed sent\n")

	return nil
}


func GetFileExtension(mimetype string) string {
	exts, err := mime.ExtensionsByType(mimetype)
	if err != nil || len(exts) == 0 {
		// Fallback: Extract from MIME type manually
		parts := strings.Split(mimetype, "/")
		if len(parts) == 2 {
			return "." + parts[1]
		}
		return ""
	}
	
	// Append the first extension returned by mime.ExtensionsByType
	return exts[0]
}


func LogStatusWithMedia(jid, contactName, whatsappUsername, statusText, statusJID string, mediaData []byte, mimetype string, updateTitle bool) error {
	
	message := GetMessageByJID(jid, contactName, whatsappUsername, statusJID)
	if message != nil {
		return &AlreadyExistsError{};
	}

	jid = JIDToPhoneNumber(jid)

	fmt.Printf("\nFinding forum...\n")
	threadID, err := FindOrCreateForum(jid, contactName, whatsappUsername)
	if err != nil {
		return err
	}
	fmt.Printf("\nFound forum!\n")

	if updateTitle {
		session.ChannelEdit(threadID, &discordgo.ChannelEdit{
			Name: fmt.Sprintf("%s / %s / %s", jid, contactName, whatsappUsername),
		})
	}

	currentTime := FormatDateWithOrdinal(time.Now())

	footer := &discordgo.MessageEmbedFooter{
		Text: fmt.Sprintf("%s • %s", currentTime, statusJID),
	}

	embed := &discordgo.MessageEmbed{
		Title:       "",
		Description: statusText,
		Footer:      footer,
		Color:       0xffbeef,
	}

	file := &discordgo.File{
		Name:   "whatsapp-status-log",
	}

	msg := &discordgo.MessageSend{
		Embeds: []*discordgo.MessageEmbed{embed},
		Files:  []*discordgo.File{file},
	}

	if strings.HasPrefix(mimetype, "image/") {
		embed.Image = &discordgo.MessageEmbedImage{
			URL: "attachment://whatsapp-status-log.png",
		}
		file.Reader = bytes.NewReader(utils.ConvertToPNG(mediaData))
		file.Name = file.Name + ".png"
	} else if strings.HasPrefix(mimetype, "video/") {
		embed.Video = &discordgo.MessageEmbedVideo{
			URL: "attachment://whatsapp-status-log.mp4",
		}
		file.Reader = bytes.NewReader(utils.ConvertToMP4(mediaData))
		file.Name = file.Name + ".mp4"
	} else if strings.HasPrefix(mimetype, "audio/") {
		// No direct way of embedding
		file.Reader = bytes.NewReader(utils.ConvertToMP3(mediaData))
		file.Name = file.Name + ".mp3"
	}

	_, err = session.ChannelMessageSendComplex(threadID, msg)
	if err != nil {
		return fmt.Errorf("failed to send status message with media: %v", err)
	}

	//fmt.Printf("\nEmbed sent: %s\n", embed)
	fmt.Printf("\nEmbed sent\n")

	return nil
}

func GetMessageByJID(jid, contactName, whatsappUsername, statusJID string) *discordgo.Message {
	jid = JIDToPhoneNumber(jid)

	threadID, err := FindOrCreateForum(jid, contactName, whatsappUsername)
	if err != nil {
		return nil
	}

	messages, err := session.ChannelMessages(threadID, 50, "", "", "")
	if err != nil {
		return nil
	}


	for _, message := range messages {

		if len(message.Embeds) != 1 {
			continue
		}

		embed := message.Embeds[0]
		footer := embed.Footer

		if footer == nil {
			continue
		}

		thisStatusJID := strings.Split(footer.Text, " • ")[1]
		if thisStatusJID == statusJID {
			return message;
		}
	}

	return nil
}

func React(jid, contactName, whatsappUsername, statusJID string, reaction string) {
	message := GetMessageByJID(jid, contactName, whatsappUsername, statusJID)
	if message == nil {
		return;
	}

	jid = JIDToPhoneNumber(jid)

	threadID, err := FindOrCreateForum(jid, contactName, whatsappUsername)
	if err != nil {
		return
	}

	session.MessageReactionAdd(threadID, message.ID, reaction)
}

func CloseBot() {
	if session != nil {
		fmt.Println("Shutting down the bot...")
		session.Close()
	}
}
