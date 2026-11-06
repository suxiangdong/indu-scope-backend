package wechat

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/glog"
)

const (
	defaultAPIBaseURL          = "https://api.weixin.qq.com"
	accessTokenRefreshInterval = 30 * time.Minute
)

type LoginSession struct {
	OpenID     string
	SessionKey string
}

type clientConfig struct {
	AppID      string
	AppSecret  string
	APIBaseURL string
}

type code2SessionResponse struct {
	OpenID     string `json:"openid"`
	SessionKey string `json:"session_key"`
	ErrCode    int    `json:"errcode"`
	ErrMsg     string `json:"errmsg"`
}

type accessTokenResponse struct {
	AccessToken string `json:"access_token"`
	ExpiresIn   int    `json:"expires_in"`
	ErrCode     int    `json:"errcode"`
	ErrMsg      string `json:"errmsg"`
}

// The WeChat app access token is shared by all requests handled by this
// process. It is an app-level credential, not a user session token, and must
// not be fetched again for every phone authorization.
var accessTokenCache struct {
	sync.Mutex
	appID       string
	apiBaseURL  string
	token       string
	expiresAt   time.Time
	refreshedAt time.Time
}

type phoneNumberResponse struct {
	PhoneInfo struct {
		PhoneNumber string `json:"phoneNumber"`
		PurePhone   string `json:"purePhoneNumber"`
		// WeChat has returned countryCode as both 86 and "86" across
		// environments. The phone-login flow does not use it, so retain the
		// raw JSON and accept either representation.
		CountryCode json.RawMessage `json:"countryCode"`
	} `json:"phone_info"`
	ErrCode int    `json:"errcode"`
	ErrMsg  string `json:"errmsg"`
}

// ResolveLoginCode exchanges wx.login's temporary code for the OpenID that is
// kept in the server-side anonymous session. The session_key never leaves the
// backend.
func ResolveLoginCode(ctx context.Context, code string) (LoginSession, error) {
	config, err := loadConfig(ctx)
	if err != nil {
		return LoginSession{}, err
	}
	params := url.Values{
		"appid":      []string{config.AppID},
		"secret":     []string{config.AppSecret},
		"js_code":    []string{strings.TrimSpace(code)},
		"grant_type": []string{"authorization_code"},
	}
	var response code2SessionResponse
	if err := getJSON(ctx, config.APIBaseURL+"/sns/jscode2session?"+params.Encode(), &response); err != nil {
		return LoginSession{}, err
	}
	if response.ErrCode != 0 || response.OpenID == "" {
		return LoginSession{}, fmt.Errorf("wechat code2session failed: %d %s", response.ErrCode, response.ErrMsg)
	}
	return LoginSession{OpenID: response.OpenID, SessionKey: response.SessionKey}, nil
}

// ResolvePhoneCode calls the official getuserphonenumber API. The component
// code is single-use and is exchanged only on the backend.
func ResolvePhoneCode(ctx context.Context, code string) (string, error) {
	config, err := loadConfig(ctx)
	if err != nil {
		return "", err
	}
	accessToken, err := getAccessToken(ctx, config)
	if err != nil {
		return "", err
	}
	payload, err := json.Marshal(struct {
		Code string `json:"code"`
	}{Code: strings.TrimSpace(code)})
	if err != nil {
		return "", fmt.Errorf("encode wechat phone request: %w", err)
	}
	endpoint := config.APIBaseURL + "/wxa/business/getuserphonenumber?access_token=" + url.QueryEscape(accessToken)
	request, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewReader(payload))
	if err != nil {
		return "", err
	}
	request.Header.Set("Content-Type", "application/json")
	responseBody, err := (&http.Client{Timeout: 8 * time.Second}).Do(request)
	if err != nil {
		return "", fmt.Errorf("request wechat phone API: %w", err)
	}
	defer responseBody.Body.Close()
	var response phoneNumberResponse
	if err := json.NewDecoder(responseBody.Body).Decode(&response); err != nil {
		return "", fmt.Errorf("decode wechat phone API response: %w", err)
	}
	if response.ErrCode != 0 || response.PhoneInfo.PurePhone == "" {
		return "", fmt.Errorf("wechat phone API failed: %d %s", response.ErrCode, response.ErrMsg)
	}
	return response.PhoneInfo.PurePhone, nil
}

