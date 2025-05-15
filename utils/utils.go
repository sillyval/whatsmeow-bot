package utils

import (
	"bytes"
	"context"
	"encoding/base64"
	"fmt"
	"image"
	"image/png"
	"os/exec"
	"regexp"
	"strings"
	"time"

	"io"
	"net/http"

	"github.com/golang/protobuf/proto"
	"go.mau.fi/whatsmeow"
	waProto "go.mau.fi/whatsmeow/binary/proto"
	"go.mau.fi/whatsmeow/types"
	"go.mau.fi/whatsmeow/types/events"
)

func getLondon() *time.Location {
    result, _ := time.LoadLocation("Europe/London")
    return result
}

func React(client *whatsmeow.Client, messageEvent *events.Message, emoji string) bool {
	if len(emoji) == 0 {
		fmt.Println("Please provide an emoji to react with.")
		return false
	}

	if messageEvent == nil || messageEvent.Info.ID == "" {
		fmt.Println("Invalid message or message ID is empty.")
		return false
	}

	if emoji == "⏳" {
		return SendTyping(client, messageEvent.Info.Chat, nil)
	} else if emoji == "❌" || emoji == "✅" {
		return SendStopTyping(client, messageEvent.Info.Chat, nil)
	}

	reaction := client.BuildReaction(messageEvent.Info.Chat, messageEvent.Info.Sender, messageEvent.Info.ID, emoji)

	_, err := client.SendMessage(context.Background(), messageEvent.Info.Chat, reaction)
	if err != nil {
		fmt.Printf("Failed to send reaction: %v\n", err)
		return false
	} else {
		return true
	}
}

func SendMessageRead(client *whatsmeow.Client, chatJID types.JID, senderJID types.JID, messageJID types.MessageID) {
	londonLoc := getLondon()
	now := time.Now().In(londonLoc)
	client.MarkRead([]types.MessageID{messageJID}, now, chatJID, senderJID)
}
func SendAudioMessageRead(client *whatsmeow.Client, chatJID types.JID, senderJID types.JID, messageJID types.MessageID) {
	londonLoc := getLondon()
	now := time.Now().In(londonLoc)
	client.MarkRead([]types.MessageID{messageJID}, now, chatJID, senderJID, types.ReceiptTypePlayed)
}
func SendTyping(client *whatsmeow.Client, chatJID types.JID, isAudio *bool) bool {
	if client != nil {
		var prescenceMedia types.ChatPresenceMedia
		if isAudio != nil {
			if *isAudio {
				prescenceMedia = types.ChatPresenceMediaAudio
			} else {
				prescenceMedia = types.ChatPresenceMediaText
			}
		} else {
			prescenceMedia = types.ChatPresenceMediaText
		}
		err := client.SendChatPresence(chatJID, types.ChatPresenceComposing, prescenceMedia)
		if err != nil {
			fmt.Printf("Error sending typing prescence to %v: %v\n", chatJID, err)
		} else {
			return true
		}
	}
	return false
}
func SendStopTyping(client *whatsmeow.Client, chatJID types.JID, isAudio *bool) bool {
	if client != nil {
		var prescenceMedia types.ChatPresenceMedia
		if isAudio != nil {
			if *isAudio {
				prescenceMedia = types.ChatPresenceMediaAudio
			} else {
				prescenceMedia = types.ChatPresenceMediaText
			}
		} else {
			prescenceMedia = types.ChatPresenceMediaText
		}
		err := client.SendChatPresence(chatJID, types.ChatPresencePaused, prescenceMedia)
		if err != nil {
			fmt.Printf("Error sending stopped typing prescence to %v: %v\n", chatJID, err)
		} else {
			return true
		}
	}
	return false
}

