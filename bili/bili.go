package bili

import (
	"bilibili-ticket-go/bili/models/response"
	"bilibili-ticket-go/utils"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/hex"
	"encoding/pem"
	"errors"
	"fmt"
	"github.com/imroc/req/v3"
	"net/http"
	"net/url"
	"regexp"
	"time"
)

const model = "SM-S9080"

// 2025/07/19 test project_id=103601

type Client struct {
	http         *req.Client
	cookie       http.CookieJar
	buvid        string
	refreshToken string
	appVersion   *response.BiliAppVersionStruct
	buvidfp      string
	webglfp      string
	canvasfp     string
	wbi          *wbiKey
}

var logger = utils.GetLogger("bili-client", nil)

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

func (c *Client) GetQRCodeUrlAndKey() (error, *response.GetQRLoginKeyStruct) {
	res, err := c.http.R().Get("https://passport.bilibili.com/x/passport-login/web/qrcode/generate?source=main-fe-header")
	if err != nil {
		return err, nil
	}
	var r response.DataRoot[response.GetQRLoginKeyStruct]
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

func (c *Client) GetQRLoginState(qrcodeKey string) (error, *response.VerifyQRLoginStateStruct) {
	res, err := c.http.R().SetQueryParam("qrcode_key", qrcodeKey).Get("https://passport.bilibili.com/x/passport-login/web/qrcode/poll")
	if err != nil {
		return err, nil
	}
	var r response.DataRoot[response.VerifyQRLoginStateStruct]
	err = res.Unmarshal(&r)
	if err != nil {
		return err, nil
	}
	c.refreshToken = r.Data.RefreshToken
	if r.Code == 0 {
		err := c.getBuvid34AndBnut()
		if err != nil {
			return err, nil
		}
	}
	return nil, &r.Data
}

func (c *Client) GetLoginStatus() (error, *response.GetLoginInfoStruct) {
	res, err := c.http.R().Get("https://api.bilibili.com/x/web-interface/nav")
	if err != nil {
		return err, nil
	}
	var r response.DataRoot[response.GetLoginInfoStruct]
	err = res.Unmarshal(&r)
	if err != nil {
		return err, nil
	}
	return nil, &r.Data
}

func (c *Client) CheckAndUpdateCookie() (error, bool) {
	logger.Debug("Checking and updating cookie...")
	err, st := c.GetLoginStatus()
	if err != nil {
		return err, false
	}
	if !st.Login {
		logger.Debugf("User is not logged in, cannot refresh cookie.\n")
		return nil, false
	}
	err, stat := c.checkNeedRefresh()
	if err != nil || !stat {
		if !stat {
			logger.Debug("No need to refresh cookie.")
		}
		return err, false
	}
	oldCSRF := c.getCSRFFromCookie()
	logger.Debugf("Old CSRF Token: %s", oldCSRF)
	oldRefreshToken := c.refreshToken
	cp, err := getCorrespondPath(time.Now().UnixMilli())
	if err != nil {
		return err, false
	}
	err, CSRFKey := c.getRefreshCSRF(cp)
	if err != nil {
		return err, false
	}
	logger.Debugf("CSRF Key: %s", CSRFKey)
	err, newRefreshToken := c.refreshCookie(oldCSRF, CSRFKey, oldRefreshToken)
	if err != nil {
		return err, false
	}
	logger.Debugf("New Refresh Token: %s", newRefreshToken)
	c.refreshToken = newRefreshToken
	newCSRF := c.getCSRFFromCookie()
	logger.Debugf("New CSRF Token: %s", newCSRF)
	err = c.setPreviousCookieInvalid(newCSRF, oldRefreshToken)
	if err != nil {
		return err, false
	}
	return nil, true
}

func (c *Client) getBuvid34AndBnut() error {
	_, err := c.http.R().Head("https://www.bilibili.com/")
	if err != nil {
		return err
	}
	res, err := c.http.R().Get("https://api.bilibili.com/x/frontend/finger/spi")
	var r response.DataRoot[response.GetBVUID34Struct]
	err = res.Unmarshal(&r)
	if err != nil {
		return err
	}
	parsedURL, _ := url.Parse("https://www.bilibili.com/")
	c.cookie.SetCookies(parsedURL, []*http.Cookie{
		{
			Name:        "buvid3",
			Value:       r.Data.BVUID3,
			Quoted:      false,
			Path:        "/",
			Domain:      "bilibili.com",
			Expires:     time.Time{},
			RawExpires:  "",
			MaxAge:      60 * 60 * 24 * 365,
			Secure:      false,
			HttpOnly:    false,
			SameSite:    0,
			Partitioned: false,
			Raw:         "",
			Unparsed:    nil,
		},
		{
			Name:        "buvid4",
			Value:       r.Data.BVUID4,
			Quoted:      false,
			Path:        "/",
			Domain:      "bilibili.com",
			Expires:     time.Time{},
			RawExpires:  "",
			MaxAge:      60 * 60 * 24 * 365,
			Secure:      false,
			HttpOnly:    false,
			SameSite:    0,
			Partitioned: false,
			Raw:         "",
			Unparsed:    nil,
		},
	})
	return nil
}

func (c *Client) checkNeedRefresh() (error, bool) {
	res, err := c.http.R().Get("https://passport.bilibili.com/x/passport-login/web/cookie/info")
	if err != nil {
		return err, false
	}
	var r response.DataRoot[response.NeedRefreshStruct]
	err = res.Unmarshal(&r)
	if err != nil {
		return err, false
	}
	if err = r.CheckValid(); err != nil {
		return err, false
	}
	return nil, r.Data.NeedRefresh
}

func (c *Client) getRefreshCSRF(correspondPath string) (error, string) {
	res, err := c.http.R().Get(fmt.Sprintf("https://www.bilibili.com/correspond/1/%s", correspondPath))
	if err != nil {
		return err, ""
	}
	s := res.String()
	re := regexp.MustCompile(`<div id="1-name">(.*?)</div>`)
	matches := re.FindStringSubmatch(s)

	if len(matches) > 1 {
		content := matches[1]
		return nil, content
	} else {
		return errors.New("cannot Parser RefreshToken From HTML"), ""
	}
}

func (c *Client) refreshCookie(csrf string, refreshCsrfToken string, refreshToken string) (error, string) {
	res, err := c.http.R().SetFormData(map[string]string{
		"refresh_token": refreshToken,
		"source":        "main-fe-header",
		"csrf_token":    refreshCsrfToken,
		"csrf":          csrf,
	}).Post("https://passport.bilibili.com/x/passport-login/web/cookie/refresh")
	if err != nil {
		return err, ""
	}
	var r response.DataRoot[response.RefreshTokenStruct]
	err = res.Unmarshal(&r)
	if err != nil {
		return err, ""
	}
	if err = r.CheckValid(); err != nil {
		return err, ""
	}
	return nil, r.Data.RefreshToken
}

func (c *Client) setPreviousCookieInvalid(newCsrf string, oldRefreshToken string) error {
	res, err := c.http.R().SetFormData(map[string]string{
		"refresh_token": oldRefreshToken,
		"csrf":          newCsrf,
	}).Post("https://passport.bilibili.com/x/passport-login/web/cookie/refresh")
	if err != nil {
		return err
	}
	var r response.DataRoot[interface{}]
	err = res.Unmarshal(&r)
	if err != nil {
		return err
	}
	if err = r.CheckValid(); err != nil {
		return err
	}
	return nil
}

func (c *Client) getCSRFFromCookie() string {
	parsedURL, _ := url.Parse("https://www.bilibili.com/")
	for _, cookie := range c.cookie.Cookies(parsedURL) {
		if cookie.Name == "bili_jct" {
			return cookie.Value
		}
	}
	return ""
}

func (c *Client) RefreshNewBiliTicket() (error, bool) {
	parsedURL, _ := url.Parse("https://www.bilibili.com/")
	for _, cookie := range c.cookie.Cookies(parsedURL) {
		if cookie.Name == "bili_ticket" {
			if cookie.Expires.Sub(time.Now()) >= 1*time.Minute {
				return nil, false
			}
		}
	}
	ts := time.Now().UnixMilli()
	hexsign := utils.HmacSha256ToHex("XgwSnGZ1p", fmt.Sprintf("ts%d", ts))
	res, err := c.http.R().SetFormData(map[string]string{
		"key_id":      "ec02",
		"hexsign":     hexsign,
		"context[ts]": fmt.Sprintf("%d", ts),
	}).Post("https://api.bilibili.com/bapis/bilibili.api.ticket.v1.Ticket/GenWebTicket")
	if err != nil {
		return err, false
	}
	var r response.DataRoot[response.BiliTicketStruct]
	err = res.Unmarshal(&r)
	if err != nil {
		return err, false
	}
	if err = r.CheckValid(); err != nil {
		return err, false
	}
	c.cookie.SetCookies(parsedURL, []*http.Cookie{
		{
			Name:        "bili_ticket",
			Value:       r.Data.Ticket,
			Quoted:      false,
			Path:        "/",
			Domain:      "bilibili.com",
			Expires:     time.Time{},
			RawExpires:  "",
			MaxAge:      r.Data.TTL,
			Secure:      false,
			HttpOnly:    false,
			SameSite:    0,
			Partitioned: false,
			Raw:         "",
			Unparsed:    nil,
		},
	})
	return nil, false
}

func getAppLatestVersion() (error, *response.BiliAppVersionStruct) {
	res, err := req.Get("https://app.bilibili.com/x/v2/version?mobi_app=android")
	if err != nil {
		return err, nil
	}
	var r response.DataRoot[[]response.BiliAppVersionStruct]
	err = res.Unmarshal(&r)
	if err != nil {
		return err, nil
	}
	if err = r.CheckValid(); err != nil {
		return err, nil
	}
	return nil, &r.Data[0]
}

func getCorrespondPath(ts int64) (string, error) {
	const publicKeyPEM = `
-----BEGIN PUBLIC KEY-----
MIGfMA0GCSqGSIb3DQEBAQUAA4GNADCBiQKBgQDLgd2OAkcGVtoE3ThUREbio0Eg
Uc/prcajMKXvkCKFCWhJYJcLkcM2DKKcSeFpD/j6Boy538YXnR6VhcuUJOhH2x71
nzPjfdTcqMz7djHum0qSZA0AyCBDABUqCrfNgCiJ00Ra7GmRj+YCK1NJEuewlb40
JNrRuoEUXpabUzGB8QIDAQAB
-----END PUBLIC KEY-----
`
	pubKeyBlock, _ := pem.Decode([]byte(publicKeyPEM))
	hash := sha256.New()
	random := rand.Reader
	msg := []byte(fmt.Sprintf("refresh_%d", ts))
	var pub *rsa.PublicKey
	pubInterface, parseErr := x509.ParsePKIXPublicKey(pubKeyBlock.Bytes)
	if parseErr != nil {
		return "", parseErr
	}
	pub = pubInterface.(*rsa.PublicKey)
	encryptedData, encryptErr := rsa.EncryptOAEP(hash, random, pub, msg, nil)
	if encryptErr != nil {
		return "", encryptErr
	}
	return hex.EncodeToString(encryptedData), nil
}
