package iplocate

import (
	"encoding/json"
	"fmt"
	"strings"
)

type HTTPGetter interface {
	Get(url string) ([]byte, error)
}

type IPActivties struct {
	HTTPClient HTTPGetter
}

// GetIP Fetches the public IP Address
func (a *IPActivties) GetIP() (string, error) {
	resp, err := a.HTTPClient.Get("https://icanhazip.com")
	if err != nil {
		return "", err
	}

	ip := strings.TrimSpace(string(resp))
	return ip, nil
}

func (a *IPActivties) GetLocationInfo(ip string) (string, error) {
	resp, err := a.HTTPClient.Get("https://ipapi.co/" + ip + "/json/")
	if err != nil {
		return "", err
	}

	var data struct {
		City    string `json:"city"`
		Country string `json:"country"`
		Region  string `json:"region"`
	}

	err = json.Unmarshal(resp, &data)
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("%s, %s, %s", data.City, data.Region, data.Country), nil
}