func Message(client *whatsmeow.Client, messageEvent *events.Message, message string) bool {
	if client != nil {
		textMessage := &waProto.Message{
			Conversation: proto.String(message),
		}

		_, err := client.SendMessage(context.Background(), messageEvent.Info.Chat, textMessage)
		if err != nil {
			fmt.Printf("Failed to send '%s': %s", message, err)
			return false
		} else {
			return true
		}
	} else {
		fmt.Println("Client is not initialized")
		return false
	}
}
func SendMessageToJID(client *whatsmeow.Client, message *waProto.Message, jid types.JID) bool {
	if client != nil {
		_, err := client.SendMessage(context.Background(), jid, message)
		if err != nil {
			fmt.Printf("Failed to send '%s': %s", message, err)
			return false
		} else {
			return true
		}
	} else {
		fmt.Println("Client is not initialized")
		return false
	}
}
func Reply(client *whatsmeow.Client, messageEvent *events.Message, message string) *whatsmeow.SendResponse {
	if client != nil {
		quotedMessage := &waProto.Message{
			ExtendedTextMessage: &waProto.ExtendedTextMessage{
				Text: proto.String(message),
				ContextInfo: &waProto.ContextInfo{
					StanzaID:    proto.String(messageEvent.Info.ID),
					Participant: proto.String(messageEvent.Info.Sender.User+"@"+messageEvent.Info.Sender.Server),
					QuotedMessage: messageEvent.Message,
				},
				//BackgroundArgb: proto.Uint32(0xFF000000), // black
				//TextArgb: proto.Uint32(0xFFFFFFFF), // white
			},
		}

		resp, err := client.SendMessage(context.Background(), messageEvent.Info.Chat, quotedMessage)
		if err != nil {
			fmt.Printf("Failed to send '%s': %s", message, err)
			return nil
		} else {
			return &resp
		}
	} else {
		fmt.Println("Client is not initialized")
		return nil
	}
}
func MatchCaseReplace(input, search, replacement string) string {
	re := regexp.MustCompile("(?i)" + regexp.QuoteMeta(search))
	return re.ReplaceAllStringFunc(input, func(matched string) string {
		var result strings.Builder
		for i, c := range matched {
			if i >= len(replacement) {
				result.WriteRune(c)
				continue
			}
			r := rune(replacement[i])
			if 'A' <= c && c <= 'Z' {
				result.WriteRune(rune(strings.ToUpper(string(r))[0]))
			} else {
				result.WriteRune(rune(strings.ToLower(string(r))[0]))
			}
		}
		return result.String()
	})
}

