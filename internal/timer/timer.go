package timer

import "time"

// Stop does not close the channel, to prevent a read from the channel succeeding incorrectly.
// To ensure the channel is empty after a call to Stop, check the return value and drain the
// channel. For example, assuming the program has not received from t.C already:
// ```
//
//	if !t.Stop() {
//		<-t.C
//	}
//
// ```
func StopTimer(timer *time.Timer) {
	if !timer.Stop() {
		select {
		case <-timer.C:
		default:
		}
	}
}
