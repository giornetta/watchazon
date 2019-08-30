package locator

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/giornetta/watchazon"
)

type apiResponse struct {
	Response struct {
		View []struct {
			Result []struct {
				Location struct {
					Address struct {
						Label      string `json:"Label"`
						Country    string `json:"Country"`
						State      string `json:"State"`
						County     string `json:"County"`
						City       string `json:"City"`
						District   string `json:"District"`
						Street     string `json:"Street"`
						PostalCode string `json:"PostalCode"`
					} `json:"Address"`
				} `json:"Location"`
			} `json:"Result"`
		} `json:"View"`
	} `json:"Response"`
}

type service struct {
	appID   string
	appCode string
}

func New(appId, appCode string) watchazon.Locator {
	return &service{
		appID:   appId,
		appCode: appCode,
	}
}

func (s *service) Locate(lat, long float32) (watchazon.Domain, error) {
	c := http.Client{}

	url := fmt.Sprintf("https://reverse.geocoder.api.here.com/6.2/reversegeocode.json?app_id=%s&app_code=%s&mode=trackPosition&pos=%f,%f,0&maxresults=1", s.appID, s.appCode, lat, long)

	res, err := c.Get(url)
	if err != nil || res.StatusCode != 200 {
		return "", fmt.Errorf("could not get location: %v", err)
	}

	var b apiResponse
	if err := json.NewDecoder(res.Body).Decode(&b); err != nil {
		return "", err
	}

	var domain watchazon.Domain
	switch b.Response.View[0].Result[0].Location.Address.Country {
	case "ITA":
		domain = "it"
	case "DEU":
		domain = "de"
	case "ESP":
		domain = "es"
	default:
		domain = "com"
	}

	return domain, nil
}
