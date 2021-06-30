package exec

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestBasic__justAfter(t *testing.T) {
	loc, _ := time.LoadLocation("America/Los_Angeles")
	now := time.Date(2021, time.March, 21, 9, 12, 0, 0, loc)

	sched := Schedule{
		Interval: 10 * time.Minute,
		Windows: []string{
			"09:10",
		},
	}

	// Happy path
	require.True(t, justAfter(now, sched))

	sched.Windows = []string{"09:10"}
	require.True(t, justAfter(now, sched))

	// Sad paths
	sched.Windows = []string{"09:21"}
	require.False(t, justAfter(now, sched))

	sched.Windows = []string{"09:00"}
	require.False(t, justAfter(now, sched))

	sched.Windows = []string{"08:21"}
	require.False(t, justAfter(now, sched))

	sched.Windows = []string{"15:47"}
	require.False(t, justAfter(now, sched))
}
