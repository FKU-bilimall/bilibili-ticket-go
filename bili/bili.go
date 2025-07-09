package bili

import (
	"bilibili-ticket-go/bili/models/response"
	"bilibili-ticket-go/utils"
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"github.com/imroc/req/v3"
	"net/http"
	"net/url"
	"sort"
	"strings"
)

const appKey = "1d8b6e7d45233436"
const appSec = "560c52ccd288fed045859ed18bffd973"
const stringVer = "8.51.0"
const buildVer = 8510500
const model = "SM-S9080"

type Client struct {
	http         *req.Client
	cookie       http.CookieJar
	buvid        string
	refreshToken string
}

func GetNewClient(jar http.CookieJar, buvid string) *Client {
	var id = buvid
	if id == "" {
		id = utils.GenerateBUVID()
	}
	logger.Debugf("Client BUVID: %s", id)
	c := req.C()
	if jar != nil {
		c.SetCookieJar(jar)
	}
	c.SetUserAgent(fmt.Sprintf(
		`Mozilla/5.0 BiliDroid/%s (bbcallen@gmail.com) os/android model/%s mobi_app/android build/%d channel/bili innerVer/%d osVer/12 network/2`,
		stringVer, model, buildVer, buildVer,
	)).
		SetCommonHeader("app-key", "android64").
		SetCommonHeader("buvid", id).
		SetTLSFingerprintAndroid().
		ImpersonateChrome()
	return &Client{
		http:   c,
		buvid:  id,
		cookie: jar,
	}
}

var logger = utils.GetLogger("bili-client", nil)

// GetQRCodeUrlAndKey retrieves the QR code URL and key for Bilibili login.
func (c *Client) GetQRCodeUrlAndKey() (error, *response.GetQRLoginKeyPayload) {
	res, err := c.http.R().Get("https://passport.bilibili.com/x/passport-login/web/qrcode/generate?source=main-fe-header")
	if err != nil {
		return err, nil
	}
	var r response.Root[response.GetQRLoginKeyPayload]
	err = res.Unmarshal(&r)
	if err != nil {
		return err, nil
	}
	return nil, &r.Data
}

func (c *Client) GetBUVID() string {
	return c.buvid
}

func (c *Client) GetQRLoginState(qrcodeKey string) (error, *response.VerifyQRLoginStatePayload) {
	res, err := c.http.R().SetQueryParam("qrcode_key", qrcodeKey).Get("https://passport.bilibili.com/x/passport-login/web/qrcode/poll")
	if err != nil {
		return err, nil
	}
	var r response.Root[response.VerifyQRLoginStatePayload]
	err = res.Unmarshal(&r)
	if err != nil {
		return err, nil
	}
	c.refreshToken = r.Data.RefreshToken
	return nil, &r.Data
}

func (c *Client) AppSignWithQueries(req req.Request) req.Request {
	query := req.URL.Query()
	var keys []string
	for key := range query {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	var sortedParams []string
	for _, key := range keys {
		values := query[key]
		for _, value := range values {
			sortedParams = append(sortedParams, fmt.Sprintf("%s=%s", key, value))
		}
	}
	sortedQueryString := strings.Join(sortedParams, "&")
	encoded := url.QueryEscape(sortedQueryString) + appSec
	sign := hex.EncodeToString(md5.New().Sum([]byte(encoded)))
	logger.Debugf("Queries: %s, App Sign: %s", sortedQueryString, sign)
	req.URL.Query().Set("sign", sign)
	return req
}

func (c *Client) GetLoginStatus() (error, *response.GetLoginInfoPayload) {
	res, err := c.http.R().Get("https://api.bilibili.com/x/web-interface/nav")
	if err != nil {
		return err, nil
	}
	var r response.Root[response.GetLoginInfoPayload]
	err = res.Unmarshal(&r)
	if err != nil {
		return err, nil
	}
	return nil, &r.Data
}
