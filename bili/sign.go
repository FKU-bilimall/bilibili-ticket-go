package bili

import (
	"bilibili-ticket-go/bili/models/response"
	"bilibili-ticket-go/utils"
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"
)

type wbiKey struct {
	mixin  string
	expire time.Time
}

var mixinKeyEncTab = [64]int{46, 47, 18, 2, 53, 8, 23, 32, 15, 50, 10, 31, 58, 3, 45, 35, 27, 43, 5, 49, 33, 9, 42, 19, 29, 28, 14, 39, 12, 38, 41, 13, 37, 48, 7, 16, 24, 55, 40, 61, 26, 17, 0, 1, 60, 51, 30, 4, 22, 25, 54, 21, 56, 59, 6, 63, 57, 62, 11, 36, 20, 34, 44, 52}

const appKey = "1d8b6e7d45233436"
const appSec = "560c52ccd288fed045859ed18bffd973"

func (w *wbiKey) isExpired() bool {
	return time.Now().After(w.expire)
}

func (c *Client) refreshWbiToken() error {
	res, err := c.http.R().Get("https://api.bilibili.com/x/web-interface/nav")
	if err != nil {
		return err
	}
	var r response.DataRoot[response.WbiStruct]
	err = res.Unmarshal(&r)
	if err != nil {
		return err
	}
	cl := utils.GetFileNameWithoutExt(r.Data.WbiImg.ImgUrl) + utils.GetFileNameWithoutExt(r.Data.WbiImg.SubUrl)
	var builder strings.Builder
	for _, index := range mixinKeyEncTab {
		builder.WriteByte(cl[index])
	}
	key := builder.String()[0:32]
	c.wbi.mixin = key
	now := time.Now()
	expired := now.Add(1 * time.Hour)
	tomorrow := now.Add(24 * time.Hour).Truncate(24 * time.Hour)
	if utils.IsNextDayInCST(now, expired) {
		expired = tomorrow
	}
	c.wbi = &wbiKey{
		mixin:  key,
		expire: expired,
	}
	return nil
}

func (c *Client) GetSignedParameterWithAbi(forceUpdate bool, u *url.URL) error {
	if c.wbi == nil || c.wbi.isExpired() || forceUpdate {
		err := c.refreshWbiToken()
		if err != nil {
			return err
		}
	}
	values := u.Query()
	values.Del("w_rid")
	values.Set("wts", fmt.Sprintf("%d", time.Now().Unix()))
	wbi := md5.Sum([]byte(values.Encode() + c.wbi.mixin))
	logger.Debugf("Queries: %s, WBI Sign: %s", values.Encode(), hex.EncodeToString(wbi[:]))
	values.Set("w_rid", hex.EncodeToString(wbi[:]))
	u.RawQuery = values.Encode()
	return nil
}

func (c *Client) getSignedParameterWithApp(u *url.URL) {
	values := u.Query()
	values.Del("sign")
	values.Set("appkey", appKey)
	sign := md5.Sum([]byte(values.Encode() + appSec))
	logger.Debugf("Queries: %s, App Sign: %s", values.Encode(), hex.EncodeToString(sign[:]))
	values.Set("sign", hex.EncodeToString(sign[:]))
	u.RawQuery = values.Encode()
}

func (c *Client) identifyCookieSign() http.Cookie {
	u, _ := url.Parse(fmt.Sprintf("https://example.com/?ts=%d", time.Now().Unix()))
	c.getSignedParameterWithApp(u)
	return http.Cookie{
		Name:  "identify",
		Value: u.Query().Encode(),
	}
}
