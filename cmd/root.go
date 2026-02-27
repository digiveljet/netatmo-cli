package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/digiveljet/netatmo-cli/api"
	"github.com/digiveljet/netatmo-cli/auth"
	"github.com/digiveljet/netatmo-cli/config"
)

func Execute() error {
	if len(os.Args) < 2 {
		return runCurrent("tree")
	}

	switch os.Args[1] {
	case "auth":
		return runAuth()
	case "current", "now":
		format := "tree"
		if len(os.Args) > 2 {
			format = os.Args[2]
		}
		return runCurrent(format)
	case "temp":
		return runTemp()
	case "json":
		return runJSON()
	case "status":
		return runStatus()
	case "help", "--help", "-h":
		printHelp()
		return nil
	case "version", "--version":
		fmt.Println("netatmo-cli v0.1.0")
		return nil
	default:
		return fmt.Errorf("unknown command: %s — run 'netatmo help' for usage", os.Args[1])
	}
}

func runAuth() error {
	if len(os.Args) < 4 {
		fmt.Println("Usage: netatmo auth <client_id> <client_secret>")
		fmt.Println()
		fmt.Println("Get your credentials at: https://dev.netatmo.com/apps")
		return nil
	}

	clientID := os.Args[2]
	clientSecret := os.Args[3]

	cfg, err := auth.Login(clientID, clientSecret)
	if err != nil {
		return err
	}

	if err := cfg.Save(); err != nil {
		return fmt.Errorf("failed to save config: %w", err)
	}

	fmt.Println("✅ Authorized and saved!")
	fmt.Printf("Config: %s\n", config.Path())
	return nil
}

func runCurrent(format string) error {
	cfg, err := auth.EnsureToken()
	if err != nil {
		return err
	}

	stations, err := api.GetStations(cfg)
	if err != nil {
		return err
	}

	for _, device := range stations.Body.Devices {
		switch format {
		case "tree":
			printTree(device)
		case "compact":
			printCompact(device)
		default:
			printTree(device)
		}
	}
	return nil
}

func runTemp() error {
	cfg, err := auth.EnsureToken()
	if err != nil {
		return err
	}

	stations, err := api.GetStations(cfg)
	if err != nil {
		return err
	}

	for _, device := range stations.Body.Devices {
		fmt.Printf("📍 %s\n", device.StationName)
		d := device.DashboardData
		if d.Temperature != nil {
			trend := api.TrendArrow(d.TempTrend)
			fmt.Printf("  %-12s  %5.1f°C %s", device.ModuleName, *d.Temperature, trend)
			if d.TempMin != nil && d.TempMax != nil {
				fmt.Printf("  (↓%.1f ↑%.1f)", *d.TempMin, *d.TempMax)
			}
			fmt.Println()
		}
		for _, m := range device.Modules {
			if m.DashboardData.Temperature != nil {
				trend := api.TrendArrow(m.DashboardData.TempTrend)
				fmt.Printf("  %-12s  %5.1f°C %s", m.ModuleName, *m.DashboardData.Temperature, trend)
				if m.DashboardData.TempMin != nil && m.DashboardData.TempMax != nil {
					fmt.Printf("  (↓%.1f ↑%.1f)", *m.DashboardData.TempMin, *m.DashboardData.TempMax)
				}
				fmt.Println()
			}
		}
	}
	return nil
}

func runJSON() error {
	cfg, err := auth.EnsureToken()
	if err != nil {
		return err
	}

	stations, err := api.GetStations(cfg)
	if err != nil {
		return err
	}

	data, err := json.MarshalIndent(stations.Body.Devices, "", "  ")
	if err != nil {
		return err
	}
	fmt.Println(string(data))
	return nil
}

func runStatus() error {
	cfg, err := auth.EnsureToken()
	if err != nil {
		return err
	}

	stations, err := api.GetStations(cfg)
	if err != nil {
		return err
	}

	for _, device := range stations.Body.Devices {
		fmt.Printf("📍 %s (%s, %s)\n", device.StationName, device.Place.City, device.Place.Country)
		fmt.Printf("  Base station: %s\n", device.ID)
		fmt.Printf("  Modules: %d\n", len(device.Modules))
		for _, m := range device.Modules {
			fmt.Printf("    • %-12s  %s  🔋 %d%% %s  RF: %d\n",
				m.ModuleName,
				api.ModuleTypeName(m.Type),
				m.BatteryPercent,
				api.BatteryIcon(m.BatteryPercent),
				m.RFStatus,
			)
		}
	}
	return nil
}

