package utils

import (
	"encoding/json"
	"fmt"
	"os"
	"sync"
	"time"

	"go.mau.fi/whatsmeow"
)

type MessageEntry struct {
    Content       string    `json:"content"`
    SenderPushName string   `json:"sender_push_name"`
    SenderJID     string    `json:"sender_jid"`
    QuotedJID     *string   `json:"quoted_jid"`
    Timestamp     time.Time `json:"timestamp"`
    ImageBase64   *string   `json:"image_base64,omitempty"`
}

var (
    chainFilePath  = "conversation_messages.json"
    mutex          = &sync.Mutex{}
    expiryDuration = 24 * time.Hour
)

// LoadMessages reads the conversation messages from the JSON file.
func LoadMessages() (map[string]MessageEntry, error) {
    mutex.Lock()
    defer mutex.Unlock()

    file, err := os.Open(chainFilePath)
    if err != nil {
        if os.IsNotExist(err) {
            return make(map[string]MessageEntry), nil
        }
        return nil, err
    }
    defer file.Close()

    var messages map[string]MessageEntry
    decoder := json.NewDecoder(file)
    if err := decoder.Decode(&messages); err != nil {
        return nil, err
    }

    return messages, nil
}

// SaveMessages writes the conversation messages to the JSON file.
func SaveMessages(messages map[string]MessageEntry) error {
    mutex.Lock()
    defer mutex.Unlock()

    file, err := os.Create(chainFilePath)
    if err != nil {
        return err
    }
    defer file.Close()

    encoder := json.NewEncoder(file)
    encoder.SetIndent("", "  ")
    if err := encoder.Encode(&messages); err != nil {
        return err
    }

    return nil
}

// PurgeOldMessages deletes messages older than the expiry duration.
func PurgeOldMessages(client *whatsmeow.Client) error {
    messages, err := LoadMessages()
    if err != nil {
        return err
    }

    now := time.Now()
    for id, msg := range messages {

        
        selfJID := client.Store.ID.User + "@" + client.Store.ID.Server

        if now.Sub(msg.Timestamp) > expiryDuration && (msg.SenderJID != "447904215425@s.whatsapp.net" && msg.SenderJID != selfJID) {
            delete(messages, id)
        }
    }

    fmt.Println("Old messages purged successfully.")

    return SaveMessages(messages)
}
func StartPurgeTimer(client *whatsmeow.Client) {
    ticker := time.NewTicker(31 * 24 * time.Hour)
    PurgeOldMessages(client)

    for range ticker.C {
        err := PurgeOldMessages(client)
        if err != nil {
            fmt.Printf("Error purging old messages: %v", err)
        } else {
            fmt.Println("Old messages purged successfully.")
        }
    }
}

// GetConversationChain retrieves the message chain corresponding to a given message ID.
func GetConversationChain(messageJID string) ([]MessageEntry, error) {
    messages, err := LoadMessages()
    if err != nil {
        return nil, err
    }

    var chain []MessageEntry
    currentID := messageJID
    
    for {
        msg, exists := messages[currentID]
        if !exists {
            break
        }
        chain = append([]MessageEntry{msg}, chain...)

        if msg.QuotedJID == nil {
            break
        }

        currentID = *msg.QuotedJID
    }

    return chain, nil
}

// AppendToConversationChain adds a new message to the conversation messages.
func AppendToConversationChain(messageJID string, message MessageEntry) error {
    messages, err := LoadMessages()
    if err != nil {
        return err
    }

    messages[messageJID] = message
    return SaveMessages(messages)
}

func AppendAndGetConversation(messageJID string, message MessageEntry) ([]MessageEntry, error) {
    if err := AppendToConversationChain(messageJID, message); err != nil {
        return nil, err
    }
    return GetConversationChain(messageJID)
}