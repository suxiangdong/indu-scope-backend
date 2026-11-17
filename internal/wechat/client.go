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
	accessTokenLockTTL         = 15 * time.Second
	invalidAccessTokenErrCode  = 40001
	accessTokenRedisPrefix     = "indu_scope:wechat:access_token:"
	accessTokenRedisLockPrefix = "indu_scope:wechat:access_token_lock:"
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

// accessTokenCache is the process-local mirror of the app-level token. The
// Redis copy is shared across instances so stable_token is not refreshed by
// every replica independently.
var accessTokenCache struct {
	sync.Mutex
	appID       string
	apiBaseURL  string
	token       string
	expiresAt   time.Time
	refreshedAt time.Time
}

var wechatHTTPClient = &http.Client{Timeout: 8 * time.Second}

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
	payload, err := json.Marshal(struct {
		Code string `json:"code"`
	}{Code: strings.TrimSpace(code)})
	if err != nil {
		return "", fmt.Errorf("encode wechat phone request: %w", err)
	}
	for attempt := 0; attempt < 2; attempt++ {
		accessToken, err := getAccessToken(ctx, config)
		if err != nil {
			return "", err
		}
		response, err := requestPhoneNumber(ctx, config, accessToken, payload)
		if err != nil {
			return "", err
		}
		if response.ErrCode == invalidAccessTokenErrCode && attempt == 0 {
			// Another process may have refreshed the app token. Clear only the
			// token used by this request, then obtain a stable replacement once.
			glog.Warningf(ctx, "wechat phone API rejected the cached access token; refreshing stable access token")
			invalidateAccessToken(ctx, config, accessToken)
			continue
		}
		if response.ErrCode != 0 || response.PhoneInfo.PurePhone == "" {
			return "", fmt.Errorf("wechat phone API failed: %d %s", response.ErrCode, response.ErrMsg)
		}
		return response.PhoneInfo.PurePhone, nil
	}
	return "", fmt.Errorf("wechat phone API failed: %d invalid access token", invalidAccessTokenErrCode)
}

func requestPhoneNumber(ctx context.Context, config clientConfig, accessToken string, payload []byte) (phoneNumberResponse, error) {
	endpoint := config.APIBaseURL + "/wxa/business/getuserphonenumber?access_token=" + url.QueryEscape(accessToken)
	request, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewReader(payload))
	if err != nil {
		return phoneNumberResponse{}, err
	}
	request.Header.Set("Content-Type", "application/json")
	responseBody, err := wechatHTTPClient.Do(request)
	if err != nil {
		return phoneNumberResponse{}, fmt.Errorf("request wechat phone API: %w", err)
	}
	defer responseBody.Body.Close()
	if responseBody.StatusCode < http.StatusOK || responseBody.StatusCode >= http.StatusMultipleChoices {
		payload, _ := io.ReadAll(io.LimitReader(responseBody.Body, 1024))
		return phoneNumberResponse{}, fmt.Errorf("wechat phone API returned HTTP %d: %s", responseBody.StatusCode, strings.TrimSpace(string(payload)))
	}
	var response phoneNumberResponse
	if err := json.NewDecoder(io.LimitReader(responseBody.Body, 1<<20)).Decode(&response); err != nil {
		return phoneNumberResponse{}, fmt.Errorf("decode wechat phone API response: %w", err)
	}
	return response, nil
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
	if !strings.HasPrefix(strings.ToLower(baseURL), "http://") && !strings.HasPrefix(strings.ToLower(baseURL), "https://") {
		baseURL = "https://" + baseURL
	}
	if strings.HasPrefix(strings.ToLower(baseURL), "http://") {
		return clientConfig{}, fmt.Errorf("wechat.api_base_url must use HTTPS")
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
	if token, expiresAt, ok := loadSharedAccessToken(ctx, config); ok {
		setLocalAccessToken(config, token, expiresAt)
		return token, nil
	}

	lockKey := accessTokenRedisLockPrefix + config.AppID
	lockValue := fmt.Sprintf("%d", time.Now().UnixNano())
	for attempt := 0; attempt < 20; attempt++ {
		if token, expiresAt, ok := loadSharedAccessToken(ctx, config); ok {
			setLocalAccessToken(config, token, expiresAt)
			return token, nil
		}
		acquired, err := acquireAccessTokenLock(ctx, lockKey, lockValue)
		if err != nil {
			// Redis is already checked during startup, but keep phone login
			// available if the shared cache has a transient failure.
			return fetchAccessToken(ctx, config)
		}
		if acquired {
			defer releaseAccessTokenLock(ctx, lockKey, lockValue)
			if token, expiresAt, ok := loadSharedAccessToken(ctx, config); ok {
				setLocalAccessToken(config, token, expiresAt)
				return token, nil
			}
			token, err := fetchAccessToken(ctx, config)
			if err != nil {
				return "", err
			}
			_ = storeSharedAccessToken(ctx, config, token, accessTokenCache.expiresAt)
			return token, nil
		}
		timer := time.NewTimer(100 * time.Millisecond)
		select {
		case <-ctx.Done():
			timer.Stop()
			return "", ctx.Err()
		case <-timer.C:
		}
	}
	return fetchAccessToken(ctx, config)
}

