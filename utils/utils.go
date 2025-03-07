package utils

import (
	"context"
	"fmt"
	"mime"

    "net/http"
	"github.com/golang/protobuf/proto"
	"go.mau.fi/whatsmeow"
	waProto "go.mau.fi/whatsmeow/binary/proto"
	"go.mau.fi/whatsmeow/types/events"
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
        var quotedMessage *waProto.Message
        if quoted := messageEvent.Message.GetConversation(); quoted != "" {
            quotedMessage = &waProto.Message{
                ExtendedTextMessage: &waProto.ExtendedTextMessage{
                    Text: proto.String(message),
                    ContextInfo: &waProto.ContextInfo{
                        StanzaID:    proto.String(messageEvent.Info.ID),
                        Participant: proto.String(messageEvent.Info.Sender.String()),
                        QuotedMessage: messageEvent.Message,
                    },
                },
            }
        }

        if quotedMessage != nil {
            _, err := client.SendMessage(context.Background(), messageEvent.Info.Chat, quotedMessage)
            if err != nil {
                fmt.Printf("Failed to send '%s': %s", message, err)
                return false
            } else {
                return true
            }
        }
    } else {
        fmt.Println("Client is not initialized")
        return false
    }
    return false
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

	uploaded, err := client.Upload(context.Background(), imageData, whatsmeow.MediaImage)
	if err != nil {
		fmt.Printf("Failed to upload file: %v", err)
		return false
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
				StanzaID:    proto.String(messageEvent.Info.ID), // Reply to the original message
				Participant: proto.String(messageEvent.Info.Sender.String()),
				QuotedMessage: messageEvent.Message, // The message being replied to
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