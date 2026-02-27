# netatmo-cli

A minimal CLI for Netatmo Weather Station. Single binary, zero dependencies.

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
2. Authorize:

```bash
netatmo auth <client_id> <client_secret>
```

3. Open the URL in your browser and authorize
4. Done!

## Usage

```bash
netatmo              # tree view of all readings
netatmo temp         # temperatures only
netatmo json         # raw JSON (pipe to jq)
netatmo status       # station info + battery levels
netatmo current compact  # one-line summary
```

## Example output

```
📍 Home  (Joensuu @ 22:30)
│
├── Indoor [Indoor]
│   🌡  21.5°C →  (↓19.2 ↑22.1)
│   💧 52%
│   🟢 CO₂ 650 ppm
│   🔊 38 dB
│   🔵 1013.2 mb →
└── Outdoor [Outdoor]  🔋 78%
    🌡  -5.2°C ↓  (↓-8.1 ↑-2.3)
    💧 89%
```

## Config

Stored at `~/.config/netatmo-cli/config.json`. Tokens auto-refresh.

## License

MIT
