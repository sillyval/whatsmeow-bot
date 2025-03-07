package main

import (
    "context"
    "encoding/json"
    "fmt"
    "io/ioutil"
    "os"
    "os/signal"
    "strings"
    "regexp"
    "syscall"

    "whatsmeow-bot/commands" // Replace with your actual module path

    "go.mau.fi/whatsmeow"
    "go.mau.fi/whatsmeow/store/sqlstore"
    "go.mau.fi/whatsmeow/types/events"
    waLog "go.mau.fi/whatsmeow/util/log"

    _ "github.com/mattn/go-sqlite3"
)

type Config struct {
    Prefixes []string `json:"prefixes"`
}

var argRegex = regexp.MustCompile(`"((?:[^"\\]|\\.|[\n])*)"|(\S+)`)

func LoadConfig(filename string) (*Config, error) {
    configBytes, err := ioutil.ReadFile(filename)
    if err != nil {
        return nil, err
    }
    var config Config
    if err := json.Unmarshal(configBytes, &config); err != nil {
        return nil, err
    }
    return &config, nil
}

func parseArguments(input string) []string {
    matches := argRegex.FindAllStringSubmatch(input, -1)
    var args []string
    for _, match := range matches {
        if match[1] != "" { // Quoted argument
            unescaped := strings.ReplaceAll(match[1], `\"`, `"`)
            args = append(args, unescaped)
        } else { // Unquoted argument
            args = append(args, match[2])
        }
    }
    return args
}

func eventHandler(client *whatsmeow.Client, config *Config) func(evt interface{}) {
    return func(evt interface{}) {
        if v, ok := evt.(*events.Message); ok {

            //fmt.Printf("New message: \n%+v\n", v.Message)

            var messageBody string
            if v.Message.Conversation != nil {
                messageBody = *v.Message.Conversation
            } else if v.Message.ExtendedTextMessage != nil {
                messageBody = *v.Message.ExtendedTextMessage.Text
            } else if v.Message.ImageMessage.Caption != nil {
                messageBody = *v.Message.ImageMessage.Caption
            } else if v.Message.VideoMessage.Caption != nil {
                messageBody = *v.Message.VideoMessage.Caption
            }

            var messagePrefix string
            for _, prefix := range config.Prefixes {
                if strings.HasPrefix(messageBody, prefix) {
                    messagePrefix = prefix
                    break
                }
            }

            if messagePrefix == "" {
                return
            }

            args := parseArguments(messageBody[len(messagePrefix):])
            if len(args) == 0 {
                return
            }
            commandName := strings.ToLower(args[0])
            args = args[1:]

            cmd := commands.GetCommand(commandName)
            if cmd != nil {
                cmd.Execute(client, v, args)
            } else {
                fmt.Printf("Command '%s' not recognized.\n", commandName)
            }
        }
    }
}

func main() {

    fmt.Println("Starting client...")

    dbLog := waLog.Stdout("Database", "ERROR", true) // DEBUG for full log
    clientLog := waLog.Stdout("Client", "ERROR", true) // ditto
    container, err := sqlstore.New("sqlite3", "file:credentials.db?_foreign_keys=on", dbLog)
    if err != nil {
        panic(err)
    }
    deviceStore, err := container.GetFirstDevice()
    if err != nil {
        panic(err)
    }
    client := whatsmeow.NewClient(deviceStore, clientLog)

    config, err := LoadConfig("config.json")
    if err != nil {
        panic(err)
    }

    client.AddEventHandler(eventHandler(client, config))

    if client.Store.ID == nil {
        qrChan, _ := client.GetQRChannel(context.Background())
        err = client.Connect()
        if err != nil {
            panic(err)
        }
        for evt := range qrChan {
            if evt.Event == "code" {
                fmt.Println("Scan this QR code:", evt.Code)
            }
        }
    } else {
        err = client.Connect()
        if err != nil {
            panic(err)
        }
        fmt.Println("Client running!")
    }

    c := make(chan os.Signal, 1)
    signal.Notify(c, os.Interrupt, syscall.SIGTERM)
    <-c

    client.Disconnect()
}