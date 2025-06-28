package bili

import (
	"bilibili-ticket-go/bili/models/response"
	"bilibili-ticket-go/utils"
	"github.com/imroc/req/v3"
	cookiejar "github.com/juju/persistent-cookiejar"
)

const appKey = "1d8b6e7d45233436"
const appSec = "560c52ccd288fed045859ed18bffd973"

type Client struct {
	http         *req.Client
	cookie       *cookiejar.Jar
	buvid        string
	refreshToken string
}

func GetNewClient(jar *cookiejar.Jar, buvid string) *Client {
	var id = buvid
	if id == "" {
		id = utils.GenerateBUVID()
	}
	c := req.C()
	c.SetCookieJar(jar)
	c.SetUserAgent("Mozilla/5.0 BiliDroid/8.51.0 (bbcallen@gmail.com) os/android model/SM-S9080 mobi_app/android build/8510500 channel/bili innerVer/8510510 osVer/12 network/2").
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
