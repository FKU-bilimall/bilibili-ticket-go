package bili

import (
	"bilibili-ticket-go/bili/models/api"
	"bilibili-ticket-go/utils/hashs"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/hex"
	"encoding/pem"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"regexp"
	"time"

	"github.com/imroc/req/v3"
)

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
	var r api.MainApiDataRoot[api.GetBVUID34Struct]
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
	var r api.MainApiDataRoot[api.NeedRefreshStruct]
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
		"source":        "main_web",
		"refresh_csrf":  refreshCsrfToken,
		"csrf":          csrf,
	}).Post("https://passport.bilibili.com/x/passport-login/web/cookie/refresh")
	if err != nil {
		return err, ""
	}
	var r api.MainApiDataRoot[api.RefreshTokenStruct]
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
	}).Post("https://passport.bilibili.com/x/passport-login/web/confirm/refresh")
	if err != nil {
		return err
	}
	var r api.MainApiDataRoot[interface{}]
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
			if cookie.Expires.Sub(time.Now()) >= 1*time.Hour {
				return nil, false
			}
		}
	}
	ts := time.Now().Unix()
	hexsign := hash.HmacSha256ToHex("XgwSnGZ1p", fmt.Sprintf("ts%d", ts))
	res, err := c.http.R().SetQueryParams(map[string]string{
		"key_id":      "ec02",
		"hexsign":     hexsign,
		"context[ts]": fmt.Sprintf("%d", ts),
		"csrf":        c.getCSRFFromCookie(),
	}).Post("https://api.bilibili.com/bapis/bilibili.api.ticket.v1.Ticket/GenWebTicket")
	if err != nil {
		return err, false
	}
	var r api.MainApiDataRoot[api.BiliTicketStruct]
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
	return nil, true
}

func getAppLatestVersion() (error, *api.BiliAppVersionStruct) {
	res, err := req.Get("https://app.bilibili.com/x/v2/version?mobi_app=android")
	if err != nil {
		return err, nil
	}
	var r api.MainApiDataRoot[[]api.BiliAppVersionStruct]
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
