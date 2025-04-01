package cats

import (
	"time"

	"fmt"
	"io/ioutil"
	"strconv"
	"strings"

	"go.mau.fi/whatsmeow"
	"go.mau.fi/whatsmeow/types"
	"go.mau.fi/whatsmeow/types/events"

	_ "github.com/mattn/go-sqlite3"

	"whatsmeow-bot/utils"
)



const (
	newsletterJIDString     = "120363193700067731@newsletter"
	testNewsletterJIDString = "120363403157417087@newsletter"
	imageDir                = "/home/oliver/whatsmeow-bot/cats/"
	currentFile             = "/home/oliver/whatsmeow-bot/cats/current.txt"
	totalFile               = "/home/oliver/whatsmeow-bot/cats/total.txt"
	captionFile             = "/home/oliver/whatsmeow-bot/cats/caption.txt"

	testMode = false
)

func Start(client *whatsmeow.Client) {
	fmt.Println("Started cat image uploader")

	// preview
	currentIndex, err := readInt(currentFile)
	if err != nil {
		fmt.Printf("Error reading current index: %v", err)
		return
	}
	fmt.Printf("Current index: %v\n", currentIndex)

	totalImages, err := readInt(totalFile)
	if err != nil {
		fmt.Printf("Error reading total images: %v", err)
		return
	}
	fmt.Printf("Total indexes: %v\n", totalImages)

	PreviewNextUpload(client, currentIndex, totalImages, nil)

	fmt.Println("Setting up timer...")
	scheduleUploads(client)
}

func CurrentIndex() int {
	currentIndex, err := readInt(currentFile)
	if err != nil {
		fmt.Printf("Error reading current index: %v", err)
		return -1
	}

	return currentIndex
}
func TotalImages() int {
	totalImages, err := readInt(totalFile)
	if err != nil {
		fmt.Printf("Error reading total images: %v", err)
		return -1
	}

	return totalImages
}
func NextIndex(currentIndex int) int {
	totalImages := TotalImages()

	nextIndex := currentIndex + 1
	if nextIndex > totalImages {
		nextIndex = 1
	}

	return nextIndex
}
func Caption() string {
	caption, err := readStr(captionFile)
	if err != nil {
		fmt.Printf("Error reading caption: %v", err)
		caption = ""
	}
	return caption
}



func Skip(client *whatsmeow.Client) {
	fmt.Println("Skipping this image")

	newsletterJID, _ := types.ParseJID(testNewsletterJIDString)

	currentIndex := CurrentIndex()
	totalImages := TotalImages()
	nextIndex := NextIndex(currentIndex)
	fmt.Printf("Current index: %v/%v\n", currentIndex, totalImages)

	utils.NewsletterMessage(client, newsletterJID, fmt.Sprintf("Skipping this image! Next: %v/%v", NextIndex(nextIndex), totalImages))

	writeInt(currentFile, nextIndex)

	PreviewNextUpload(client, nextIndex, totalImages, nil)
}

func OnMessage(client *whatsmeow.Client, messageEvent *events.Message) {

	newsletterJID, _ := types.ParseJID(testNewsletterJIDString)

	if messageEvent.Info.Chat != newsletterJID {
		return
	}

	messageBody := utils.GetMessageBody(messageEvent.Message)

	if strings.HasPrefix(messageBody, " ") {
		// Bot message
		return
	}

	if !strings.HasPrefix(messageBody, "caption: ") {
		// Bot message
		return
	}

	messageBody = strings.Replace(messageBody, "caption: ", "", 1)

	writeStr(captionFile, messageBody)

	utils.NewsletterMessage(client, newsletterJID, fmt.Sprintf(" updated caption of next image to `%s`!", messageBody))
	utils.NewsletterReact(client, messageEvent, "✅")
}

