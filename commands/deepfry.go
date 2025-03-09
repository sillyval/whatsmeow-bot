package commands

import (
	"bytes"
	"fmt"
	"image"
	"image/jpeg"
	"image/png"
    "net/url"

	"github.com/disintegration/imaging"
	"go.mau.fi/whatsmeow"
	waProto "go.mau.fi/whatsmeow/binary/proto"
	"go.mau.fi/whatsmeow/types/events"

	"whatsmeow-bot/utils"
)

func init() {
    RegisterCommand(&DeepfryCommand{})
}

type DeepfryCommand struct{}

func deepfryImage(img image.Image) ([]byte, error) {
    img = imaging.AdjustContrast(img, 80)
    img = imaging.AdjustSaturation(img, 100)
    
    var buf bytes.Buffer
    err := jpeg.Encode(&buf, img, &jpeg.Options{Quality: 1})
    if err != nil {
        return nil, err
    }
    return buf.Bytes(), nil
}
func isValidURL(link string) bool {
    _, err := url.ParseRequestURI(link)
    return err == nil
}

func (c *DeepfryCommand) Execute(client *whatsmeow.Client, message *events.Message, args []string) {

	var imageMessage *waProto.ImageMessage
	var stickerMessage *waProto.StickerMessage

	// Check if the message contains an image or a sticker
	if message.Message.ImageMessage != nil {
		imageMessage = message.Message.ImageMessage
	} else if message.Message.StickerMessage != nil {
		stickerMessage = message.Message.StickerMessage
	} else if len(args) > 0 && isValidURL(args[0]) {
		// handled later on
	} else {
		
		// Check for quoted messages that might contain an image or sticker
		if message.Message.ExtendedTextMessage == nil {
			utils.Reply(client, message, "No image or sticker found")
			utils.React(client, message, "❌")
			return
		}
	
		contextInfo := message.Message.ExtendedTextMessage.GetContextInfo()
		if contextInfo == nil {
			utils.Reply(client, message, "No image or sticker found")
			utils.React(client, message, "❌")
			return
		}
	
		if contextInfo.QuotedMessage == nil {
			utils.Reply(client, message, "No image or sticker found")
			utils.React(client, message, "❌")
			return
		}

		// Check if quoted message contains an image or sticker
		if contextInfo.QuotedMessage.ImageMessage != nil {
			imageMessage = contextInfo.QuotedMessage.GetImageMessage()
		} else if contextInfo.QuotedMessage.StickerMessage != nil {
			stickerMessage = contextInfo.QuotedMessage.GetStickerMessage()
		}
	}

	var media []byte
	var mimetype string
	var err error

	// Handle image or sticker extraction
	if imageMessage != nil {
		media, mimetype, err = utils.GetImageFromMessage(client, imageMessage)
	} else if stickerMessage != nil {
		media, mimetype, err = utils.GetStickerFromMessage(client, stickerMessage)
	} else if len(args) > 0 && isValidURL(args[0]) {
        media, mimetype, err = utils.DownloadMediaFromURL(args[0])
	} else {
		fmt.Printf("url: %s, is valid? %t", args[0], isValidURL(args[0]))
		utils.Reply(client, message, "No image or sticker found")
		utils.React(client, message, "❌")
		return
	}

	if err != nil {
		fmt.Println(err)
		utils.Reply(client, message, "Failed to download media.")
		utils.React(client, message, "❌")
		return
	}

	// Decode image (whether it is an image or sticker, we decode it the same way)
	img, format, err := image.Decode(bytes.NewReader(media))
	if err != nil {
		fmt.Println(err)
		utils.Reply(client, message, "Unsupported image format.")
		utils.React(client, message, "❌")
		return
	}

	// Convert WebP stickers to PNG if needed
	if format == "webp" {
		buf := new(bytes.Buffer)
		err = png.Encode(buf, img)
		if err != nil {
			fmt.Println(err)
			utils.Reply(client, message, "Failed to convert to PNG.")
			utils.React(client, message, "❌")
			return
		}
		img, _, err = image.Decode(buf)
		if err != nil {
			utils.Reply(client, message, "Failed to process image.")
			utils.React(client, message, "❌")
			return
		}
	}

	// Apply deep fry effect to the image (either sticker or image)
	deepfriedImage, err := deepfryImage(img)
	if err != nil {
		utils.Reply(client, message, "Failed to process image.")
		utils.React(client, message, "❌")
		return
	}

	success := utils.ReplyImage(client, message, deepfriedImage, mimetype, "")
	if success {
		utils.React(client, message, "✅")
	} else {
		utils.React(client, message, "❌")
	}
}

func (c *DeepfryCommand) Name() string {
    return "deepfry"
}

func (c *DeepfryCommand) Description() string {
    return "Deepfries a given image"
}