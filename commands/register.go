package commands

import (
    "strings"
)

var commandRegistry = make(map[string]Command)

func GetAllCommands() []Command {
    commands := []Command{}
    for _, cmd := range commandRegistry {
        commands = append(commands, cmd)
    }
    return commands
}

func RegisterCommand(command Command) {
    commandName := command.Name()
    commandAliases := strings.Split(commandName, ",")
    for _, alias := range commandAliases {
        commandRegistry[alias] = command
    }
}

func GetCommand(name string) Command {
    return commandRegistry[name]
}