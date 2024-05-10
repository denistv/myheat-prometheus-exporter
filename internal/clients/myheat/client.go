package myheat

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"

	"github.com/denistv/wdlogger"
)

const endpointURL = "https://my.myheat.net/api/request/"

type action string

const (
	actionGetDevices    action = "getDevices"
	actionGetDeviceInfo action = "getDeviceInfo"
)

const (
	// К сожалению, поставщик API в своей документации не сообщает все возможные значения в Response,
	// поэтому здесь перечислены только те, которые мне известны.
	devSeverityNormal     = 1
	devSeverityLowBalance = 32
)

const successResponse = 0

const EnvTypeRoomTemperature = "room_temperature"

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

func NewClient(cfg Config, l wdlogger.Logger) *Client {
	return &Client{
		cfg:        cfg,
		logger:     l,
		httpClient: http.DefaultClient,
	}

}

type Client struct {
	cfg        Config
	logger     wdlogger.Logger
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
	Action action `json:"action"`
	Login  string `json:"login"`
	Key    string `json:"key"`
}

type GetDevicesResponse struct {
	Data        map[string][]Device `json:"data"`
	Err         int64               `json:"err"`
	RefreshPage bool                `json:"refreshPage"`
}

type Device struct {
	ID           int64  `json:"id"`
	Name         string `json:"name"`
	City         string `json:"city"`
	Severity     int64  `json:"severity"`
	SeverityDesc string `json:"severityDesc"`
}

func (c *Client) GetDevices(ctx context.Context) (GetDevicesResponse, error) {
	c.logger.Info("GetDevices - send request")
	defer func() {
		c.logger.Info("GetDevices - request completed")
	}()

	req := NewGetDevicesRequest(c.cfg.Login, c.cfg.Key)

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

	if res.Err != successResponse {
		c.logger.Error("server returned error", wdlogger.NewInt64Field("err", res.Err))
		return GetDevicesResponse{}, fmt.Errorf("server returned error")
	}

	return res, nil
}

func NewGetDeviceInfoRequest(login, key string, deviceID int64) GetDeviceInfoRequest {
	return GetDeviceInfoRequest{
		Action:   actionGetDeviceInfo,
		Login:    login,
		Key:      key,
		DeviceID: deviceID,
	}
}

type GetDeviceInfoRequest struct {
	Action   action `json:"action"`
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
			Demand       bool    `json:"demand"`
			ID           int64   `json:"id"`
			Name         string  `json:"name"`
			Severity     int64   `json:"severity"`
			SeverityDesc string  `json:"severityDesc"`
			Target       float64 `json:"target"`
			Type         string  `json:"type"`
			Value        float64 `json:"value"`
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
			// TODO write test for unmarshal value: "38.56874939532035"
			//TargetTemp    int64       `json:"targetTemp"`
		} `json:"heaters"`
		Severity     int64   `json:"severity"`
		SeverityDesc string  `json:"severityDesc"`
		WeatherTemp  float64 `json:"weatherTemp,string"` // Example: "1.4600000000000364"
	} `json:"data"`
	Err         int64 `json:"err"`
	RefreshPage bool  `json:"refreshPage"`
}

func (c *Client) GetDeviceInfo(ctx context.Context, id int64) (GetDeviceInfoResponse, error) {
	req := NewGetDeviceInfoRequest(c.cfg.Login, c.cfg.Key, id)

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

	if res.Err != successResponse {
		c.logger.Error("server returned error", wdlogger.NewInt64Field("err", res.Err))
		return GetDeviceInfoResponse{}, fmt.Errorf("server returned error")
	}

	return res, nil
}
