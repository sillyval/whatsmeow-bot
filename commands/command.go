package commands

import (
    "go.mau.fi/whatsmeow"
    "go.mau.fi/whatsmeow/types/events"
    "io/ioutil"
    "encoding/json"
)

type SecretConfig struct {
    OpenAIKey string `json:"openai-key"`
}

func LoadSecretConfig(filename string) (*SecretConfig, error) {
    configBytes, err := ioutil.ReadFile(filename)
    if err != nil {
        return nil, err
    }
    var config SecretConfig
    if err := json.Unmarshal(configBytes, &config); err != nil {
        return nil, err
    }
    return &config, nil
}

// Command interface to be implemented by all commands
type Command interface {
    Execute(client *whatsmeow.Client, message *events.Message, args []string)
    Name() string
    Description() string
}