func ReplySystem(client *whatsmeow.Client, messageEvent *events.Message, response string) *whatsmeow.SendResponse {
    systemResponse := "\u200B" + response
	systemResponse = MatchCaseReplace(systemResponse, "nigg", "nagg")
    return Reply(client, messageEvent, systemResponse)
}
func ReplyImageSystem(client *whatsmeow.Client, messageEvent *events.Message, imageData []byte, mimeType string, caption string) *whatsmeow.SendResponse {
    systemCaption := "\u200B<🤖> " + caption
    return ReplyImage(client, messageEvent, imageData, mimeType, systemCaption)
}
func IsSystemMessage(message *waProto.Message) bool {
    if message == nil {
        return false
    }
    messageContent := GetMessageBody(message)
    return strings.HasPrefix(messageContent, "\u200B") || strings.HasPrefix(messageContent, " ")
}
func IsGPTMessage(message *waProto.Message) bool {
    if message == nil {
        return false
    }
    messageContent := GetMessageBody(message)
    return strings.HasPrefix(messageContent, "\u200B")
}
func GetMessageBody(message *waProto.Message) string {
	var messageBody string
	if message.Conversation != nil {
		messageBody = *message.Conversation
	} else if message.ExtendedTextMessage != nil {
		messageBody = *message.ExtendedTextMessage.Text
	} else if message.ImageMessage != nil && message.ImageMessage.Caption != nil {
		messageBody = *message.ImageMessage.Caption
	} else if message.VideoMessage != nil && message.VideoMessage.Caption != nil {
		messageBody = *message.VideoMessage.Caption
	} else if message.DocumentMessage != nil && message.DocumentMessage.Caption != nil {
		messageBody = *message.DocumentMessage.Caption
	}
	return messageBody
}
func GetMessageStanza(message *waProto.Message) types.MessageID {
	var messageStanza types.MessageID
	if message.ExtendedTextMessage != nil {
		messageStanza = *message.GetExtendedTextMessage().ContextInfo.StanzaID
	} else if message.ImageMessage != nil {
		messageStanza =  *message.GetImageMessage().ContextInfo.StanzaID
	} else if message.VideoMessage != nil {
		messageStanza = *message.GetVideoMessage().ContextInfo.StanzaID
	} else if message.DocumentMessage != nil {
		messageStanza = *message.GetDocumentMessage().ContextInfo.StanzaID
	}
	return messageStanza
}
func GetMessageContextInfo(message *waProto.Message) *waProto.ContextInfo {
	if message.ExtendedTextMessage != nil {
		return message.ExtendedTextMessage.GetContextInfo()
	} else if message.ImageMessage != nil && message.ImageMessage.Caption != nil {
		return message.ImageMessage.GetContextInfo()
	} else if message.VideoMessage != nil && message.VideoMessage.Caption != nil {
		return message.VideoMessage.GetContextInfo()
	}
	return nil
}
func ReplyMessage(client *whatsmeow.Client, chatJID types.JID, message *waProto.Message, contextInfo *waProto.ContextInfo, reply string) bool {
	if client != nil {
		quotedMessage := &waProto.Message{
			ExtendedTextMessage: &waProto.ExtendedTextMessage{
				Text: &reply,
				ContextInfo: contextInfo,
			},
		}

		_, err := client.SendMessage(context.Background(), chatJID, quotedMessage)
		if err != nil {
			fmt.Printf("Failed to send '%s': %s", message, err)
			return false
		} else {
			return true
		}
	} else {
		fmt.Println("Client is not initialized")
		return false
	}
}
func GetImageMessageFromData(client *whatsmeow.Client, imageData[]byte, caption string) *waProto.Message {
	uploaded, err := client.Upload(context.Background(), imageData, whatsmeow.MediaImage)
	if err != nil {
		fmt.Printf("Failed to upload file: %v", err)
		return nil
	}
	message := &waProto.Message{ImageMessage: &waProto.ImageMessage{
		Caption:       proto.String(caption),
		URL:           proto.String(uploaded.URL),
		DirectPath:    proto.String(uploaded.DirectPath),
		MediaKey:      uploaded.MediaKey,
		Mimetype:      proto.String(http.DetectContentType(imageData)),
		FileEncSHA256: uploaded.FileEncSHA256,
		FileSHA256:    uploaded.FileSHA256,
		FileLength:    proto.Uint64(uint64(len(imageData))),
	}}
	return message
}
func GetMediaFromMessage(client *whatsmeow.Client, message *waProto.Message) ([]byte, string, error) {
	// Ensure that the client and message are not nil
	if client == nil || message == nil {
		return nil, "", fmt.Errorf("invalid client or message")
	}

	// Handle Image Message
	imgMsg := message.GetImageMessage()
	if imgMsg != nil {
		data, err := client.Download(imgMsg)
		if err != nil {
			return nil, "", fmt.Errorf("failed to download image: %v", err)
		}
		mimeType := imgMsg.GetMimetype()

		return data, mimeType, nil
	}

	// Handle Audio Message
	if audioMsg := message.GetAudioMessage(); audioMsg != nil {
		data, err := client.Download(audioMsg)
		if err != nil {
			return nil, "", fmt.Errorf("failed to download audio: %v", err)
		}
		mimeType := audioMsg.GetMimetype()
		// exts, _ := mime.ExtensionsByType(audioMsg.GetMimetype())
		// if len(exts) == 0 {
		// 	return nil, "", fmt.Errorf("could not determine extension for audio mimetype %s", audioMsg.GetMimetype())
		// }
		return data, mimeType, nil
	}

	// Handle Video Message
	if videoMsg := message.GetVideoMessage(); videoMsg != nil {
		data, err := client.Download(videoMsg)
		if err != nil {
			return nil, "", fmt.Errorf("failed to download video: %v", err)
		}
		mimeType := videoMsg.GetMimetype()
		// exts, _ := mime.ExtensionsByType(videoMsg.GetMimetype())
		// if len(exts) == 0 {
		// 	return nil, "", fmt.Errorf("could not determine extension for video mimetype %s", videoMsg.GetMimetype())
		// }
		return data, mimeType, nil
	}

	// Handle Document Message
	if docMsg := message.GetDocumentMessage(); docMsg != nil {
		fmt.Printf("Document message!")
		data, err := client.Download(docMsg)
		if err != nil {
			return nil, "", fmt.Errorf("failed to download document: %v", err)
		}
		mimeType := docMsg.GetMimetype()

		return data, mimeType, nil
	}

	return nil, "", fmt.Errorf("no media found in message")
}
func DownloadAndEncodeImage(client *whatsmeow.Client, message *waProto.Message) (*string, error) {
    mediaData, mimeType, err := GetMediaFromMessage(client, message)

	if mediaData == nil {
		return nil, fmt.Errorf("no media found")
	}

    if err != nil {
        return nil, fmt.Errorf("failed to download media: %v", err)
    }

    if strings.HasPrefix(mimeType, "image/") {
        base64Data := base64.StdEncoding.EncodeToString(mediaData)
        return &base64Data, nil
    }

    return nil, nil
}
func GetImageFromMessage(client *whatsmeow.Client, imageMessage *waProto.ImageMessage) ([]byte, string, error) {
	// Ensure that the client and message are not nil
	if client == nil || imageMessage == nil {
		return nil, "", fmt.Errorf("invalid client or message")
	}

	data, err := client.Download(imageMessage)
	if err != nil {
		return nil, "", fmt.Errorf("failed to download image: %v", err)
	}
	mimeType := imageMessage.GetMimetype()

	return data, mimeType, nil
}
func ConvertToPNG(mediaData []byte) []byte {
	img, _, err := image.Decode(bytes.NewReader(mediaData))
	if err != nil {
		fmt.Printf("\nfailed to decode image: %v\n", err)
		return []byte{ }
	}

	var pngBuffer bytes.Buffer
	err = png.Encode(&pngBuffer, img)
	if err != nil {
		fmt.Printf("\nfailed to encode image to PNG: %v\n", err)
		return []byte{ }
	}

	return pngBuffer.Bytes()
}
func ConvertToMP4(mediaData []byte) []byte {
	cmd := exec.Command("ffmpeg",
		"-y",                           // Overwrite output files without asking
		"-i", "pipe:0",                 // Input from stdin
		"-movflags", "frag_keyframe+empty_moov", // Fix for non-seekable output
		"-pix_fmt", "yuv420p",          // Ensure compatibility
		"-vf", "scale=trunc(iw/2)*2:trunc(ih/2)*2", // Ensure even dimensions
		"-c:v", "libx264",              // Use H.264 video codec
		"-preset", "fast",              // Encoding preset
		"-crf", "23",                   // Quality level (23 is default)
		"-c:a", "aac",                  // Audio codec
		"-b:a", "128k",                 // Audio bitrate
		"-f", "mp4",                    // Output format
		"pipe:1")                       // Output to stdout

	cmd.Stdin = bytes.NewReader(mediaData)
	var output bytes.Buffer
	cmd.Stdout = &output
	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		fmt.Printf("\nFFmpeg error: %v, details: %s\n", err, stderr.String())
		return nil
	}

	return output.Bytes()
}
func ConvertToMP3(mediaData []byte) []byte {
	cmd := exec.Command("ffmpeg",
		"-y",                   // Overwrite output files without asking
		"-i", "pipe:0",         // Input comes from stdin
		"-vn",                  // Disable video (audio-only output)
		"-c:a", "libmp3lame",   // Use LAME for MP3 encoding
		"-q:a", "2",            // VBR quality (~190-220 kbps, better than CBR 192k)
		"-ar", "44100",         // Set audio sampling rate to 44.1kHz
		"-ac", "2",             // Set stereo audio
		"-f", "mp3",            // Output format
		"pipe:1")               // Output to stdout

	// Set stdin and stdout
	cmd.Stdin = bytes.NewReader(mediaData)
	var output bytes.Buffer
	cmd.Stdout = &output
	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	// Execute FFmpeg command
	if err := cmd.Run(); err != nil {
		fmt.Printf("\nFFmpeg error: %v, details: %s\n", err, stderr.String())
		return nil
	}

	return output.Bytes()
}
func DownloadMediaFromURL(mediaURL string) ([]byte, string, error) {
	resp, err := http.Get(mediaURL)
	if err != nil {
		return nil, "", fmt.Errorf("error downloading media: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, "", fmt.Errorf("failed to download media: HTTP %d", resp.StatusCode)
	}

	mimeType := resp.Header.Get("Content-Type")
	media, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, "", fmt.Errorf("failed to read media content: %v", err)
	}

	return media, mimeType, nil
}
func GetStickerFromMessage(client *whatsmeow.Client, stickerMessage *waProto.StickerMessage) ([]byte, string, error) {
	if client == nil || stickerMessage == nil {
		return nil, "", fmt.Errorf("invalid client or sticker message")
	}
	mimeType := "image/webp"

	mediaKey := stickerMessage.GetMediaKey()
	mediaEncSHA256 := stickerMessage.GetFileEncSHA256()
	mediaSHA256 := stickerMessage.GetFileSHA256()
	mediaURL := stickerMessage.GetURL()
	mediaDirectPath := stickerMessage.GetDirectPath()

	imageMessage := &waProto.ImageMessage{
		Caption:       proto.String(""),
		URL:           proto.String(mediaURL),
		DirectPath:    proto.String(mediaDirectPath),
		MediaKey:      mediaKey,
		Mimetype:      proto.String(mimeType),
		FileEncSHA256: mediaEncSHA256,
		FileSHA256:    mediaSHA256,
		//FileLength:    proto.Uint64(uint64(len(imageData))),
	}

	return GetImageFromMessage(client, imageMessage)
}
func SendImage(client *whatsmeow.Client, messageEvent *events.Message, imageData []byte, mimeType string, caption string) bool {
	if client == nil {
		fmt.Println("Client is not initialized")
		return false
	}

	var err error
	message := GetImageMessageFromData(client, imageData, caption)

	_, err = client.SendMessage(context.Background(), messageEvent.Info.Chat, message)
	if err != nil {
		fmt.Printf("Error sending image message: %v", err)
		return false
	} else {
		return true
	}

}
func ReplyImage(client *whatsmeow.Client, messageEvent *events.Message, imageData []byte, mimeType string, caption string) *whatsmeow.SendResponse {
	if client == nil {
		fmt.Println("Client is not initialized")
		return nil
	}

	uploaded, err := client.Upload(context.Background(), imageData, whatsmeow.MediaImage)
	if err != nil {
		fmt.Printf("Failed to upload file: %v\n", err)
		return nil
	}

	imageMessage := &waProto.Message{
		ImageMessage: &waProto.ImageMessage{
			Caption:       proto.String(caption),
			URL:           proto.String(uploaded.URL),
			DirectPath:    proto.String(uploaded.DirectPath),
			MediaKey:      uploaded.MediaKey,
			Mimetype:      proto.String(mimeType),
			FileEncSHA256: uploaded.FileEncSHA256,
			FileSHA256:    uploaded.FileSHA256,
			FileLength:    proto.Uint64(uint64(len(imageData))),
			ContextInfo: &waProto.ContextInfo{
				StanzaID:    proto.String(messageEvent.Info.ID),
				Participant: proto.String(messageEvent.Info.Sender.String()),
				QuotedMessage: messageEvent.Message,
			},
		},
	}

	resp, err := client.SendMessage(context.Background(), messageEvent.Info.Chat, imageMessage)
	if err != nil {
		fmt.Printf("Error sending image reply: %v\n", err)
		return nil
	}

	return &resp
}
func ReplyImageMessage(client *whatsmeow.Client, messageEvent *events.Message, imageMessage *waProto.Message) bool {
	if client == nil {
		fmt.Println("Client is not initialized")
		return false
	}

	imageMessage.ImageMessage.ContextInfo = &waProto.ContextInfo{
		StanzaID:    proto.String(messageEvent.Info.ID),
		Participant: proto.String(messageEvent.Info.Sender.String()),
		QuotedMessage: messageEvent.Message,
	}

	_, err := client.SendMessage(context.Background(), messageEvent.Info.Chat, imageMessage)
	if err != nil {
		fmt.Printf("Error sending image reply: %v\n", err)
		return false
	}

	return true
}
func ReplyImageToQuoted(client *whatsmeow.Client, message *events.Message, quotedMsgID string, imageData []byte, mimeType string, caption string) bool {
	jid := message.Info.Chat

	uploaded, err := client.Upload(context.Background(), imageData, whatsmeow.MediaImage)
	if err != nil {
		return false
	}

	msg := &waProto.Message{
		ImageMessage: &waProto.ImageMessage{
			URL:        proto.String(uploaded.URL),
			Mimetype:   proto.String(mimeType),
			Caption:    proto.String(caption),
			DirectPath: proto.String(uploaded.DirectPath),
			MediaKey:   uploaded.MediaKey,
			FileEncSHA256: uploaded.FileEncSHA256,
			FileSHA256: uploaded.FileSHA256,
			FileLength: proto.Uint64(uint64(len(imageData))),
			ContextInfo: &waProto.ContextInfo{
				StanzaID:  proto.String(quotedMsgID),
				Participant: proto.String(message.Info.Sender.String()),
			},
		},
	}

	_, err = client.SendMessage(context.Background(), jid, msg)
	return err == nil
}
func ReplyImageToChat(client *whatsmeow.Client, message *events.Message, jid types.JID, quotedMsgID string, imageData []byte, mimeType string, caption string) bool {
	uploaded, err := client.Upload(context.Background(), imageData, whatsmeow.MediaImage)
	if err != nil {
		return false
	}

	msg := &waProto.Message{
		ImageMessage: &waProto.ImageMessage{
			URL:        proto.String(uploaded.URL),
			Mimetype:   proto.String(mimeType),
			Caption:    proto.String(caption),
			DirectPath: proto.String(uploaded.DirectPath),
			MediaKey:   uploaded.MediaKey,
			FileEncSHA256: uploaded.FileEncSHA256,
			FileSHA256: uploaded.FileSHA256,
			FileLength: proto.Uint64(uint64(len(imageData))),
			ContextInfo: &waProto.ContextInfo{
				StanzaID:  proto.String(quotedMsgID),
				Participant: proto.String(message.Info.Sender.String()),
			},
		},
	}

	_, err = client.SendMessage(context.Background(), jid, msg)
	return err == nil
}
func SendMessageWithMentions(client *whatsmeow.Client, chatJID types.JID, message string, mentions []string) bool {
	if client == nil {
		fmt.Println("Client is not initialized")
		return false
	}

	var err error

	// chatJID, err := types.ParseJID(chatJIDString)
	// if err != nil {
	//     fmt.Println("Invalid JID format")
	//     return false
	// }

	var mentionedJIDs []string
	for _, mention := range mentions {
		if !strings.HasSuffix(mention, "@s.whatsapp.net") {
			mention = mention + "@s.whatsapp.net"
		}
		mentionedJIDs = append(mentionedJIDs, mention)

		atStr := "@" + mention[:strings.Index(mention, "@")]
		message = strings.ReplaceAll(message, "@"+mention[:strings.Index(mention, "@")], atStr)
	}

	contextInfo := &waProto.ContextInfo{
		MentionedJID: mentionedJIDs,
	}

	textMessage := &waProto.Message{
		ExtendedTextMessage: &waProto.ExtendedTextMessage{
			Text:        proto.String(message),
			ContextInfo: contextInfo,
		},
	}

	_, err = client.SendMessage(context.Background(), chatJID, textMessage)
	if err != nil {
		fmt.Printf("Failed to send message with mentions: %v\n", err)
		return false
	}

	return true
}
func IsJIDMentioned(msg *events.Message, targetJID string) bool {

	number := strings.Split(targetJID, "@")[0]
	if strings.Contains(GetMessageBody(msg.Message), number) {
		return true
	}

    extendedTextMsg := msg.Message.GetExtendedTextMessage()
    if extendedTextMsg == nil {
        return false
    }

    contextInfo := extendedTextMsg.GetContextInfo()
    if contextInfo == nil {
        return false
    }

    for _, mentionedJID := range contextInfo.GetMentionedJID() {

		//fmt.Println("Mention:",mentionedJID)
		//fmt.Println("Looking to match:",targetJID)
	
        if mentionedJID == targetJID {
            return true
        }
    }

	return false
}
func IsDMS(message *events.Message) bool {
	return !message.Info.IsGroup
}
func SendExtendedMessageWithMentions(client *whatsmeow.Client, chatJID types.JID, message *waProto.Message, mentions []string) *whatsmeow.SendResponse {
	if client == nil {
		fmt.Println("Client is not initialized")
		return nil
	}

	var err error

	var mentionedJIDs []string
	for _, mention := range mentions {
		if !strings.HasSuffix(mention, "@s.whatsapp.net") {
			mention = mention + "@s.whatsapp.net"
		}
		mentionedJIDs = append(mentionedJIDs, mention)
	}

	contextInfo := message.ExtendedTextMessage.GetContextInfo();
	if contextInfo == nil {
		contextInfo = &waProto.ContextInfo{
			MentionedJID: mentionedJIDs,
		}
	}
	contextInfo.MentionedJID = mentionedJIDs

	message.ExtendedTextMessage.ContextInfo = contextInfo;

	fmt.Println(message.ExtendedTextMessage.ContextInfo);

	response, err := client.SendMessage(context.Background(), chatJID, message)
	if err != nil {
		fmt.Printf("Failed to send message with mentions: %v\n", err)
		return nil
	}

	return &response
}
func GetQuotedMessage(message *events.Message) *waProto.Message {
    if message.Message.ExtendedTextMessage != nil && message.Message.ExtendedTextMessage.ContextInfo != nil {
        quotedMessage := message.Message.ExtendedTextMessage.ContextInfo.QuotedMessage
        if quotedMessage != nil {
            return quotedMessage
        }
    }
	if message.Message.ImageMessage != nil && message.Message.ImageMessage.ContextInfo != nil {
        quotedMessage := message.Message.ImageMessage.ContextInfo.QuotedMessage
        if quotedMessage != nil {
            return quotedMessage
        }
    }
	if message.Message.VideoMessage != nil && message.Message.VideoMessage.ContextInfo != nil {
        quotedMessage := message.Message.VideoMessage.ContextInfo.QuotedMessage
        if quotedMessage != nil {
            return quotedMessage
        }
    }
	if message.Message.AudioMessage != nil && message.Message.AudioMessage.ContextInfo != nil {
        quotedMessage := message.Message.AudioMessage.ContextInfo.QuotedMessage
        if quotedMessage != nil {
            return quotedMessage
        }
    }
    return nil
}
func GetReplyChain(message *events.Message) []string {
    var chain []string
    currentMessage := message

    for currentMessage != nil {
        var messageContent string

        if currentMessage.Message.Conversation != nil {
            messageContent = *currentMessage.Message.Conversation
        } else if currentMessage.Message.ExtendedTextMessage != nil {
            messageContent = *currentMessage.Message.ExtendedTextMessage.Text
        }

        if messageContent != "" {
            chain = append([]string{messageContent}, chain...)
        }

        if currentMessage.Message.ExtendedTextMessage != nil && currentMessage.Message.ExtendedTextMessage.ContextInfo != nil {
            quotedMessage := currentMessage.Message.ExtendedTextMessage.ContextInfo.QuotedMessage
            if quotedMessage != nil {
                currentMessage = &events.Message{
                    Info:    currentMessage.Info,
                    Message: quotedMessage,
                }
            } else {
                break
            }
        } else {
            break
        }
    }

	fmt.Println(chain)

    return chain
}
func GetContactName(client *whatsmeow.Client, jid types.JID) (string, bool) {

	contact, err := client.Store.Contacts.GetContact(jid)
	if err == nil {
		if contact.FullName != "" {
			return contact.FullName, true
		} else if contact.PushName != "" {
			return contact.PushName, false
		} else {
			return jid.String(), false
		}
	} else {
		return jid.String(), false
	}
}
func GetPushName(client *whatsmeow.Client, jid types.JID) (string, bool) {
	contact, err := client.Store.Contacts.GetContact(jid)
	if err == nil {
		if contact.PushName != "" {
			return contact.PushName, true
		} else if contact.FullName != "" {
			return contact.FullName, false
		} else {
			return jid.String(), false
		}
	} else {
		return jid.String(), false
	}
}
func IsStatusPost(message *events.Message) bool {
	return message.Info.Chat.String() == "status@broadcast"
}
func IsDeletedMessage(msg *events.Message) bool {
	if msg.Message.GetProtocolMessage() != nil {
		protoMsg := msg.Message.GetProtocolMessage()
		if *protoMsg.Type == waProto.ProtocolMessage_REVOKE {
			return true
		}
	}
	return false
}
func GetDeletedMessageID(msg *events.Message) *string {
	if msg.Message.GetProtocolMessage() != nil {
		protoMsg := msg.Message.GetProtocolMessage()
		if *protoMsg.Type == waProto.ProtocolMessage_REVOKE {
			return protoMsg.Key.ID
		}
	}
	return nil
}
func IsIncomingMessage(msg *events.Message) bool {
	return msg.Message.GetProtocolMessage() == nil
}
func IsEditedMessage(msg *events.Message) bool {
	if msg.Message.GetProtocolMessage() != nil {
		protoMsg := msg.Message.GetProtocolMessage()
		if *protoMsg.Type == waProto.ProtocolMessage_MESSAGE_EDIT {
			return true
		}
	}
	return false
}
func StripJID(jid types.JID) types.JID {
	localPart := jid.User
	domainPart := jid.Server

	localParts := strings.SplitN(localPart, ":", 2)
	phoneNumber := localParts[0]

	return types.NewJID(phoneNumber, domainPart)
}

