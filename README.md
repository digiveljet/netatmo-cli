# netatmo-cli

A fast, minimal CLI for Netatmo Weather Station. Single binary, zero dependencies.

Built with Go.

## Features

- 🌡 Temperature, humidity, CO₂, noise, pressure, rain, wind
- 🌳 Beautiful tree view with emoji indicators
- 🔋 Battery and RF status for modules
- 🔄 Automatic token refresh
- 📦 Single binary, zero dependencies (~9MB)

## Install

```bash
go install github.com/digiveljet/netatmo-cli@latest
```

Or build from source:

```bash
git clone https://github.com/digiveljet/netatmo-cli.git
cd netatmo-cli
go build -o netatmo .
```

## Setup

1. Create an app at [dev.netatmo.com/apps](https://dev.netatmo.com/apps)
2. Run:

```bash
netatmo auth <client_id> <client_secret>
```

3. Authorize in your browser
4. Done!

You can also generate tokens directly at dev.netatmo.com and save them manually to `~/.config/netatmo-cli/config.json`:

```json
{
  "client_id": "your_client_id",
  "client_secret": "your_client_secret",
  "access_token": "your_access_token",
  "refresh_token": "your_refresh_token"
}
```

## Usage

```bash
netatmo                    # tree view of all readings
netatmo temp               # temperatures only
netatmo json               # raw JSON (pipe to jq)
netatmo status             # station info + battery levels
netatmo current compact    # one-line summary
```

## Example output

### Tree view (`netatmo`)

```
📍 Home  (Joensuu @ 22:48)
│
├── Indoor [Indoor]
│   🌡  24.3°C →  (↓21.6 ↑24.3)
│   💧 29%
│   🟡 CO₂ 882 ppm
│   🔊 40 dB
│   🔵 1025.9 mb →
└── Outdoor [Outdoor]  🔋 61%
    🌡  1.3°C →  (↓-3.3 ↑2.1)
    💧 96%
```

### Temperatures only (`netatmo temp`)

```
📍 Home
  Indoor          24.3°C →  (↓21.6 ↑24.3)
  Outdoor          1.3°C →  (↓-3.3 ↑2.1)
```

### Station status (`netatmo status`)

```
📍 Home (Joensuu, FI)
  Base station: 70:ee:50:xx:xx:xx
  Modules: 1
    • Outdoor      Outdoor  🔋 61% ███░  RF: 65
```

### JSON (`netatmo json`)

Full JSON output for scripting. Pipe to `jq` for filtering:

```bash
netatmo json | jq '.[0].dashboard_data.Temperature'
```

## CO₂ indicators

| Level | Icon | Meaning |
|-------|------|---------|
| < 800 ppm | 🟢 | Good |
| 800–1000 ppm | 🟡 | Ventilate |
| > 1000 ppm | 🔴 | Poor air quality |

## Temperature trends

- `↑` rising
- `↓` falling
- `→` stable

## Supported modules

- **NAMain** — Base station (indoor)
- **NAModule1** — Outdoor module
- **NAModule2** — Wind gauge
- **NAModule3** — Rain gauge
- **NAModule4** — Additional indoor module

## Config

Stored at `~/.config/netatmo-cli/config.json`. Tokens auto-refresh — you shouldn't need to re-authorize unless you revoke access.

## Requirements

- A Netatmo Weather Station
- A Netatmo developer app ([create one here](https://dev.netatmo.com/apps))
- Go 1.21+ (to build)

## License

MIT
