package commands

import (
	"fmt"
	"go.mau.fi/whatsmeow"
	"go.mau.fi/whatsmeow/types/events"
	"sort"
	"strings"
	"whatsmeow-bot/utils"
)

func init() {
	RegisterCommand(&HelpCommand{})
}

type HelpCommand struct{}

func (c *HelpCommand) Execute(client *whatsmeow.Client, message *events.Message, args []string) *string {
	var response strings.Builder
	commands := GetAllCommands()
	if len(commands) == 0 {
		utils.Reply(client, message, "No available commands.")
		return nil
	}

	response.WriteString("```\n┌─── Available Commands ───\n")

	seenCommands := make(map[string]bool) // To track already added commands

	var commandList []string

	// Collect commands and their descriptions
	for _, command := range commands {
		if description := command.Description(); description != "" {
			names := command.Name()
			if seenCommands[names] {
				continue // Skip duplicate commands
			}
			seenCommands[names] = true

			nameSlice := strings.Split(names, ",")
			fullName := strings.Join(nameSlice, "/") // Format as "cmd1/cmd2/..."
			commandList = append(commandList, fmt.Sprintf("%s - %s", fullName, description))
		}
	}

	// Sort the commands alphabetically
	sort.Strings(commandList)

	// Add sorted commands to the response
	for _, command := range commandList {
		response.WriteString(fmt.Sprintf("│ %s\n", command))
	}

	response.WriteString("└────────────\n```")

	utils.Reply(client, message, response.String())

	return nil
}

func (c *HelpCommand) Name() string {
	return "help"
}

func (c *HelpCommand) Description() string {
	return "Displays this help message"
}
