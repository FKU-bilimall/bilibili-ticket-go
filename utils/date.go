package utils

import "time"

func IsNextDayInCST(from time.Time, target time.Time) bool {
	loc, _ := time.LoadLocation("Asia/Shanghai")
	now := from.In(loc)
	afterHour := target.In(loc)

	return now.Format("20060102") != afterHour.Format("20060102")
}
