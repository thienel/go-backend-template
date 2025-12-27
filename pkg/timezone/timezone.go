package timezone

import "time"

// VietnamLocation is the Vietnam timezone (UTC+7)
var VietnamLocation *time.Location

func init() {
	loc, err := time.LoadLocation("Asia/Ho_Chi_Minh")
	if err != nil {
		loc = time.FixedZone("UTC+7", 7*60*60)
	}
	VietnamLocation = loc
}

// Now returns the current time in Vietnam timezone
func Now() time.Time {
	return time.Now().In(VietnamLocation)
}

// ToVietnam converts a time to Vietnam timezone
func ToVietnam(t time.Time) time.Time {
	return t.In(VietnamLocation)
}
