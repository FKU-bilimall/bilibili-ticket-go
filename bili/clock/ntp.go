package clock

import (
	"bilibili-ticket-go/bili/models/api"
	"time"

	"github.com/beevik/ntp"
	"github.com/imroc/req/v3"
)

func GetBilibiliClockOffset() (time.Duration, error) {
	res, err := req.R().EnableTrace().Get("https://api.live.bilibili.com/xlive/open-interface/v1/rtc/getTimestamp")
	if err != nil {
		return 0, err
	}
	now := time.Now()
	var r api.MainApiDataRoot[api.RTCTimestamp]
	err = res.Unmarshal(&r)
	if err != nil {
		return 0, err
	}
	t := res.TraceInfo()
	NO := t.FirstResponseTime + t.ResponseTime
	return time.UnixMilli(r.Data.Microtime).Add(NO).Sub(now), nil
}

func GetAliyunClockOffset() (time.Duration, error) {
	q, err := ntp.Query("ntp.aliyun.com")
	if err != nil {
		return 0, err
	}
	return q.ClockOffset, nil
}
