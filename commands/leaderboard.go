package commands

import (
    "bytes"
    "fmt"
    "io/ioutil"
    "os"
    "os/exec"

    "go.mau.fi/whatsmeow"
    "go.mau.fi/whatsmeow/types/events"

    "whatsmeow-bot/utils"
)

func init() {
    RegisterCommand(&LeaderboardCommand{})
}

type LeaderboardCommand struct{}

func (c *LeaderboardCommand) Execute(client *whatsmeow.Client, message *events.Message, args []string) *string {
    // Check if the message has a document attachment
    if message.Message.DocumentMessage == nil || message.Message.DocumentMessage.Mimetype == nil ||
        *message.Message.DocumentMessage.Mimetype != "text/plain" {

        reply := "Please attach a valid text file."
        utils.Reply(client, message, reply)
		utils.React(client, message, "❌")
        return nil
    }

    // Download the text file
    data, mimeType, err := utils.GetMediaFromMessage(client, message.Message)

    if err != nil || mimeType != "text/plain" {
        reply := "Error downloading the text file."
        utils.Reply(client, message, reply)
		utils.React(client, message, "❌")
        return nil
    }

    // Create a temporary file for the text
    tmpFile, err := ioutil.TempFile("", "whatsapp_text_*.txt")
    if err != nil {
        reply := "Unable to create temporary file."
        utils.Reply(client, message, reply)
		utils.React(client, message, "❌")
        return nil
    }
    defer os.Remove(tmpFile.Name()) // Clean up

    // Write data to the temporary file
    if _, err := tmpFile.Write(data); err != nil {
        reply := "Failed to write data to file."
        utils.Reply(client, message, reply)
		utils.React(client, message, "❌")
        return nil
    }
    tmpFile.Close()

    // Execute the script
    cmd := exec.Command("venv/bin/python", "chat_stats/stats.py", "-f", tmpFile.Name())
    var out bytes.Buffer
    var stderr bytes.Buffer
    cmd.Stdout = &out
    cmd.Stderr = &stderr

    if err := cmd.Run(); err != nil {
        reply := fmt.Sprintf("Failed to execute script: %v", stderr.String())
        utils.Reply(client, message, reply)
		utils.React(client, message, "❌")
        return nil
    }

    // Reply with the result from the command
    result := out.String()
    utils.Reply(client, message, result)
	utils.React(client, message, "✅")
    return nil
}

func (c *LeaderboardCommand) Name() string {
    return "leaderboard"
}

func (c *LeaderboardCommand) Description() string {
    return "Generates chat statistics from attached chat export file"
}