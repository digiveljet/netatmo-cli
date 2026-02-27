package api

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/digiveljet/netatmo-cli/auth"
	"github.com/digiveljet/netatmo-cli/config"
)

const stationsURL = "https://api.netatmo.com/api/getstationsdata"

type StationsResponse struct {
	Body struct {
		Devices []Device `json:"devices"`
	} `json:"body"`
	Status string `json:"status"`
	Error  *struct {
		Code    int    `json:"code"`
		Message string `json:"message"`
	} `json:"error"`
}

type Device struct {
	ID            string        `json:"_id"`
	StationName   string        `json:"station_name"`
	ModuleName    string        `json:"module_name"`
	Type          string        `json:"type"`
	DashboardData DashboardData `json:"dashboard_data"`
	DataType      []string      `json:"data_type"`
	Modules       []Module      `json:"modules"`
	Place         Place         `json:"place"`
}

type Module struct {
	ID            string        `json:"_id"`
	ModuleName    string        `json:"module_name"`
	Type          string        `json:"type"`
	DashboardData DashboardData `json:"dashboard_data"`
	DataType      []string      `json:"data_type"`
	BatteryPercent int          `json:"battery_percent"`
	RFStatus      int           `json:"rf_status"`
}

type DashboardData struct {
	Temperature      *float64 `json:"Temperature"`
	Humidity         *int     `json:"Humidity"`
	CO2              *int     `json:"CO2"`
	Noise            *int     `json:"Noise"`
	Pressure         *float64 `json:"Pressure"`
	AbsolutePressure *float64 `json:"AbsolutePressure"`
	Rain             *float64 `json:"Rain"`
	Rain1Hour        *float64 `json:"sum_rain_1"`
	Rain1Day         *float64 `json:"sum_rain_24"`
	WindAngle        *int     `json:"WindAngle"`
	WindStrength     *int     `json:"WindStrength"`
	GustAngle        *int     `json:"GustAngle"`
	GustStrength     *int     `json:"GustStrength"`
	TempMin          *float64 `json:"min_temp"`
	TempMax          *float64 `json:"max_temp"`
	TempTrend        string   `json:"temp_trend"`
	PressureTrend    string   `json:"pressure_trend"`
	TimeUTC          int64    `json:"time_utc"`
}

type Place struct {
	City     string    `json:"city"`
	Country  string    `json:"country"`
	Timezone string    `json:"timezone"`
	Location []float64 `json:"location"`
	Altitude int       `json:"altitude"`
}

func GetStations(cfg *config.Config) (*StationsResponse, error) {
	body, err := auth.AuthenticatedGet(cfg, stationsURL)
	if err != nil {
		return nil, err
	}

	var resp StationsResponse
	if err := json.Unmarshal(body, &resp); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}
	if resp.Error != nil {
		return nil, fmt.Errorf("API error %d: %s", resp.Error.Code, resp.Error.Message)
	}

	return &resp, nil
}

func FormatTime(ts int64) string {
	if ts == 0 {
		return "n/a"
	}
	return time.Unix(ts, 0).Format("15:04")
}

func TrendArrow(trend string) string {
	switch trend {
	case "up":
		return "↑"
	case "down":
		return "↓"
	case "stable":
		return "→"
	default:
		return ""
	}
}

func ModuleTypeName(t string) string {
	switch t {
	case "NAMain":
		return "Indoor"
	case "NAModule1":
		return "Outdoor"
	case "NAModule2":
		return "Wind"
	case "NAModule3":
		return "Rain"
	case "NAModule4":
		return "Indoor+"
	default:
		return t
	}
}

func BatteryIcon(percent int) string {
	switch {
	case percent > 75:
		return "████"
	case percent > 50:
		return "███░"
	case percent > 25:
		return "██░░"
	case percent > 10:
		return "█░░░"
	default:
		return "░░░░"
	}
}
