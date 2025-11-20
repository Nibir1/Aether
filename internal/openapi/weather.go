// internal/openapi/weather.go
//
// Weather integration using MET Norway’s free, key-less API.
// Docs: https://api.met.no/
// This API requires NO key, only a User-Agent string.
//

package openapi

import (
	"context"
	"encoding/json"
	"fmt"
	"time"
)

// Weather represents a normalized hourly weather entry.
type Weather struct {
	Time        time.Time
	Temperature float64
	Humidity    float64
	WindSpeed   float64
	Summary     string
}

// metNorwayResponse matches only fields we care about.
type metNorwayResponse struct {
	Properties struct {
		Timeseries []struct {
			Time string `json:"time"`
			Data struct {
				Instant struct {
					Details struct {
						Temperature float64 `json:"air_temperature"`
						Humidity    float64 `json:"relative_humidity"`
						WindSpeed   float64 `json:"wind_speed"`
					} `json:"details"`
				} `json:"instant"`
			} `json:"data"`
		} `json:"timeseries"`
	} `json:"properties"`
}

// WeatherAt retrieves normalized weather data for latitude/longitude.
func (c *Client) WeatherAt(ctx context.Context, lat, lon float64, hours int) ([]Weather, error) {
	if hours <= 0 {
		hours = 12
	}
	if hours > 24 {
		hours = 24
	}

	url := fmt.Sprintf(
		"https://api.met.no/weatherapi/locationforecast/2.0/compact?lat=%f&lon=%f",
		lat, lon,
	)

	body, _, err := c.getJSON(ctx, url)
	if err != nil {
		return nil, err
	}

	var resp metNorwayResponse
	if err := json.Unmarshal(body, &resp); err != nil {
		return nil, err
	}

	out := []Weather{}
	for i, t := range resp.Properties.Timeseries {
		if i >= hours {
			break
		}
		tm, _ := time.Parse(time.RFC3339, t.Time)
		w := Weather{
			Time:        tm,
			Temperature: t.Data.Instant.Details.Temperature,
			Humidity:    t.Data.Instant.Details.Humidity,
			WindSpeed:   t.Data.Instant.Details.WindSpeed,
			Summary: fmt.Sprintf("%.1f°C, %.0f%% humidity, %.1f m/s wind",
				t.Data.Instant.Details.Temperature,
				t.Data.Instant.Details.Humidity,
				t.Data.Instant.Details.WindSpeed,
			),
		}
		out = append(out, w)
	}

	return out, nil
}
