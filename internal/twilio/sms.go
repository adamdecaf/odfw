package twilio

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
)

type SMSConfig struct {
	AccountSid string
	AuthToken  string
	From       string
	To         []string
}

func SendSMS(client *http.Client, cfg *SMSConfig, to, message string) error {
	location := "https://api.twilio.com/2010-04-01/Accounts/" + cfg.AccountSid + "/Messages.json"

	msgData := url.Values{}
	msgData.Set("To", to)
	msgData.Set("From", cfg.From)
	msgData.Set("Body", message)

	req, _ := http.NewRequest("POST", location, strings.NewReader(msgData.Encode()))
	req.SetBasicAuth(cfg.AccountSid, cfg.AuthToken)
	req.Header.Add("Accept", "application/json")
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")

	resp, _ := client.Do(req)
	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		data := make(map[string]interface{})
		if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
			return err
		} else {
			fmt.Println(data["sid"])
		}
	} else {
		fmt.Println(resp.Status)
	}
	return nil
}

func SendAllSMS(client *http.Client, cfg *SMSConfig, message string) error {
	var firstErr error
	for i := range cfg.To {
		if err := SendSMS(client, cfg, cfg.To[i], message); err != nil {
			if firstErr == nil {
				firstErr = err
			}
			fmt.Printf("ERROR: %v", err)
		}
	}
	return firstErr
}
