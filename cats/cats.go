package cats

import (
	"time"

	"fmt"
	"io/ioutil"
	"strconv"
	"strings"

	"go.mau.fi/whatsmeow"
	"go.mau.fi/whatsmeow/types"

	_ "github.com/mattn/go-sqlite3"

	"whatsmeow-bot/utils"
)



const (
	newsletterJIDString = "120363193700067731@newsletter"
	imageDir      = "/home/oliver/whatsmeow-bot/cat_images/"
	currentFile   = "/home/oliver/whatsmeow-bot/cat_images/current.txt"
	totalFile     = "/home/oliver/whatsmeow-bot/cat_images/total.txt"
)

func Start(client *whatsmeow.Client) {
	fmt.Println("Started cat image uploader")
	scheduleUploads(client)
}

func scheduleUploads(client *whatsmeow.Client) {
	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()

	prevMinute := -1

	for {
		now := time.Now()

		nowHour, nowMinute := now.Hour(), now.Minute()

		//fmt.Printf("%v:%v\n", nowHour, nowMinute)
		
		if (prevMinute != nowMinute) && ((nowHour == 7 && nowMinute == 0) || (nowHour == 22 && nowMinute == 0)) {
			fmt.Printf("%v:%v, uploading!!\n", nowHour, nowMinute)
			uploadCatImage(client)
			prevMinute = nowMinute
		}
		<-ticker.C
	}
}

func uploadCatImage(client *whatsmeow.Client) {
	currentIndex, err := readIndex(currentFile)
	if err != nil {
		fmt.Printf("Error reading current index: %v", err)
		return
	}

	totalImages, err := readIndex(totalFile)
	if err != nil {
		fmt.Printf("Error reading total images: %v", err)
		return
	}

	nextIndex := currentIndex + 1
	if nextIndex > totalImages {
		nextIndex = 1 // Loop back to first image
	}

	imagePath := fmt.Sprintf("%scat_%d.jpg", imageDir, nextIndex)

	imageData, err := ioutil.ReadFile(imagePath)
	if err != nil {
		fmt.Printf("Error reading image %s: %v\n", imagePath, err)
		return
	}

	newsletterJID, _ := types.ParseJID(newsletterJIDString)

	utils.NewsletterSendImage(client, newsletterJID, "", imageData)

	fmt.Printf("Successfully sent cat_%d.jpg to newsletter\n", nextIndex)
	writeIndex(currentFile, nextIndex)
}

func readIndex(filename string) (int, error) {
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

func writeIndex(filename string, index int) {
	err := ioutil.WriteFile(filename, []byte(strconv.Itoa(index)), 0644)
	if err != nil {
		fmt.Printf("Failed to write %s: %v", filename, err)
	}
}
