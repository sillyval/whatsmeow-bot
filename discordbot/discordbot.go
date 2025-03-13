package discordbot

import (
	"bytes"
	"fmt"
	"log"
	"strings"
	"mime"

	"github.com/bwmarrin/discordgo"
)

var (
	ForumChannelID = "1349513744082014309" 
	GuildID = "1347216013891997798"
	session        *discordgo.Session
)

func InitBot(token string) error {
	var err error
	session, err = discordgo.New("Bot " + token)

	if err != nil {
		return fmt.Errorf("error creating Discord session: %v", err)
	}

	session.Identify.Intents = discordgo.IntentsGuildMessages | discordgo.IntentsGuilds | discordgo.IntentsGuildMessageReactions

	err = session.Open()
	if err != nil {
		return fmt.Errorf("error opening Discord session: %v", err)
	}

	log.Println("Discord bot is running...")
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
		log.Printf("Error retrieving active threads: %v\n", err)
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

func LogStatus(jid, contactName, whatsappUsername, statusText string, colour *int) error {
	threadID, err := FindOrCreateForum(jid, contactName, whatsappUsername)
	if err != nil {
		return err
	}

	if colour == nil {
		colourDefault := 0xffbeef // cotton candy pink :3
		colour = &colourDefault
	}

	embed := &discordgo.MessageEmbed{
		Title:       "",
		Description: statusText,
		Color:       *colour,
	}

	session.ChannelEdit(threadID, &discordgo.ChannelEdit{
		Name: fmt.Sprintf("%s / %s / %s", jid, contactName, whatsappUsername),
	})

	_, err = session.ChannelMessageSendEmbed(threadID, embed)
	if err != nil {
		return fmt.Errorf("failed to send status message: %v", err)
	}

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


func LogStatusWithMedia(jid string, contactName string, whatsappUsername string, statusText string, mediaData []byte, mimetype string) error {
	threadID, err := FindOrCreateForum(jid, contactName, whatsappUsername)
	if err != nil {
		return err
	}

	// Update thread name if necessary
	_, err = session.ChannelEditComplex(threadID, &discordgo.ChannelEdit{
		Name: fmt.Sprintf("%s / %s / %s", jid, contactName, whatsappUsername),
	})
	if err != nil {
		log.Println("Couldn't update the name of the thread:", err)
	}

	embed := &discordgo.MessageEmbed{
		Title:       "",
		Description: statusText,
		Color:       0xffbeef,
	}

	file := &discordgo.File{
		Name:   "whatsapp-status-log",
		Reader: bytes.NewReader(mediaData),
	}

	msg := &discordgo.MessageSend{
		Embeds: []*discordgo.MessageEmbed{embed},
		Files:  []*discordgo.File{file},
	}

	fileExt := GetFileExtension(mimetype)
	file.Name = file.Name + fileExt

	if strings.HasPrefix(mimetype, "image/") {
		embed.Image = &discordgo.MessageEmbedImage{
			URL: "attachment://whatsapp-status-log." + fileExt,
		}
	} else if strings.HasPrefix(mimetype, "video/") {
		embed.Video = &discordgo.MessageEmbedVideo{
			URL: "attachment://whatsapp-status-log." + fileExt,
		}
	} else if strings.HasPrefix(mimetype, "audio/") {
		// No direct way of sending it
	}

	_, err = session.ChannelMessageSendComplex(threadID, msg)
	if err != nil {
		return fmt.Errorf("failed to send status message with media: %v", err)
	}

	return nil
}


func CloseBot() {
	if session != nil {
		log.Println("Shutting down the bot...")
		session.Close()
	}
}
