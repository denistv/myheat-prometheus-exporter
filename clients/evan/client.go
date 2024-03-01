package evan

import (
	logger "github.com/denistv/wdlogger"
	"net/http"
)

const endpointURL = "https://my2.myheat.net/api/request"

type Action string

const actionGetDevices Action = "getDevices"

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
	Action Action
	Login  string
	Key    string
}

type GetDevicesResponse struct {
}

func (c *Client) GetDevices(req GetDevicesRequest) (GetDevicesResponse, error) {
	return GetDevicesResponse{}, nil
}

type GetDevicesInfoRequest struct {
}

type GetDevicesInfoResponse struct {
	Data map[string]struct {
		Devices []Device `json:"devices"`
	} `json:"data"`
	Err         int64
	RefreshPage bool
}

type Device struct {
	ID           int64
	Name         string
	City         string
	Severity     int64
	SeverityDesc string
}

func (c *Client) GetDevicesInfo(req GetDevicesInfoRequest) (GetDevicesInfoResponse, error) {
	return GetDevicesInfoResponse{}, nil
}
