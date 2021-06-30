package exec

import (
	"context"
	"fmt"
	"time"

	"github.com/adamdecaf/odfw/internal/stocksched"
)

func Basic(ctx context.Context, cfg *Config) {
	ticker := time.NewTicker(cfg.Schedule.Interval)
	for {
		select {
		case <-ticker.C:
			if err := basicTick(cfg); err != nil {
				fmt.Printf("ERROR: %v", err)
			}

		case <-ctx.Done():
			ticker.Stop()
		}
	}

}

func basicTick(cfg *Config) error {
	if !justAfter(time.Now(), cfg.Schedule) {
		return nil
	}

	// Download trout counts
	troutCounts, err := stocksched.LoadTroutCounts(stocksched.ODFWHttpClient, cfg.Species.Trout.Zone)
	if err != nil {
		fmt.Printf("ERROR: %v", err)
	}
	fmt.Printf("COUNTS: %#v\n", troutCounts)

	// Send our SMS messages
	// SendAllSMS(client *http.Client, cfg *SMSConfig, message string) error

	return nil
}

func justAfter(now time.Time, sched Schedule) bool {
	for i := range sched.Windows {
		min, max := minmax(now, sched.Windows[i], sched.Interval)
		if min.Before(now) && max.After(now) {
			return true
		}
	}
	return false
}

func minmax(now time.Time, window string, interval time.Duration) (time.Time, time.Time) {
	min, _ := time.Parse("MST 2006-01-02 15:04", fmt.Sprintf("%s %s", now.Format("MST 2006-01-02"), window))
	return min, min.Add(interval)
}
