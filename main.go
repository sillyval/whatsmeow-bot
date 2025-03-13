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
    "whatsmeow-bot/discordbot"
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

            //fmt.Printf("New message: \n%+v\n", v.Message)

            messageBody := utils.GetMessageBody(v.Message)

            var messagePrefix string
            for _, prefix := range config.Prefixes {
                if strings.HasPrefix(messageBody, prefix) {
                    messagePrefix = prefix
                    break
                }
            }

            args := parseArguments(messageBody[len(messagePrefix):])
            if len(args) != 0 && messagePrefix != "" {
                commandName := strings.ToLower(args[0])
                args = args[1:]

                cmd := commands.GetCommand(commandName)
                if cmd != nil {
                    response := cmd.Execute(client, v, args)

                    if response != nil && v.Info.IsFromMe && strings.Contains(commandName, "status") {
                        jid := v.Info.Sender
                        jidString := jid.String()
                        contactName := utils.GetContactName(client, jid)
                        pushUsername := utils.GetPushName(client, jid)

                        var colourRGB int
                        var statusText string
                        if commandName == "colourstatus" {
                            colourARGB := stringToARGB(args[0])
                            colourRGB = int(colourARGB & 0x00FFFFFF)
                            statusText = strings.Join(args[2:], " ")
                        } else {
                            colourRGB = int(0x000000)
                            statusText = strings.Join(args, " ")
                        }

                        logTextStatus(jidString, contactName, pushUsername, statusText, *response, &colourRGB)
                    }
                } else {
                    fmt.Printf("Command '%s' not recognized.\n", commandName)
                }
            }

            // logger
           
            if utils.IsStatusPost(v) && utils.IsIncomingMessage(v) {

                if messageBody == "" {
                    messageBody = "-# *no body provided*"
                }

                jid := v.Info.Sender
                jidString := jid.String()
                messageJID := v.Info.ID
                contactName := utils.GetContactName(client, jid)
                pushUsername := utils.GetPushName(client, jid)
                colourARGB := v.Message.ExtendedTextMessage.GetBackgroundArgb()
                colourRGB := int(colourARGB & 0x00FFFFFF)
 
                fmt.Printf("\nStatus `%s` logged\n", messageJID)
    
                if v.Message.ImageMessage == nil && v.Message.VideoMessage == nil && v.Message.AudioMessage == nil {
                    logTextStatus(jidString, contactName, pushUsername, messageBody, messageJID, &colourRGB)
                } else {
                    imageData, mimetype, err := utils.GetMediaFromMessage(client, v)
                    if err == nil {
                        logMediaStatus(jidString, contactName, pushUsername, messageBody, messageJID, imageData, mimetype)
                    }
                }
            } else if utils.IsStatusPost(v) && utils.IsDeletedMessage(v) {

                jid := v.Info.Sender
                jidString := jid.String()
                messageJID := utils.GetDeletedMessageID(v)
                contactName := utils.GetContactName(client, jid)
                pushUsername := utils.GetPushName(client, jid)

                discordbot.React(jidString, contactName, pushUsername, *messageJID, "❌")
            }
        }
    }
}

func initialiseDiscordBot(secret_config *SecretConfig) {
    err := discordbot.InitBot(secret_config.DiscordToken)
	if err != nil {
		fmt.Printf("Failed to start discord bot: %v\n", err)
	}
	//defer discordbot.CloseBot()
}

func logTextStatus(jid, contactName, whatsappName, statusText, statusJID string, colour *int) {
    fmt.Printf("Logging `%s` sent by `%s`", statusText, contactName)
    
    err := discordbot.LogStatus(jid, contactName, whatsappName, statusText, statusJID, colour)
	if err != nil {
		fmt.Printf("Error logging status: %s\n", err)
	}
}
func logMediaStatus(jid, contactName, whatsappName, captionText, statusJID string, imageData []byte, mimetype string) {
    fmt.Printf("Logging `%s` (and an image) sent by `%s`", captionText, contactName)
    err := discordbot.LogStatusWithMedia(jid, contactName, whatsappName, captionText, statusJID, imageData, mimetype)
	if err != nil {
		fmt.Printf("Error logging image status: %s\n", err)
	}
}

func main() {

    fmt.Println("Starting client...")

    config, err := LoadConfig("config.json")
    if err != nil {
        panic(err)
    }
    secret_config, err := LoadSecretConfig("secret-config.json")
    if err != nil {
        panic(err)
    }

    initialiseDiscordBot(secret_config)

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