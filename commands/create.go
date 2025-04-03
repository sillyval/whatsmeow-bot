package commands

import (
	"encoding/base64"
	"encoding/json"
    "io"
	"fmt"
	"net/http"
	"strings"
	"time"
	"whatsmeow-bot/utils"

	"go.mau.fi/whatsmeow"
	"go.mau.fi/whatsmeow/types/events"
)

type CreateCommand struct{
    secretConfig SecretConfig
}

func init() {
    secretConfig, err := LoadSecretConfig("secret-config.json")
    if err != nil {
        fmt.Println("Failed to load secret config:", err)
        return
    }
    RegisterCommand(&CreateCommand{secretConfig: *secretConfig})
}

func (c *CreateCommand) Execute(client *whatsmeow.Client, message *events.Message, args []string) *string {
    if utils.IsSystemMessage(message.Message) {
        return nil
    }

    utils.React(client, message, "⏳")

    if len(args) == 0 {
        utils.React(client, message, "❌")
        utils.Reply(client, message, "Please provide a prompt")
        return nil
    }

    newInput := strings.Join(args, " ")

    senderJID := message.Info.Sender.String()
    messageEntry := utils.MessageEntry{
        Content:        newInput,
        SenderJID:      senderJID,
        SenderPushName: message.Info.PushName,
        Timestamp:      message.Info.Timestamp,
    }

    quotedMessage := utils.GetQuotedMessage(message)
    var conversationJID *string
    if quotedMessage != nil {
        contextInfo := utils.GetMessageContextInfo(message.Message)
        conversationJID = contextInfo.StanzaID
    }

    messageEntry.QuotedJID = conversationJID

    err := utils.AppendToConversationChain(message.Info.ID, messageEntry)
    if err != nil {
        utils.Reply(client, message, fmt.Sprintf("Couldn't add prompt to context: %v", err))
        return nil
    }

    conversation, chainErr := utils.GetConversationChain(message.Info.ID)
    if chainErr != nil {
        utils.Reply(client, message, fmt.Sprintf("Couldn't retrieve conversation chain: %v", chainErr))
        return nil
    }

    byteData, imageBase64, apiErr := c.callDalle(conversation)

    if apiErr != nil {
        utils.React(client, message, "❌")
        utils.Reply(client, message, fmt.Sprintf("Could not reach Dalle: %v", apiErr))
        return nil
    }

    resp := utils.ReplyImageSystem(client, message, byteData, "image/png", "generated image")
    responseMessageEntry := utils.MessageEntry{
        Content:        "generated image.",
        SenderJID:      client.Store.ID.User,
        SenderPushName: "ValAI",
        Timestamp:      time.Now(),
        ImageBase64:    imageBase64,
    }
    responseMessageID := resp.ID

    responseMessageEntry.QuotedJID = &message.Info.ID
    err = utils.AppendToConversationChain(responseMessageID, responseMessageEntry)
    if err != nil {
        utils.Reply(client, message, fmt.Sprintf("Couldn't add AI message to context: %v", err))
        return nil
    }

    utils.React(client, message, "✅")

    return nil
}

func (c *CreateCommand) callDalle(contextChain []utils.MessageEntry) ([]byte, *string, error) {
    apiKey := c.secretConfig.OpenAIKey
    url := "https://api.openai.com/v1/images/generations"

    // Extract prompt from the conversation
    lastMessage := contextChain[len(contextChain)-1]
    prompt := lastMessage.Content

    requestBody := map[string]interface{}{
        "model":  "dall-e-3", // or "dall-e-3" based on desired model
        "prompt": prompt,
        "size":   "1024x1024",
        "n":      1,
    }

    jsonBody, err := json.Marshal(requestBody)
    if err != nil {
        return nil, nil, err
    }

    req, err := http.NewRequest("POST", url, strings.NewReader(string(jsonBody)))
    if err != nil {
        return nil, nil, err
    }

    req.Header.Set("Content-Type", "application/json")
    req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", apiKey))

    client := &http.Client{}
    resp, err := client.Do(req)
    if err != nil {
        return nil, nil, err
    }
    defer resp.Body.Close()

    var responseData map[string]interface{}
    err = json.NewDecoder(resp.Body).Decode(&responseData)
    if err != nil {
        return nil, nil, err
    }

    data, ok := responseData["data"].([]interface{})
    if !ok || len(data) == 0 {
        return nil, nil, fmt.Errorf("unexpected API response format")
    }

    imageData := data[0].(map[string]interface{})
    imageURL, urlOk := imageData["url"].(string)
    if !urlOk {
        return nil, nil, fmt.Errorf("unexpected image url format in response")
    }

    // Fetch the image data and encode it to Base64
    imgBytes, err := fetchImageBytes(imageURL)
    if err != nil {
        return nil, nil, fmt.Errorf("failed to fetch image bytes: %v", err)
    }

    base64Data := base64.StdEncoding.EncodeToString(imgBytes)

    return imgBytes, &base64Data, nil
}

// Fetch the image bytes from a given URL
func fetchImageBytes(imageURL string) ([]byte, error) {
    response, err := http.Get(imageURL)
    if err != nil {
        return nil, err
    }
    defer response.Body.Close()

    if response.StatusCode != http.StatusOK {
        return nil, fmt.Errorf("failed to fetch image, status code: %d", response.StatusCode)
    }

    imgData, err := io.ReadAll(response.Body)
    if err != nil {
        return nil, err
    }

    return imgData, nil
}

func (c *CreateCommand) Name() string {
    return "create,generate"
}

func (c *CreateCommand) Description() string {
    return "generate a Dalle image"
}
