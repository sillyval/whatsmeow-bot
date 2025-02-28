package commands

var commandRegistry = make(map[string]Command)

func RegisterCommand(command Command) {
    commandRegistry[command.Name()] = command
}

func GetCommand(name string) Command {
    return commandRegistry[name]
}