package clock

import (
	api2 "bilibili-ticket-go/models/bili/api"
	"time"

	"github.com/beevik/ntp"
	"github.com/imroc/req/v3"
)

func GetBilibiliClockOffset() (time.Duration, error) {
	now := time.Now()
	res, err := req.R().EnableTrace().Get("https://api.live.bilibili.com/xlive/open-interface/v1/rtc/getTimestamp")
	if err != nil {
		return 0, err
	}
	var r api2.MainApiDataRoot[api2.RTCTimestamp]
	err = res.Unmarshal(&r)
	if err != nil {
		return 0, err
	}
	t := res.TraceInfo()
	NetworkOffset := t.FirstResponseTime + t.ResponseTime
	return time.UnixMilli(r.Data.Microtime).Add(-NetworkOffset).Sub(now), nil
}

// GetNTPClockOffset queries the given NTP server and returns the clock offset.
// Recommended NTP server: ntp.aliyun.com
func GetNTPClockOffset(ntpServerAddr string) (time.Duration, error) {
	q, err := ntp.Query(ntpServerAddr)
	if err != nil {
		return 0, err
	}
	return q.ClockOffset, nil
}
