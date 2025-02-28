package commands

import (
    "go.mau.fi/whatsmeow"
    "go.mau.fi/whatsmeow/types/events"
)

// Command interface to be implemented by all commands
type Command interface {
    Execute(client *whatsmeow.Client, message *events.Message, args []string)
    Name() string
}