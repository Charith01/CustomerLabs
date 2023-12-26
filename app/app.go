package app

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"sync"

	constants "github.com/Charith01/CustomerLabs/constant"
	helper "github.com/Charith01/CustomerLabs/helpers"
	"github.com/Charith01/CustomerLabs/models"
)

func FormWebHookRequest(input map[string]string) models.WebhookRequest {
	webhookPayload := models.WebhookRequest{
		Event:           input["ev"],
		EventType:       input["et"],
		AppID:           input["id"],
		UserID:          input["uid"],
		MessageID:       input["mid"],
		PageTitle:       input["t"],
		PageURL:         input["p"],
		BrowserLanguage: input["l"],
		ScreenSize:      input["sc"],
		Attributes:      make(map[string]models.Attribute, 0),
		UserTraits:      make(map[string]models.Attribute, 0),
	}

	for key, keyValue := range input {
		if len(key) >= 4 && key[:4] == "atrk" {
			index := key[4:]
			value := input["atrv"+index]
			inputType := input["atrt"+index]
			webhookPayload.Attributes[keyValue] = models.Attribute{
				Type:  inputType,
				Value: value,
			}
		} else if len(key) >= 5 && key[:5] == "uatrk" {
			index := key[5:]
			value := input["uatrv"+index]
			inputType := input["uatrt"+index]
			webhookPayload.UserTraits[keyValue] = models.Attribute{
				Type:  inputType,
				Value: value,
			}
		}
	}
	return webhookPayload
}

func SendReqToWebhook(requestBody models.WebhookRequest) error {
	url := "https://webhook.site/" + constants.WebhookURL

	jsonData, err := json.Marshal(requestBody)
	if err != nil {
		return err
	}
	resp, err := http.Post(url, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return err
	}
	// defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return errors.New("webhook Request Failed")
	}

	log.Printf("WebHook request request sent successfully")
	return nil
}

func SendEventToWorker(writer http.ResponseWriter, request *http.Request) {

	var payload map[string]string
	var wg sync.WaitGroup
	errChan := make(chan error, 1)

	err := json.NewDecoder(request.Body).Decode(&payload)
	if err != nil {
		log.Printf("Error decoding request body: %v", err)
		helper.JsonWriterString(writer, http.StatusBadRequest, "decode error : "+err.Error())
		return
	}
	wg.Add(1)

	go func() {
		defer wg.Done()
		webHookPayload := FormWebHookRequest(payload)
		err = SendReqToWebhook(webHookPayload)
		if err != nil {
			errChan <- err
		}
		close(errChan)
	}()
	wg.Wait()
	if err := <-errChan; err != nil {
		fmt.Println("a", err)
		log.Printf("Internal Server Error: %v", err)
		helper.JsonWriterString(writer, http.StatusInternalServerError, "Internal Server Error : "+err.Error())
		return
	}

	helper.JsonWriterString(writer, http.StatusOK, "Event Sent Successfully")
}