func scheduleUploads(client *whatsmeow.Client) {

	londonLoc, err := time.LoadLocation("Europe/London")
    if err != nil {
        fmt.Printf("Error loading Europe/London location: %v\n", err)
        return
    }

	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()

	prevMinute := -1

	for {
		now := time.Now().In(londonLoc)

		nowHour, nowMinute := now.Hour(), now.Minute()

		//fmt.Printf("%v:%v\n", nowHour, nowMinute)
		
		if (prevMinute != nowMinute) && ((nowHour == 7 && nowMinute == 0) || (nowHour == 22 && nowMinute == 0)) {
			fmt.Printf("%v:%v, uploading!!\n", nowHour, nowMinute)
			uploadCatImage(client)
		}
		prevMinute = nowMinute
		
		<-ticker.C
	}
}

func PreviewNextUpload(client *whatsmeow.Client, currentIndex, totalImages int, caption *string) {
	nextIndex := currentIndex + 1
	if nextIndex > totalImages {
		nextIndex = 1 // Loop back to first image
	}
	fmt.Printf("Previewing index: %v\n", nextIndex)

	imagePath := fmt.Sprintf("%scat_%d.jpg", imageDir, nextIndex)

	imageData, err := ioutil.ReadFile(imagePath)
	if err != nil {
		fmt.Printf("Error reading image %s: %v\n", imagePath, err)
		return
	}

	newsletterJID, _ := types.ParseJID(testNewsletterJIDString)
	fmt.Printf("Sending to preview newsletter: %v\n", newsletterJID)

	if caption == nil {
		captionText := " next cat image that will be uploaded! send a message starting with `caption: ` to set the caption, new captions will update it"
		caption = &captionText
	}

	utils.NewsletterSendImage(client, newsletterJID, *caption, imageData)

	fmt.Printf("Successfully sent cat_%d.jpg to newsletter\n", nextIndex)
}

func uploadCatImage(client *whatsmeow.Client) {
	currentIndex := CurrentIndex()
	totalImages := TotalImages()
	nextIndex := NextIndex(currentIndex)
	fmt.Printf("Current index: %v/%v\n", currentIndex, totalImages)

	caption := Caption()
	writeStr(captionFile, "")

	fmt.Printf("Caption: %v\n", caption)

	imagePath := fmt.Sprintf("%scat_%d.jpg", imageDir, nextIndex)

	imageData, err := ioutil.ReadFile(imagePath)
	if err != nil {
		fmt.Printf("Error reading image %s: %v\n", imagePath, err)
		return
	}

	var jidString string
	if testMode {
		jidString = testNewsletterJIDString
	} else {
		jidString = newsletterJIDString
	}

	newsletterJID, _ := types.ParseJID(jidString)
	fmt.Printf("Sending to newsletter: %v\n", newsletterJID)

	utils.NewsletterSendImage(client, newsletterJID, caption, imageData)

	fmt.Printf("Successfully sent cat_%d.jpg to newsletter\n", nextIndex)
	writeInt(currentFile, nextIndex)

	PreviewNextUpload(client, nextIndex, totalImages, nil)
}

func readInt(filename string) (int, error) {
	data, err := ioutil.ReadFile(filename)
	if err != nil {
		return 0, err
	}

	index, err := strconv.Atoi(strings.TrimSpace(string(data)))
	if err != nil {
		return 0, err
	}
	return index, nil
}
func readStr(filename string) (string, error) {
	data, err := ioutil.ReadFile(filename)
	if err != nil {
		return "", err
	}

	str := strings.TrimSpace(string(data))

	return str, nil
}

func writeInt(filename string, index int) {
	err := ioutil.WriteFile(filename, []byte(strconv.Itoa(index)), 0644)
	if err != nil {
		fmt.Printf("Failed to write %s: %v", filename, err)
	}
}
func writeStr(filename string, str string) {
	err := ioutil.WriteFile(filename, []byte(str), 0644)
	if err != nil {
		fmt.Printf("Failed to write %s: %v", filename, err)
	}
}
