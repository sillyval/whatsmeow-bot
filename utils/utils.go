package utils

import (
	"context"
	"fmt"
	"mime"
	"strings"

	"io"
	"net/http"

	"github.com/golang/protobuf/proto"
	"go.mau.fi/whatsmeow"
	waProto "go.mau.fi/whatsmeow/binary/proto"
	"go.mau.fi/whatsmeow/types/events"
	"go.mau.fi/whatsmeow/types"
)

func React(client *whatsmeow.Client, messageEvent *events.Message, emoji string) bool {
	if len(emoji) == 0 {
		fmt.Println("Please provide an emoji to react with.")
		return false
	}

	if messageEvent == nil || messageEvent.Info.ID == "" {
		fmt.Println("Invalid message or message ID is empty.")
		return false
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
func Reply(client *whatsmeow.Client, messageEvent *events.Message, message string) bool {
	if client != nil {
		quotedMessage := &waProto.Message{
			ExtendedTextMessage: &waProto.ExtendedTextMessage{
				Text: proto.String(message),
				ContextInfo: &waProto.ContextInfo{
					StanzaID:    proto.String(messageEvent.Info.ID),
					Participant: proto.String(messageEvent.Info.Sender.String()),
					QuotedMessage: messageEvent.Message,
				},
			},
		}

		_, err := client.SendMessage(context.Background(), messageEvent.Info.Chat, quotedMessage)
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
	}
	return messageBody
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
func GetMediaFromMessage(client *whatsmeow.Client, message *events.Message) ([]byte, string, error) {
	// Ensure that the client and message are not nil
	if client == nil || message == nil {
		return nil, "", fmt.Errorf("invalid client or message")
	}

	// Handle Image Message
	if imgMsg := message.Message.GetImageMessage(); imgMsg != nil {
		data, err := client.Download(imgMsg)
		if err != nil {
			return nil, "", fmt.Errorf("failed to download image: %v", err)
		}
		mimeType := imgMsg.GetMimetype()

		return data, mimeType, nil
	}

	// Handle Audio Message
	if audioMsg := message.Message.GetAudioMessage(); audioMsg != nil {
		data, err := client.Download(audioMsg)
		if err != nil {
			return nil, "", fmt.Errorf("failed to download audio: %v", err)
		}
		exts, _ := mime.ExtensionsByType(audioMsg.GetMimetype())
		if len(exts) == 0 {
			return nil, "", fmt.Errorf("could not determine extension for audio mimetype %s", audioMsg.GetMimetype())
		}
		return data, exts[0], nil
	}

	// Handle Video Message
	if videoMsg := message.Message.GetVideoMessage(); videoMsg != nil {
		data, err := client.Download(videoMsg)
		if err != nil {
			return nil, "", fmt.Errorf("failed to download video: %v", err)
		}
		exts, _ := mime.ExtensionsByType(videoMsg.GetMimetype())
		if len(exts) == 0 {
			return nil, "", fmt.Errorf("could not determine extension for video mimetype %s", videoMsg.GetMimetype())
		}
		return data, exts[0], nil
	}

	// Handle Document Message
	if docMsg := message.Message.GetDocumentMessage(); docMsg != nil {
		data, err := client.Download(docMsg)
		if err != nil {
			return nil, "", fmt.Errorf("failed to download document: %v", err)
		}
		exts, _ := mime.ExtensionsByType(docMsg.GetMimetype())
		if len(exts) == 0 {
			return nil, "", fmt.Errorf("could not determine extension for document mimetype %s", docMsg.GetMimetype())
		}
		return data, exts[0], nil
	}

	return nil, "", fmt.Errorf("no media found in message")
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
func ReplyImage(client *whatsmeow.Client, messageEvent *events.Message, imageData []byte, mimeType string, caption string) bool {
	if client == nil {
		fmt.Println("Client is not initialized")
		return false
	}

	uploaded, err := client.Upload(context.Background(), imageData, whatsmeow.MediaImage)
	if err != nil {
		fmt.Printf("Failed to upload file: %v\n", err)
		return false
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

	_, err = client.SendMessage(context.Background(), messageEvent.Info.Chat, imageMessage)
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