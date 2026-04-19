package delivery

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

type httpDoer interface {
	Do(req *http.Request) (*http.Response, error)
}

// GoogleMapsProvider uses Google Directions API for route estimates.
type GoogleMapsProvider struct {
	apiKey string
	client httpDoer
}

// NewGoogleMapsProvider creates a Google Maps route provider.
func NewGoogleMapsProvider(apiKey string) *GoogleMapsProvider {
	return &GoogleMapsProvider{
		apiKey: apiKey,
		client: &http.Client{Timeout: 5 * time.Second},
	}
}

// EstimateRoute queries Google Directions API for ETA and distance.
func (p *GoogleMapsProvider) EstimateRoute(ctx context.Context, from, to Location) (*RouteEstimate, error) {
	values := url.Values{}
	values.Set("origin", fmt.Sprintf("%f,%f", from.Latitude, from.Longitude))
	values.Set("destination", fmt.Sprintf("%f,%f", to.Latitude, to.Longitude))
	values.Set("key", p.apiKey)

	u := "https://maps.googleapis.com/maps/api/directions/json?" + values.Encode()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u, nil)
	if err != nil {
		return nil, err
	}

	resp, err := p.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 300 {
		return nil, fmt.Errorf("google maps returned status %d", resp.StatusCode)
	}

	var payload struct {
		Status string `json:"status"`
		Routes []struct {
			Legs []struct {
				Distance struct {
					Value float64 `json:"value"`
				} `json:"distance"`
				Duration struct {
					Value float64 `json:"value"`
				} `json:"duration"`
			} `json:"legs"`
		} `json:"routes"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		return nil, err
	}

	if payload.Status != "OK" || len(payload.Routes) == 0 || len(payload.Routes[0].Legs) == 0 {
		return nil, fmt.Errorf("google maps returned no route: %s", payload.Status)
	}

	leg := payload.Routes[0].Legs[0]
	distanceKM := leg.Distance.Value / 1000
	etaMinutes := int((leg.Duration.Value + 59) / 60)
	if etaMinutes < 1 {
		etaMinutes = 1
	}

	return &RouteEstimate{DistanceKM: distanceKM, ETAMinutes: etaMinutes}, nil
}

// MapboxProvider uses Mapbox Directions API for route estimates.
type MapboxProvider struct {
	apiKey string
	client httpDoer
}

// NewMapboxProvider creates a Mapbox route provider.
func NewMapboxProvider(apiKey string) *MapboxProvider {
	return &MapboxProvider{
		apiKey: apiKey,
		client: &http.Client{Timeout: 5 * time.Second},
	}
}

// EstimateRoute queries Mapbox Directions API for ETA and distance.
func (p *MapboxProvider) EstimateRoute(ctx context.Context, from, to Location) (*RouteEstimate, error) {
	coordFrom := strings.Join([]string{strconv.FormatFloat(from.Longitude, 'f', 6, 64), strconv.FormatFloat(from.Latitude, 'f', 6, 64)}, ",")
	coordTo := strings.Join([]string{strconv.FormatFloat(to.Longitude, 'f', 6, 64), strconv.FormatFloat(to.Latitude, 'f', 6, 64)}, ",")
	u := fmt.Sprintf("https://api.mapbox.com/directions/v5/mapbox/driving/%s;%s?overview=false&access_token=%s", coordFrom, coordTo, url.QueryEscape(p.apiKey))

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u, nil)
	if err != nil {
		return nil, err
	}

	resp, err := p.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 300 {
		return nil, fmt.Errorf("mapbox returned status %d", resp.StatusCode)
	}

	var payload struct {
		Routes []struct {
			Distance float64 `json:"distance"`
			Duration float64 `json:"duration"`
		} `json:"routes"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		return nil, err
	}

	if len(payload.Routes) == 0 {
		return nil, fmt.Errorf("mapbox returned no route")
	}

	route := payload.Routes[0]
	distanceKM := route.Distance / 1000
	etaMinutes := int((route.Duration + 59) / 60)
	if etaMinutes < 1 {
		etaMinutes = 1
	}

	return &RouteEstimate{DistanceKM: distanceKM, ETAMinutes: etaMinutes}, nil
}
