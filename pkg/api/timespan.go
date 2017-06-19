package api

import "time"

type TimeSpan struct {
	Days int
	Hours int
	Minutes int
	Seconds int
	Millseconds int
}

func NewTimeSpan1(days int, hours int, minutes int, seconds int, millseconds int) *TimeSpan {
	return &TimeSpan{Days:days,Hours:hours,Minutes:minutes,Seconds:seconds,Millseconds:millseconds}
}

func NewTimeSpan2(hours int, minutes int, seconds int, millseconds int) *TimeSpan {
	return &TimeSpan{Days:0,Hours:hours,Minutes:minutes,Seconds:seconds,Millseconds:millseconds}
}

func NewTimeSpan3(hours int, minutes int, seconds int) *TimeSpan {
	return &TimeSpan{Days:0,Hours:hours,Minutes:minutes,Seconds:seconds,Millseconds:0}
}

func (s *TimeSpan) GetDuration() time.Duration {
	return time.Duration(int64(time.Hour)*int64(s.Days*24 + s.Hours) + int64(time.Minute) * int64(s.Minutes) +
		int64(time.Second) * int64(s.Seconds) +  int64(time.Millisecond) * int64(s.Millseconds) )
}

