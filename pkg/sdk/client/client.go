package client

import (
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/fatedier/frp/client"
	"github.com/fatedier/frp/pkg/util/util"
)

type Client struct {
	address  string
	authUser string
	authPwd  string
}

func New(host string, port int) *Client {
	return &Client{
		address: net.JoinHostPort(host, strconv.Itoa(port)),
	}
}

func (c *Client) SetAuth(user, pwd string) {
	c.authUser = user
	c.authPwd = pwd
}

func (c *Client) GetProxyStatus(name string) (*client.ProxyStatusResp, error) {
	req, err := http.NewRequest("GET", "http://"+c.address+"/api/status", nil)
	if err != nil {
		return nil, err
	}
	content, err := c.do(req)
	if err != nil {
		return nil, err
	}
	allStatus := make(client.StatusResp)
	if err = json.Unmarshal([]byte(content), &allStatus); err != nil {
		return nil, fmt.Errorf("unmarshal http response error: %s", strings.TrimSpace(content))
	}
	for _, pss := range allStatus {
		for _, ps := range pss {
			if ps.Name == name {
				return &ps, nil
			}
		}
	}
	return nil, fmt.Errorf("no proxy status found")
}

func (c *Client) GetAllProxyStatus() (client.StatusResp, error) {
	req, err := http.NewRequest("GET", "http://"+c.address+"/api/status", nil)
	if err != nil {
		return nil, err
	}
	content, err := c.do(req)
	if err != nil {
		return nil, err
	}
	allStatus := make(client.StatusResp)
	if err = json.Unmarshal([]byte(content), &allStatus); err != nil {
		return nil, fmt.Errorf("unmarshal http response error: %s", strings.TrimSpace(content))
	}
	return allStatus, nil
}

func (c *Client) Reload(strictMode bool) error {
	v := url.Values{}
	if strictMode {
		v.Set("strictConfig", "true")
	}
	queryStr := ""
	if len(v) > 0 {
		queryStr = "?" + v.Encode()
	}
	req, err := http.NewRequest("GET", "http://"+c.address+"/api/reload"+queryStr, nil)
	if err != nil {
		return err
	}
	_, err = c.do(req)
	return err
}

func (c *Client) Stop() error {
	req, err := http.NewRequest("POST", "http://"+c.address+"/api/stop", nil)
	if err != nil {
		return err
	}
	_, err = c.do(req)
	return err
}

func (c *Client) GetConfig() (string, error) {
	req, err := http.NewRequest("GET", "http://"+c.address+"/api/config", nil)
	if err != nil {
		return "", err
	}
	return c.do(req)
}

func (c *Client) UpdateConfig(content string) error {
	req, err := http.NewRequest("PUT", "http://"+c.address+"/api/config", strings.NewReader(content))
	if err != nil {
		return err
	}
	_, err = c.do(req)
	return err
}

func (c *Client) setAuthHeader(req *http.Request) {
	if c.authUser != "" || c.authPwd != "" {
		req.Header.Set("Authorization", util.BasicAuth(c.authUser, c.authPwd))
	}
}

func (c *Client) do(req *http.Request) (string, error) {
	c.setAuthHeader(req)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return "", fmt.Errorf("api status code [%d]", resp.StatusCode)
	}
	buf, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	return string(buf), nil
}
