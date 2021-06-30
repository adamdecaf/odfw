package twilio

import (
	"net/http"
	"time"
)

var (
	TwilioHttpClient = &http.Client{
		Timeout: 10 * time.Second,
	}
)
