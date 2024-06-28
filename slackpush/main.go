package slackpush

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
)

func SendMessage(method, url string, payload interface{}) error {
	r := strings.NewReader(fmt.Sprint(payload))
	client := &http.Client{}

	req, err := http.NewRequest(method, url, r)
	if err != nil {
		log.Printf("HTTP Request Build up error %v", err)
		return err
	}

	req.Header.Add("Content-Type", "application/json")
	resp, err := client.Do(req)
	if err != nil {
		log.Printf("Send Message error %v", err)
		return err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Printf("Read Response Error %v", err)
		return err
	}

	log.Printf("Response Body: %v", string(body))
	return nil
}
