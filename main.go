package main

import (
    "context"
    "encoding/json"
    "fmt"
    "io/ioutil"
    "os"
    "os/signal"
    "strings"
    "syscall"

    "whatsmeow-bot/commands" // Replace with your actual module path

    "go.mau.fi/whatsmeow"
    "go.mau.fi/whatsmeow/store/sqlstore"
    "go.mau.fi/whatsmeow/types/events"
    waLog "go.mau.fi/whatsmeow/util/log"

    _ "github.com/mattn/go-sqlite3"
    "whatsmeow-bot/utils"
)

type Config struct {
    Prefixes []string `json:"prefixes"`
}


func parseArguments(input string) []string {
    var args []string
    current := strings.Builder{}
    inQuotes := false
    escaping := false

    for i := 0; i < len(input); i++ {
        ch := input[i]

        if escaping {
            current.WriteByte(ch)
            escaping = false
            continue
        }

        switch ch {
        case '\\':
            if inQuotes {
                escaping = true
            } else {
                current.WriteByte(ch)
            }
        case '"':
            if inQuotes {
                args = append(args, current.String())
                current.Reset()
            }
            inQuotes = !inQuotes
        case ' ':
            if inQuotes {
                current.WriteByte(ch)
            } else if current.Len() > 0 {
                args = append(args, current.String())
                current.Reset()
            }
        default:
            current.WriteByte(ch)
        }
    }

    if current.Len() > 0 {
        args = append(args, current.String())
    }

    return args
}


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


func eventHandler(client *whatsmeow.Client, config *Config) func(evt interface{}) {
    return func(evt interface{}) {
        if v, ok := evt.(*events.Message); ok {

            //fmt.Printf("New message: \n%+v\n", v.Message)

            messageBody := utils.GetMessageBody(v.Message)

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