func IsNewsletter(msg *events.Message) bool {
	return msg.Info.Chat.Server == "newsletter"
}
func NewsletterMessage(client *whatsmeow.Client, jid types.JID, messageText string) error {
	if client != nil {
		textMessage := &waProto.Message{
			Conversation: proto.String(messageText),
		}

		_, err := client.SendMessage(context.Background(), jid, textMessage, whatsmeow.SendRequestExtra{
			ID: client.GenerateMessageID(),
		})
		if err != nil {
			return fmt.Errorf("failed to send '%s': %s", textMessage, err)
		} else {
			return nil
		}
	} else {
		fmt.Println("Client is not initialized")
		return fmt.Errorf("client is not initialized")
	}
}
func NewsletterSendImage(client *whatsmeow.Client, jid types.JID, caption string, imageData []byte) error {
	if client == nil {
		return fmt.Errorf("WhatsMeow client is nil")
	}

	resp, err := client.UploadNewsletter(context.Background(), imageData, whatsmeow.MediaImage)

	if err != nil {
		return fmt.Errorf("failed to upload newsletter message: %w", err)
	}

	imageMsg := &waProto.ImageMessage{
		Caption:  proto.String(caption),
		Mimetype: proto.String("image/jpeg"),
		URL:        &resp.URL,
		DirectPath: &resp.DirectPath,
		FileSHA256: resp.FileSHA256,
		FileLength: &resp.FileLength,
		// Newsletter media isn't encrypted, so the media key and file enc sha fields are not applicable
	}
	_, err = client.SendMessage(context.Background(), jid, &waProto.Message{
		ImageMessage: imageMsg,
	}, whatsmeow.SendRequestExtra{
		// Unlike normal media, newsletters also include a "media handle" in the send request.
		MediaHandle: resp.Handle,
	})

	if err != nil {
		return fmt.Errorf("failed to send newsletter message: %w", err)
	}

	return nil
}
func NewsletterSendImageMessage(client *whatsmeow.Client, jid types.JID, caption string, imageMessage *waProto.ImageMessage, handle string) error {
	if client == nil {
		return fmt.Errorf("WhatsMeow client is nil")
	}

	_, err := client.SendMessage(context.Background(), jid, &waProto.Message{
		ImageMessage: imageMessage,
	}, whatsmeow.SendRequestExtra{
		// Unlike normal media, newsletters also include a "media handle" in the send request.
		MediaHandle: handle,
	})

	if err != nil {
		return fmt.Errorf("failed to send newsletter message: %w", err)
	}

	return nil
}
func NewsletterReact(client *whatsmeow.Client, messageEvent *events.Message, emoji string) bool {
	if len(emoji) == 0 {
		fmt.Println("Please provide an emoji to react with.")
		return false
	}

	if messageEvent == nil || messageEvent.Info.ID == "" {
		fmt.Println("Invalid message or message ID is empty.")
		return false
	}

	err := client.NewsletterSendReaction(messageEvent.Info.Chat, messageEvent.Info.ServerID, emoji, client.GenerateMessageID())
	if err != nil {
		fmt.Printf("Failed to send reaction: %v\n", err)
		return false
	} else {
		return true
	}
}