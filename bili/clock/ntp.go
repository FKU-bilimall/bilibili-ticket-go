package clock

import (
	"bilibili-ticket-go/bili/models/api"
	"github.com/beevik/ntp"
	"github.com/imroc/req/v3"
	"time"
)

func GetBilibiliClockOffset() (time.Duration, error) {
	res, err := req.R().Get("https://api.live.bilibili.com/xlive/open-interface/v1/rtc/getTimestamp")
	if err != nil {
		return 0, err
	}
	var r api.MainApiDataRoot[api.RTCTimestamp]
	err = res.Unmarshal(&r)
	if err != nil {
		return 0, err
	}
	t := res.TraceInfo()
	NO := t.FirstResponseTime + t.ResponseTime
	return time.UnixMilli(r.Data.Microtime).Add(NO).Sub(time.Now()), nil // Placeholder value, replace with actual logic to get time offset
}

func GetAliyunClockOffset() (time.Duration, error) {
	q, err := ntp.Query("ntp.aliyun.com")
	if err != nil {
		return 0, err
	}
	return q.ClockOffset, nil
}
