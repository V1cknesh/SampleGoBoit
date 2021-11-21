package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strconv"
	"strings"
)

const telegramBotToken = "2109146958:AAEGX548AcSAgv-vOOYWigroJoxreiNzrjM"

type UpdateFromTelegramApi struct {
	UpdateId        int             `json:"update_id"`
	TelegramMessage TelegramMessage `json:"message"`
}

type TelegramMessage struct {
	TextMessage string       `json:"text"`
	Chat        TelegramChat `json:"chat"`
}

type TelegramChat struct {
	Id int `json:"id"`
}

func main() {
	var updateFromTelegramApi UpdateFromTelegramApi
	var updateFromTelegramApiJson = `{
		"update_id": 1,
		"message": {
			"text": "hello world",
			"chat": {
				"id": 123456789
			}
		}
	}`

	var jsonBytes = []byte(updateFromTelegramApiJson)
	var err = json.Unmarshal(jsonBytes, &updateFromTelegramApi)

	if err != nil {
		log.Printf("Error: %s", err)
	}

	var request = httptest.NewRequest("POST", "/", bytes.NewBuffer(jsonBytes))

	var parseUpdateMessage, errParse = parseTelegramRequest(request)

	if errParse != nil {
		log.Printf("Error: %s", errParse.Error())
	}

	log.Printf("Telegram message sent : %d", parseUpdateMessage)

	sendTextToTelegramChat(-654213075, "Github Copilot makes coding fun")
}

func dummyMethod(seed string) (string, error) {

	fmt.Println("Dummy method")

	return string("hello world"), fmt.Errorf("Dummy Error")

}

func parseTelegramRequest(r *http.Request) (*UpdateFromTelegramApi, error) {
	var updateFromTelegramApi UpdateFromTelegramApi
	if err := json.NewDecoder(r.Body).Decode(&updateFromTelegramApi); err != nil {
		log.Printf("could not decode incoming update %s", err.Error())
		return nil, err
	}
	return &updateFromTelegramApi, nil
}

func HandleTelegramHook(w http.ResponseWriter, r *http.Request) {
	var updateFromTelegramApi, err = parseTelegramRequest(r)
	if err != nil {
		log.Printf("error parsing update, %s", err.Error())
		return
	}

	var sanitizeSeed = strings.Replace(updateFromTelegramApi.TelegramMessage.TextMessage, " ", "", -1)
	sanitizeSeed = strings.Replace(sanitizeSeed, "\n", "", -1)
	sanitizeSeed = strings.Replace(sanitizeSeed, "\r", "", -1)
	sanitizeSeed = strings.Replace(sanitizeSeed, "\t", "", -1)

	var seed = strings.ToLower(sanitizeSeed)
	seed = strings.TrimSpace(seed)

	if len(seed) < 1 {
		log.Printf("no seed provided")
		return
	}

	var lyric, errRapLyric = dummyMethod(sanitizeSeed)

	if errRapLyric != nil {
		log.Printf("error getting punch line, %s", errRapLyric.Error())
		return
	}

	var chatId = updateFromTelegramApi.TelegramMessage.Chat.Id
	var _, errTelegram = sendTextToTelegramChat(chatId, lyric)

	if errTelegram != nil {
		log.Printf("error sending text to telegram chat, %s", errTelegram.Error())
		return
	} else {
		log.Printf("punchline %s sent text to telegram chat id %d", lyric, chatId)
	}
}

func sendTextToTelegramChat(chatId int, text string) (string, error) {
	var telegramApi string = fmt.Sprintf("https://api.telegram.org/bot%s/sendMessage?chat_id=%d&text=%s", telegramBotToken, chatId, text)
	response, err := http.PostForm(
		telegramApi,
		url.Values{
			"chat_id": {strconv.Itoa(chatId)},
			"text":    {text},
		})

	if err != nil {
		log.Printf("error sending text to telegram chat, %s", err.Error())
		return "", err
	}

	defer response.Body.Close()

	var bodyBytes, errRead = ioutil.ReadAll(response.Body)
	if errRead != nil {
		log.Printf("error reading response body, %s", errRead.Error())
		return "", errRead
	}

	bodyString := string(bodyBytes)
	log.Printf("response body: %s", bodyString)
	return bodyString, nil
}