func printTree(device api.Device) {
	d := device.DashboardData
	fmt.Printf("📍 %s  (%s @ %s)\n", device.StationName, device.Place.City, api.FormatTime(d.TimeUTC))
	fmt.Printf("│\n")

	// Main station data
	fmt.Printf("├── %s [%s]\n", device.ModuleName, api.ModuleTypeName(device.Type))
	printDashboard("│   ", d)

	// Modules
	for i, m := range device.Modules {
		prefix := "├"
		childPrefix := "│   "
		if i == len(device.Modules)-1 {
			prefix = "└"
			childPrefix = "    "
		}
		fmt.Printf("%s── %s [%s]", prefix, m.ModuleName, api.ModuleTypeName(m.Type))
		if m.BatteryPercent > 0 {
			fmt.Printf("  🔋 %d%%", m.BatteryPercent)
		}
		fmt.Println()
		printDashboard(childPrefix, m.DashboardData)
	}
	fmt.Println()
}

func printDashboard(prefix string, d api.DashboardData) {
	var lines []string

	if d.Temperature != nil {
		trend := api.TrendArrow(d.TempTrend)
		s := fmt.Sprintf("🌡  %.1f°C %s", *d.Temperature, trend)
		if d.TempMin != nil && d.TempMax != nil {
			s += fmt.Sprintf("  (↓%.1f ↑%.1f)", *d.TempMin, *d.TempMax)
		}
		lines = append(lines, s)
	}
	if d.Humidity != nil {
		lines = append(lines, fmt.Sprintf("💧 %d%%", *d.Humidity))
	}
	if d.CO2 != nil {
		icon := "🟢"
		if *d.CO2 > 1000 {
			icon = "🔴"
		} else if *d.CO2 > 800 {
			icon = "🟡"
		}
		lines = append(lines, fmt.Sprintf("%s CO₂ %d ppm", icon, *d.CO2))
	}
	if d.Noise != nil {
		lines = append(lines, fmt.Sprintf("🔊 %d dB", *d.Noise))
	}
	if d.Pressure != nil {
		trend := api.TrendArrow(d.PressureTrend)
		lines = append(lines, fmt.Sprintf("🔵 %.1f mb %s", *d.Pressure, trend))
	}
	if d.Rain != nil {
		s := fmt.Sprintf("🌧  %.1f mm", *d.Rain)
		if d.Rain1Hour != nil {
			s += fmt.Sprintf("  (1h: %.1f)", *d.Rain1Hour)
		}
		if d.Rain1Day != nil {
			s += fmt.Sprintf("  (24h: %.1f)", *d.Rain1Day)
		}
		lines = append(lines, s)
	}
	if d.WindStrength != nil {
		s := fmt.Sprintf("💨 %d km/h", *d.WindStrength)
		if d.WindAngle != nil {
			s += fmt.Sprintf(" %s", windDirection(*d.WindAngle))
		}
		if d.GustStrength != nil {
			s += fmt.Sprintf("  (gust: %d km/h)", *d.GustStrength)
		}
		lines = append(lines, s)
	}

	for _, line := range lines {
		fmt.Printf("%s%s\n", prefix, line)
	}
}

func printCompact(device api.Device) {
	var parts []string
	d := device.DashboardData

	if d.Temperature != nil {
		parts = append(parts, fmt.Sprintf("%.1f°C", *d.Temperature))
	}
	if d.Humidity != nil {
		parts = append(parts, fmt.Sprintf("%d%%", *d.Humidity))
	}
	if d.CO2 != nil {
		parts = append(parts, fmt.Sprintf("CO₂:%d", *d.CO2))
	}

	fmt.Printf("%s: %s", device.ModuleName, strings.Join(parts, " | "))

	for _, m := range device.Modules {
		var mparts []string
		md := m.DashboardData
		if md.Temperature != nil {
			mparts = append(mparts, fmt.Sprintf("%.1f°C", *md.Temperature))
		}
		if md.Humidity != nil {
			mparts = append(mparts, fmt.Sprintf("%d%%", *md.Humidity))
		}
		if len(mparts) > 0 {
			fmt.Printf(" · %s: %s", m.ModuleName, strings.Join(mparts, " | "))
		}
	}
	fmt.Println()
}

func windDirection(angle int) string {
	dirs := []string{"N", "NNE", "NE", "ENE", "E", "ESE", "SE", "SSE", "S", "SSW", "SW", "WSW", "W", "WNW", "NW", "NNW"}
	idx := ((angle + 11) / 22) % 16
	return dirs[idx]
}

func printHelp() {
	fmt.Println(`netatmo-cli — Netatmo Weather Station CLI

Usage:
  netatmo                    Show current readings (tree view)
  netatmo current [format]   Show current readings (tree|compact)
  netatmo temp               Show temperatures only
  netatmo json               Raw JSON output
  netatmo status             Station & module info + battery
  netatmo auth <id> <secret> Authorize with Netatmo API
  netatmo version            Show version
  netatmo help               Show this help

Setup:
  1. Create an app at https://dev.netatmo.com/apps
  2. Run: netatmo auth <client_id> <client_secret>
  3. Authorize in your browser
  4. Done! Run 'netatmo' to see your data.

Config stored at: ~/.config/netatmo-cli/config.json`)
}
