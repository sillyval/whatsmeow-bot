package main

import (
    "context"
    "encoding/json"
    "fmt"
    "io/ioutil"
    "os"
    "os/signal"
    "strings"
    "strconv"
    "syscall"

    "whatsmeow-bot/commands" // Replace with your actual module path

    "go.mau.fi/whatsmeow"
    "go.mau.fi/whatsmeow/store/sqlstore"
    "go.mau.fi/whatsmeow/types/events"
    waLog "go.mau.fi/whatsmeow/util/log"

    _ "github.com/mattn/go-sqlite3"
    "whatsmeow-bot/utils"
    "whatsmeow-bot/cats"
    "github.com/skip2/go-qrcode"
)

type Config struct {
    Prefixes   []string `json:"prefixes"`
}
type SecretConfig struct {
    DiscordToken string `json:"discord-token"`
    OpenAIKey string `json:"openai-key"`
}


func parseArguments(input string) []string {
    var args []string
    current := strings.Builder{}
    inCodeBlock := false
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
            if inCodeBlock {
                escaping = true
            } else {
                current.WriteByte(ch)
            }
        case '`':
            if i+2 < len(input) && input[i+1] == '`' && input[i+2] == '`' {
                if inCodeBlock {
                    args = append(args, current.String())
                    current.Reset()
                }
                inCodeBlock = !inCodeBlock
                i += 2 // Skip the next two backticks
            } else {
                current.WriteByte(ch)
            }
        case ' ':
            if inCodeBlock {
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

func stringToARGB(colorStr string) uint32 {
	color, err := strconv.ParseUint(colorStr, 16, 32)
	if err != nil {
		fmt.Println(err)
        color = 0xFF000000
	}
	return uint32(color)
}

func eventHandler(client *whatsmeow.Client, config *Config) func(evt interface{}) {
    return func(evt interface{}) {
        if v, ok := evt.(*events.Message); ok {

            if v.Message.AudioMessage != nil {
                go utils.SendAudioMessageRead(client, v.Info.Chat, v.Info.Sender, v.Info.ID)
            } else {
                go utils.SendMessageRead(client, v.Info.Chat, v.Info.Sender, v.Info.ID)
            }

            go cats.OnMessage(client, v)

            //fmt.Printf("New message: \n%+v\n", v)

            messageBody := utils.GetMessageBody(v.Message)

            var messagePrefix string
            var commandName string
            for _, prefix := range config.Prefixes {
                if strings.HasPrefix(messageBody, prefix) {
                    messagePrefix = prefix
                    break
                }
            }

            args := parseArguments(messageBody[len(messagePrefix):])
            var firstArg string
            if len(args) == 0 {
                firstArg = ""
            } else {
                firstArg = args[0]
            }

            commandName = strings.ToLower(firstArg)
            cmd := commands.GetCommand(commandName)

            if messagePrefix == "" {
                cmd = nil
            }

            quotedMessage := utils.GetQuotedMessage(v)

            isQuotingGPT := utils.IsGPTMessage(quotedMessage)
            selfJID := client.Store.ID.User + "@" + client.Store.ID.Server
            mentionsMe := utils.IsJIDMentioned(v, selfJID)
            isDMS := utils.IsDMS(v) && (v.Info.Sender.User + "@" + v.Info.Sender.Server) != selfJID

            //fmt.Printf("Is quoting GPT: %v\nMentions me: %v\nIs DMS: %v\n\n",isQuotingGPT, mentionsMe, isDMS)

            if cmd == nil && (isQuotingGPT || mentionsMe || isDMS) {
                commandName = "gpt"
                cmd = commands.GetCommand(commandName)
            } else {
                if len(args) > 0 {
                    args = args[1:]
                } else {
                    args = []string{}
                }
            }

            if cmd != nil && !utils.IsStatusPost(v) && (messagePrefix != "" || commandName != "") {
                go func() {
                    cmd.Execute(client, v, args)  
                }()
            } else {
                //fmt.Printf("Command '%s' not recognized.\n", commandName)
            }
        }
    }
}

func main() {

    fmt.Println("Starting client...")

    config, err := LoadConfig("config.json")
    if err != nil {
        panic(err)
    }

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

    client.AddEventHandler(eventHandler(client, config))
    go cats.Start(client)
    go utils.StartPurgeTimer(client)

    if client.Store.ID == nil {
        qrChan, _ := client.GetQRChannel(context.Background())
        err = client.Connect()
        if err != nil {
            panic(err)
        }
        for evt := range qrChan {
            if evt.Event == "code" {
                qr, err := qrcode.New(evt.Code, qrcode.High)
                if err != nil {
                    fmt.Printf("Failed to generate QR code: %v", err)
                }

                fmt.Println(qr.ToSmallString(false))
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