package wecom

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"sync"
	"time"
)

type ClientConfig struct {
	APIBaseURL string
	CorpID     string
	AgentID    int
	Secret     string
}

type Client struct {
	cfg        ClientConfig
	httpClient *http.Client

	mu             sync.Mutex
	accessToken    string
	accessTokenExp time.Time
}

func NewClient(cfg ClientConfig, httpClient *http.Client) *Client {
	return &Client{
		cfg:        cfg,
		httpClient: httpClient,
	}
}

func (c *Client) SendText(ctx context.Context, msg TextMessage) error {
	payload := map[string]interface{}{
		"touser":  msg.ToUser,
		"msgtype": "text",
		"agentid": c.cfg.AgentID,
		"text": map[string]interface{}{
			"content": msg.Content,
		},
	}
	return c.sendMessage(ctx, payload)
}

func (c *Client) SendTemplateCard(ctx context.Context, msg TemplateCardMessage) error {
	if _, ok := msg.Card["task_id"]; !ok {
		msg.Card["task_id"] = "daily-help-" + strconv.FormatInt(time.Now().UnixNano(), 10)
	}
	payload := map[string]interface{}{
		"touser":        msg.ToUser,
		"msgtype":       "template_card",
		"agentid":       c.cfg.AgentID,
		"template_card": msg.Card,
	}
	return c.sendMessage(ctx, payload)
}

func (c *Client) sendMessage(ctx context.Context, payload map[string]interface{}) error {
	token, err := c.getAccessToken(ctx)
	if err != nil {
		return err
	}

	u := c.cfg.APIBaseURL + "/message/send?access_token=" + url.QueryEscape(token)
	body, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, u, bytes.NewReader(body))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")

	res, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	var out struct {
		ErrCode int    `json:"errcode"`
		ErrMsg  string `json:"errmsg"`
	}
	if err := json.NewDecoder(res.Body).Decode(&out); err != nil {
		return err
	}
	if out.ErrCode != 0 {
		return fmt.Errorf("wecom api error: %d %s", out.ErrCode, out.ErrMsg)
	}
	return nil
}

func (c *Client) getAccessToken(ctx context.Context) (string, error) {
	c.mu.Lock()
	if c.accessToken != "" && time.Now().Before(c.accessTokenExp.Add(-2*time.Minute)) {
		token := c.accessToken
		c.mu.Unlock()
		return token, nil
	}
	c.mu.Unlock()

	u := c.cfg.APIBaseURL + "/gettoken?corpid=" + url.QueryEscape(c.cfg.CorpID) + "&corpsecret=" + url.QueryEscape(c.cfg.Secret)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u, nil)
	if err != nil {
		return "", err
	}

	res, err := c.httpClient.Do(req)
	if err != nil {
		return "", err
	}
	defer res.Body.Close()

	var out struct {
		ErrCode     int    `json:"errcode"`
		ErrMsg      string `json:"errmsg"`
		AccessToken string `json:"access_token"`
		ExpiresIn   int    `json:"expires_in"`
	}
	if err := json.NewDecoder(res.Body).Decode(&out); err != nil {
		return "", err
	}
	if out.ErrCode != 0 {
		return "", fmt.Errorf("wecom gettoken error: %d %s", out.ErrCode, out.ErrMsg)
	}
	if out.AccessToken == "" || out.ExpiresIn == 0 {
		return "", errors.New("wecom gettoken 返回为空")
	}

	c.mu.Lock()
	c.accessToken = out.AccessToken
	c.accessTokenExp = time.Now().Add(time.Duration(out.ExpiresIn) * time.Second)
	c.mu.Unlock()

	return out.AccessToken, nil
}
