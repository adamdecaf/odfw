package stocksched

import (
	"net/http"
	"time"
)

var (
	ODFWHttpClient = &http.Client{
		Timeout: 10 * time.Second,
	}
)