func loadConfig(ctx context.Context) (clientConfig, error) {
	appID := strings.TrimSpace(configString(ctx, "wechat.app_id"))
	appSecret := strings.TrimSpace(configString(ctx, "wechat.app_secret"))
	if appID == "" || appSecret == "" {
		return clientConfig{}, fmt.Errorf("wechat.app_id and wechat.app_secret are required")
	}
	baseURL := strings.TrimRight(strings.TrimSpace(configString(ctx, "wechat.api_base_url")), "/")
	if baseURL == "" {
		baseURL = defaultAPIBaseURL
	}
	return clientConfig{AppID: appID, AppSecret: appSecret, APIBaseURL: baseURL}, nil
}

func configString(ctx context.Context, key string) string {
	value, err := g.Cfg().Get(ctx, key)
	if err != nil || value == nil {
		return ""
	}
	return value.String()
}

// WarmupAccessToken loads the app-level token during server startup. The token
// is still refreshed lazily every 30 minutes or after expiry by getAccessToken.
func WarmupAccessToken(ctx context.Context) error {
	config, err := loadConfig(ctx)
	if err != nil {
		return err
	}
	_, err = getAccessToken(ctx, config)
	return err
}

func getAccessToken(ctx context.Context, config clientConfig) (string, error) {
	accessTokenCache.Lock()
	defer accessTokenCache.Unlock()

	now := time.Now()
	// Refresh the process-wide token every 30 minutes. The expiry check is kept
	// as a safety net in case WeChat returns a shorter lifetime than expected.
	if accessTokenCache.token != "" &&
		accessTokenCache.appID == config.AppID &&
		accessTokenCache.apiBaseURL == config.APIBaseURL &&
		now.Sub(accessTokenCache.refreshedAt) < accessTokenRefreshInterval &&
		now.Before(accessTokenCache.expiresAt) {
		return accessTokenCache.token, nil
	}

	return fetchAccessToken(ctx, config)
}

func fetchAccessToken(ctx context.Context, config clientConfig) (string, error) {
	params := url.Values{
		"grant_type": []string{"client_credential"},
		"appid":      []string{config.AppID},
		"secret":     []string{config.AppSecret},
	}
	var response accessTokenResponse
	if err := getJSON(ctx, config.APIBaseURL+"/cgi-bin/token?"+params.Encode(), &response); err != nil {
		return "", err
	}
	if response.ErrCode != 0 || response.AccessToken == "" {
		return "", fmt.Errorf("wechat access token failed: %d %s", response.ErrCode, response.ErrMsg)
	}
	accessTokenCache.appID = config.AppID
	accessTokenCache.apiBaseURL = config.APIBaseURL
	accessTokenCache.token = response.AccessToken
	accessTokenCache.refreshedAt = time.Now()
	if response.ExpiresIn > 0 {
		accessTokenCache.expiresAt = time.Now().Add(time.Duration(response.ExpiresIn) * time.Second)
	} else {
		// Treat a malformed lifetime as immediately expiring rather than
		// retaining an unusable token indefinitely.
		accessTokenCache.expiresAt = time.Now()
	}
	return response.AccessToken, nil
}

func getJSON(ctx context.Context, endpoint string, target any) error {
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return err
	}
	response, err := (&http.Client{Timeout: 8 * time.Second}).Do(request)
	if err != nil {
		return fmt.Errorf("request wechat API: %w", err)
	}
	defer response.Body.Close()
	if response.StatusCode < http.StatusOK || response.StatusCode >= http.StatusMultipleChoices {
		payload, _ := io.ReadAll(io.LimitReader(response.Body, 1024))
		glog.Warningf(ctx, "wechat API returned HTTP %d: %s", response.StatusCode, payload)
		return fmt.Errorf("wechat API returned HTTP %d", response.StatusCode)
	}
	if err := json.NewDecoder(response.Body).Decode(target); err != nil {
		return fmt.Errorf("decode wechat API response: %w", err)
	}
	return nil
}
