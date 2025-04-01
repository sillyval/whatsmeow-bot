// utils/conversation.go

package utils

import (
	"encoding/json"
	"fmt"
	"os"
	"sync"
)

// MessageEntry represents a single message in the conversation.
type MessageEntry struct {
    MessageContent string `json:"message_content"`
    MessageJID     string `json:"message_jid"`
    SenderJID      string `json:"sender_jid"`
    SenderPushName string `json:"sender_push_name"`
}

// ConversationChain represents the conversation chain for a specific chat.
type ConversationChain struct {
    ChatJID       string         `json:"chat_jid"`
    MostRecentJID string         `json:"most_recent_jid"`
    Chain         []MessageEntry `json:"chain"`
}

var (
    chainFilePath = "conversation_chains.json"
    mutex         = &sync.Mutex{}
)

// LoadConversationChains reads the conversation chains from the JSON file.
func LoadConversationChains() ([]ConversationChain, error) {
    mutex.Lock()
    defer mutex.Unlock()

    file, err := os.Open(chainFilePath)
    if err != nil {
        if os.IsNotExist(err) {
            return []ConversationChain{}, nil
        }
        return nil, err
    }
    defer file.Close()

    var chains []ConversationChain
    decoder := json.NewDecoder(file)
    if err := decoder.Decode(&chains); err != nil {
        return nil, err
    }

    return chains, nil
}

// SaveConversationChains writes the conversation chains to the JSON file.
func SaveConversationChains(chains []ConversationChain) error {
    mutex.Lock()
    defer mutex.Unlock()

    file, err := os.Create(chainFilePath)
    if err != nil {
        return err
    }
    defer file.Close()

    encoder := json.NewEncoder(file)
    encoder.SetIndent("", "  ")
    if err := encoder.Encode(&chains); err != nil {
        return err
    }

    return nil
}

// DuplicateConversationChain duplicates the conversation chain for a given chat and conversation.
func DuplicateConversationChain(chatJID, conversationJID string) error {
    chains, err := LoadConversationChains()
    if err != nil {
        return err
    }

    // Locate the chain to duplicate
    for _, chain := range chains {
        if chain.ChatJID == chatJID && chain.MostRecentJID == conversationJID {
            // Create a new chain that is an exact copy of the old one
            duplicatedChain := ConversationChain{
                ChatJID:       chain.ChatJID,
                MostRecentJID: chain.MostRecentJID,
                Chain:         make([]MessageEntry, len(chain.Chain)),
            }
            copy(duplicatedChain.Chain, chain.Chain)

            // Append the duplicated chain to the slice of chains
            chains = append(chains, duplicatedChain)
            
            // Save the updated slice of chains
            return SaveConversationChains(chains)
        }
    }
    
    return fmt.Errorf("failed to duplicate conversation chain: chain not found")
}

// GetConversationChain retrieves the conversation chain for a specific chat.
func GetConversationChain(chatJID, conversationJID string) ([]MessageEntry, bool, error) {
    chains, err := LoadConversationChains()
    if err != nil {
        return nil, false, err
    }

    for _, chain := range chains {
        if chain.ChatJID == chatJID && chain.MostRecentJID == conversationJID {
            if len(chain.Chain) > 0 {
                return chain.Chain, true, nil
            }
            return nil, false, nil
        }
    }

    return nil, false, nil
}

// AppendToConversationChain adds a new message to the conversation chain.
func AppendToConversationChain(chatJID, conversationJID string, message MessageEntry) error {
    chains, err := LoadConversationChains()
    if err != nil {
        return err
    }

    // Locate the chain and append the message
    for i, chain := range chains {
        if chain.MostRecentJID == conversationJID {
            chains[i].Chain = append(chains[i].Chain, message)
            chains[i].MostRecentJID = message.MessageJID //Update to the most recent JID
            return SaveConversationChains(chains)
        }
    }

    // If conversationID not found, create a new chain
    newChain := ConversationChain{
        ChatJID:       chatJID,
        MostRecentJID: message.MessageJID, //Update to the most recent JID
        Chain:         []MessageEntry{message},
    }
    chains = append(chains, newChain)
    return SaveConversationChains(chains)
}

func AppendAndGetConversation(chatJID, conversationJID string, message MessageEntry) ([]MessageEntry, error) {
    err := AppendToConversationChain(chatJID, conversationJID, message)
    if err != nil {
        return []MessageEntry{}, err
    }
    chain, exists, err := GetConversationChain(chatJID, message.MessageJID)
    if err != nil {
        return []MessageEntry{}, err
    }
    if !exists {
        return []MessageEntry{}, fmt.Errorf("chain doesn't exist")
    }
	return chain, nil
}