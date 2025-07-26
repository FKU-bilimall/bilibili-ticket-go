package bili

import (
	"bilibili-ticket-go/bili/models/api"
	"bilibili-ticket-go/global"
	"bilibili-ticket-go/utils"
	"fmt"
	"github.com/imroc/req/v3"
	"net/http"
)

const model = "SM-S9080"

// 2025/07/19 test project_id=103601

type Client struct {
	http         *req.Client
	cookie       http.CookieJar
	buvid        string
	refreshToken string
	appVersion   *api.BiliAppVersionStruct
	buvidfp      string
	webglfp      string
	canvasfp     string
	wbi          *wbiKey
}

var logger = utils.GetLogger(global.GetLogger(), "bili-client", nil)

func GetNewClient(jar http.CookieJar, buvid string, refreshToken string, fingerprint string) *Client {
	var id = buvid
	if id == "" {
		id = utils.GenerateXUBUVID()
	}
	var fp = fingerprint
	if fp == "" {
		fp = utils.GetFpLocal(id, model, "")
	}
	logger.Debugf("Client BUVID: %s", id)
	c := req.C().EnableDebugLog()
	err, ver := getAppLatestVersion()
	if err != nil {
		return nil
	}
	biliClient := &Client{
		http:         c,
		buvid:        id,
		cookie:       jar,
		appVersion:   ver,
		refreshToken: refreshToken,
		buvidfp:      fingerprint,
		webglfp:      utils.RandomString("0123456789abcdef", 32),
		canvasfp:     utils.RandomString("0123456789abcdef", 32),
		wbi:          &wbiKey{},
	}
	c.SetLogger(logger)
	if jar != nil {
		c.SetCookieJar(jar)
	}
	c.SetTLSFingerprintAndroid().
		ImpersonateChrome()
	c.SetCommonCookies()
	c.WrapRoundTripFunc(func(rt req.RoundTripper) req.RoundTripFunc {
		return func(req *req.Request) (resp *req.Response, err error) {
			//Before
			var cookies []*http.Cookie
			var ua = fmt.Sprintf(
				`Mozilla/5.0 BiliDroid/%s (bbcallen@gmail.com) os/android model/%s mobi_app/android build/%d channel/bili innerVer/%d osVer/12 network/2`,
				biliClient.appVersion.Version, model, biliClient.appVersion.Build, biliClient.appVersion.Build,
			)
			copy(cookies, req.Cookies)
			if req.URL.Host == "show.bilibili.com" {
				req.SetHeader("x-requested-with", "tv.danmaku.bili")
				ua = fmt.Sprintf(
					`Mozilla/5.0 (Linux; Android 12; %s; wv) AppleWebKit/537.36 (KHTML, like Gecko) Version/4.0 Chrome/101.0.4951.61 Safari/537.36 BiliApp/%d mobi_app/android isNotchWindow/0 NotchHeight=24 mallVersion/%d mVersion/312 disable_rcmd/0 magent/BILI_H5_ANDROID_12_%s_%d`,
					model, biliClient.appVersion.Build, biliClient.appVersion.Build, biliClient.appVersion.Version, biliClient.appVersion.Build,
				)
				cookies = append(cookies,
					&http.Cookie{
						Name:  "Buvid",
						Value: biliClient.buvid,
					},
					&http.Cookie{
						Name:  "deviceFingerprint",
						Value: biliClient.buvidfp,
					},
					&http.Cookie{
						Name:  "buvid_fp",
						Value: biliClient.buvidfp,
					},
					&http.Cookie{
						Name:  "kfcFrom",
						Value: "mall_home_searchhis",
					},
					&http.Cookie{
						Name:  "from",
						Value: "mall_search_discovery",
					},
					&http.Cookie{
						Name:  "kfcSource",
						Value: "bilibiliapp",
					},
					&http.Cookie{
						Name:  "mSource",
						Value: "bilibiliapp",
					},
					&http.Cookie{
						Name:  "feSign",
						Value: getFeSign(ua, biliClient.canvasfp, biliClient.webglfp),
					},
					&http.Cookie{
						Name:  "screenInfo",
						Value: screenInfo,
					},
				)
			}
			if req.Headers.Get("Referer") != "" {
				req.SetHeader("Referer", "https://www.bilibili.com/")
			}
			req.SetHeader("User-Agent", ua)
			req.SetCookies(cookies...)
			resp, err = rt.RoundTrip(req)
			//After
			return resp, err
		}
	})
	return biliClient
}

func (c *Client) GetQRCodeUrlAndKey() (error, *api.GetQRLoginKeyStruct) {
	res, err := c.http.R().Get("https://passport.bilibili.com/x/passport-login/web/qrcode/generate?source=main-fe-header")
	if err != nil {
		return err, nil
	}
	var r api.MainApiDataRoot[api.GetQRLoginKeyStruct]
	err = res.Unmarshal(&r)
	if err != nil {
		return err, nil
	}
	if err = r.CheckValid(); err != nil {
		return err, nil
	}
	return nil, &r.Data
}

func (c *Client) GetBUVID() string {
	return c.buvid
}

func (c *Client) GetFingerprint() string {
	return c.buvidfp
}

func (c *Client) GetRefreshToken() string {
	return c.refreshToken
}

func (c *Client) GetQRLoginState(qrcodeKey string) (error, *api.VerifyQRLoginStateStruct) {
	res, err := c.http.R().SetQueryParam("qrcode_key", qrcodeKey).Get("https://passport.bilibili.com/x/passport-login/web/qrcode/poll")
	if err != nil {
		return err, nil
	}
	var r api.MainApiDataRoot[api.VerifyQRLoginStateStruct]
	err = res.Unmarshal(&r)
	if err != nil {
		return err, nil
	}
	if r.Data.Code == 0 {
		c.refreshToken = r.Data.RefreshToken
		err := c.getBuvid34AndBnut()
		if err != nil {
			logger.Warnf("getBuvid34AndBnut() err: %v", err)
		}
		err, _ = c.RefreshNewBiliTicket()
		if err != nil {
			logger.Warnf("RefreshNewBiliTicket() err: %v", err)
		}
	}
	return nil, &r.Data
}

func (c *Client) GetLoginStatus() (error, *api.GetLoginInfoStruct) {
	res, err := c.http.R().Get("https://api.bilibili.com/x/web-interface/nav")
	if err != nil {
		return err, nil
	}
	var r api.MainApiDataRoot[api.GetLoginInfoStruct]
	err = res.Unmarshal(&r)
	if err != nil {
		return err, nil
	}
	return nil, &r.Data
}
