# Stockyard Waiver

**Self-hosted digital waiver and consent form signing**

Part of the [Stockyard](https://stockyard.dev) family of self-hosted tools.

## Quick Start

```bash
curl -fsSL https://stockyard.dev/tools/waiver/install.sh | sh
```

Or with Docker:

```bash
docker run -p 9801:9801 -v waiver_data:/data ghcr.io/stockyard-dev/stockyard-waiver
```

Open `http://localhost:9801` in your browser.

## Configuration

| Variable | Default | Description |
|----------|---------|-------------|
| `PORT` | `9801` | HTTP port |
| `DATA_DIR` | `./waiver-data` | SQLite database directory |
| `STOCKYARD_LICENSE_KEY` | *(empty)* | License key for unlimited use |

## Free vs Pro

| | Free | Pro |
|-|------|-----|
| Limits | 5 records | Unlimited |
| Price | Free | Included in bundle or $29.99/mo individual |

Get a license at [stockyard.dev](https://stockyard.dev).

## License

Apache 2.0
