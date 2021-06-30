package exec

import (
	"time"

	"github.com/adamdecaf/odfw/internal/stocksched"
	"github.com/adamdecaf/odfw/internal/twilio"
)

type Config struct {
	Schedule      Schedule
	Notifications Notifications
	Species       Species
}

type Schedule struct {
	Interval time.Duration
	Windows  []string
}

type Notifications struct {
	SMS twilio.SMSConfig
}

type Species struct {
	Trout Trout
}

type Trout struct {
	Zone stocksched.Zone // e.g. "Willamette"
}