func fetchAccessToken(ctx context.Context, config clientConfig) (string, error) {
	payload := struct {
		GrantType    string `json:"grant_type"`
		AppID        string `json:"appid"`
		Secret       string `json:"secret"`
		ForceRefresh bool   `json:"force_refresh"`
	}{
		GrantType:    "client_credential",
		AppID:        config.AppID,
		Secret:       config.AppSecret,
		ForceRefresh: false,
	}
	var response accessTokenResponse
	if err := postJSON(ctx, config.APIBaseURL+"/cgi-bin/stable_token", payload, &response); err != nil {
		return "", err
	}
	if response.ErrCode != 0 || response.AccessToken == "" {
		return "", fmt.Errorf("wechat access token failed: %d %s", response.ErrCode, response.ErrMsg)
	}
	setLocalAccessToken(config, response.AccessToken, time.Now().Add(time.Duration(response.ExpiresIn)*time.Second))
	if response.ExpiresIn > 0 {
		accessTokenCache.expiresAt = time.Now().Add(time.Duration(response.ExpiresIn) * time.Second)
	} else {
		// Treat a malformed lifetime as immediately expiring rather than
		// retaining an unusable token indefinitely.
		accessTokenCache.expiresAt = time.Now()
	}
	return response.AccessToken, nil
}

func setLocalAccessToken(config clientConfig, token string, expiresAt time.Time) {
	accessTokenCache.appID = config.AppID
	accessTokenCache.apiBaseURL = config.APIBaseURL
	accessTokenCache.token = token
	accessTokenCache.refreshedAt = time.Now()
	accessTokenCache.expiresAt = expiresAt
}

func invalidateAccessToken(ctx context.Context, config clientConfig, token string) {
	accessTokenCache.Lock()
	defer accessTokenCache.Unlock()
	if accessTokenCache.appID != config.AppID || accessTokenCache.apiBaseURL != config.APIBaseURL || accessTokenCache.token != token {
		return
	}
	accessTokenCache.token = ""
	accessTokenCache.expiresAt = time.Time{}
	accessTokenCache.refreshedAt = time.Time{}
	if value, err := g.Redis().Do(ctx, "GET", accessTokenRedisPrefix+config.AppID); err == nil && value != nil && value.String() == token {
		_, _ = g.Redis().Do(ctx, "DEL", accessTokenRedisPrefix+config.AppID)
	}
}

func loadSharedAccessToken(ctx context.Context, config clientConfig) (string, time.Time, bool) {
	value, err := g.Redis().Do(ctx, "GET", accessTokenRedisPrefix+config.AppID)
	if err != nil || value == nil || value.IsNil() {
		return "", time.Time{}, false
	}
	token := strings.TrimSpace(value.String())
	if token == "" {
		return "", time.Time{}, false
	}
	ttl, err := g.Redis().Do(ctx, "TTL", accessTokenRedisPrefix+config.AppID)
	if err != nil || ttl == nil || ttl.Int64() <= 0 {
		return "", time.Time{}, false
	}
	return token, time.Now().Add(time.Duration(ttl.Int64()) * time.Second), true
}

func storeSharedAccessToken(ctx context.Context, config clientConfig, token string, expiresAt time.Time) error {
	ttl := int64(time.Until(expiresAt).Seconds()) - 60
	if ttl < 1 {
		ttl = 1
	}
	_, err := g.Redis().Do(ctx, "SET", accessTokenRedisPrefix+config.AppID, token, "EX", ttl)
	return err
}

func acquireAccessTokenLock(ctx context.Context, key, value string) (bool, error) {
	result, err := g.Redis().Do(ctx, "SET", key, value, "NX", "EX", int(accessTokenLockTTL.Seconds()))
	if err != nil {
		return false, err
	}
	return result != nil && strings.EqualFold(result.String(), "OK"), nil
}

func releaseAccessTokenLock(ctx context.Context, key, value string) {
	const releaseScript = `if redis.call('get', KEYS[1]) == ARGV[1] then return redis.call('del', KEYS[1]) else return 0 end`
	_, _ = g.Redis().Do(ctx, "EVAL", releaseScript, 1, key, value)
}

func getJSON(ctx context.Context, endpoint string, target any) error {
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return err
	}
	response, err := wechatHTTPClient.Do(request)
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

func postJSON(ctx context.Context, endpoint string, payload any, target any) error {
	body, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("encode wechat API request: %w", err)
	}
	request, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewReader(body))
	if err != nil {
		return err
	}
	request.Header.Set("Content-Type", "application/json")
	response, err := wechatHTTPClient.Do(request)
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
