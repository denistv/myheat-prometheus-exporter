package evan

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	logger "github.com/denistv/wdlogger"
	"io"
	"net/http"
)

const endpointURL = "https://my2.myheat.net/api/request/"

type Action string

const (
	actionGetDevices    Action = "getDevices"
	actionGetDeviceInfo Action = "getDeviceInfo"
)

func NewDefaultConfig() Config {
	return Config{
		EndpointURL: endpointURL,
	}
}

type Config struct {
	EndpointURL string
	Login       string
	Key         string
}

func (c *Config) Validate() error {
	if c.Key == "" {
		return errors.New("key cannot be empty")
	}

	if c.Login == "" {
		return errors.New("login cannot be empty")
	}

	return nil
}

func NewClient(cfg Config, l logger.Logger) *Client {
	return &Client{
		cfg:        cfg,
		logger:     l,
		httpClient: http.DefaultClient,
	}

}

type Client struct {
	cfg        Config
	logger     logger.Logger
	httpClient *http.Client
}

func NewGetDevicesRequest(login, key string) GetDevicesRequest {
	return GetDevicesRequest{
		Action: actionGetDevices,
		Login:  login,
		Key:    key,
	}
}

type GetDevicesRequest struct {
	Action Action `json:"action"`
	Login  string `json:"login"`
	Key    string `json:"key"`
}

type GetDevicesResponse struct {
	Data        map[string][]Device `json:"data"`
	Err         int64               `json:"err"`
	RefreshPage bool                `json:"refreshPage"`
}

func (c *Client) GetDevices(ctx context.Context) (GetDevicesResponse, error) {
	c.logger.Info("GetDevices - send request")
	defer func() {
		c.logger.Info("GetDevices - request completed")
	}()

	req := GetDevicesRequest{
		Action: actionGetDevices,
		Login:  c.cfg.Login,
		Key:    c.cfg.Key,
	}

	data, err := json.Marshal(req)
	if err != nil {
		return GetDevicesResponse{}, err
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, c.cfg.EndpointURL, bytes.NewBuffer(data))
	if err != nil {
		return GetDevicesResponse{}, err
	}

	httpReq.Header.Set("Content-Type", "application/json")

	resRaw, err := c.httpClient.Do(httpReq)
	if err != nil {
		return GetDevicesResponse{}, err
	}
	defer resRaw.Body.Close()

	data, err = io.ReadAll(resRaw.Body)
	if err != nil {
		return GetDevicesResponse{}, err
	}

	res := GetDevicesResponse{}
	err = json.Unmarshal(data, &res)
	if err != nil {
		return GetDevicesResponse{}, err
	}

	if res.Err != 0 {
		c.logger.Error("server returned error", logger.NewInt64Field("err", res.Err))
		return GetDevicesResponse{}, fmt.Errorf("server returned error")
	}

	return res, nil
}

type GetDeviceInfoRequest struct {
	Action   Action `json:"action"`
	DeviceID int64  `json:"deviceId"`
	Login    string `json:"login"`
	Key      string `json:"key"`
}

type GetDeviceInfoResponse struct {
	Data struct {
		Alarms     []interface{} `json:"alarms"`
		City       string        `json:"city"`
		DataActual bool          `json:"dataActual"`
		Engs       []interface{} `json:"engs"`
		Envs       []struct {
			Demand       bool   `json:"demand"`
			ID           int64  `json:"id"`
			Name         string `json:"name"`
			Severity     int64  `json:"severity"`
			SeverityDesc string `json:"severityDesc"`
			// TODO write test for unmarshal value: "38.56874939532035"
			Target float64 `json:"target"`
			Type   string  `json:"type"`
			Value  float64 `json:"value"`
		} `json:"envs"`
		Heaters []struct {
			BurnerHeating bool        `json:"burnerHeating"`
			BurnerWater   bool        `json:"burnerWater"`
			Disabled      bool        `json:"disabled"`
			FlowTemp      interface{} `json:"flowTemp"`
			ID            int64       `json:"id"`
			Modulation    interface{} `json:"modulation"`
			Name          string      `json:"name"`
			Pressure      interface{} `json:"pressure"`
			ReturnTemp    interface{} `json:"returnTemp"`
			//TargetTemp    int64       `json:"targetTemp"`
		} `json:"heaters"`
		Severity     int64  `json:"severity"`
		SeverityDesc string `json:"severityDesc"`
		WeatherTemp  string `json:"weatherTemp"`
	} `json:"data"`
	Err         int64 `json:"err"`
	RefreshPage bool  `json:"refreshPage"`
}

type Device struct {
	ID           int64  `json:"id"`
	Name         string `json:"name"`
	City         string `json:"city"`
	Severity     int64  `json:"severity"`
	SeverityDesc string `json:"severityDesc"`
}

func (c *Client) GetDeviceInfo(ctx context.Context, id int64) (GetDeviceInfoResponse, error) {
	req := GetDeviceInfoRequest{
		Action:   actionGetDeviceInfo,
		Login:    c.cfg.Login,
		Key:      c.cfg.Key,
		DeviceID: id,
	}
	data, err := json.Marshal(req)
	if err != nil {
		return GetDeviceInfoResponse{}, err
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, c.cfg.EndpointURL, bytes.NewBuffer(data))
	if err != nil {
		return GetDeviceInfoResponse{}, err
	}

	httpReq.Header.Set("Content-Type", "application/json")

	resRaw, err := c.httpClient.Do(httpReq)
	if err != nil {
		return GetDeviceInfoResponse{}, err
	}
	defer resRaw.Body.Close()

	data, err = io.ReadAll(resRaw.Body)
	if err != nil {
		return GetDeviceInfoResponse{}, err
	}

	res := GetDeviceInfoResponse{}
	err = json.Unmarshal(data, &res)
	if err != nil {
		return GetDeviceInfoResponse{}, err
	}

	if res.Err != 0 {
		c.logger.Error("server returned error", logger.NewInt64Field("err", res.Err))
		return GetDeviceInfoResponse{}, fmt.Errorf("server returned error")
	}

	return res, nil
}
