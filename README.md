# geth-relay

A high-performance Ethereum RPC proxy server written in Go that enhances standard JSON-RPC functionality.

## Features

- **Standard RPC Proxy**: Forward all standard Ethereum JSON-RPC methods to upstream geth node
- **Batch Request Support**: Handle multiple RPC calls in a single HTTP request (JSON-RPC 2.0 batch)
- **Request Size Limits**: Configurable limits matching geth defaults (5MB body, 100 batch items)
- **Enhanced Error Handling**: Geth-compatible error codes and timeout detection
- **Structured Logging**: Comprehensive logging with zap
- **Configuration Management**: YAML-based configuration with environment variable support
- **Health Checks**: Built-in health endpoint for monitoring
- **Graceful Shutdown**: Proper signal handling and graceful shutdown
- **Middleware**: Request logging, panic recovery, duration tracking

## Installation

### Prerequisites

- Go 1.21 or higher
- A running Ethereum node (geth) on port 8546

### Build from source

```bash
# Clone the repository
git clone https://github.com/devlongs/geth-relay.git
cd geth-relay

# Install dependencies
make install-deps

# Build the binary
make build
```

## Configuration

Create a `config.yaml` file or use environment variables:

```yaml
server:
  host: "0.0.0.0"
  port: 8545

upstream:
  url: "http://localhost:8546"
  timeout: 30s

logging:
  level: "info"
  format: "json"
```

See `configs/config.example.yaml` for all available options.

### Configuration Options

- **`server.host`**: Listen address (default: `0.0.0.0`)
- **`server.port`**: Port to listen on (default: `8545`)
- **`upstream.url`**: Upstream geth node URL (default: `http://localhost:8546`)
- **`upstream.timeout`**: Request timeout (default: `30s`)
- **`logging.level`**: Log level - `debug`, `info`, `warn`, `error` (default: `info`)
- **`logging.format`**: Log format - `json` or `console` (default: `json`)
- **`limits.max_body_size`**: Max request body size in bytes (default: `5242880` = 5MB)
- **`limits.max_batch_items`**: Max items in batch request (default: `100`)
- **`limits.max_batch_response`**: Max batch response size in bytes (default: `25000000` = 25MB)

## Usage

### Start the server

```bash
# Run with default config (looks for config.yaml in current dir or ./configs/)
./geth-relay

# Run with specific config file
./geth-relay -config /path/to/config.yaml

# Or use make
make run
```

### Test the proxy

```bash
# Single request
curl -X POST http://localhost:8545 \
  -H "Content-Type: application/json" \
  -d '{
    "jsonrpc": "2.0",
    "method": "eth_blockNumber",
    "params": [],
    "id": 1
  }'

# Batch request (multiple calls in one request)
curl -X POST http://localhost:8545 \
  -H "Content-Type: application/json" \
  -d '[
    {"jsonrpc":"2.0","method":"eth_blockNumber","params":[],"id":1},
    {"jsonrpc":"2.0","method":"eth_gasPrice","params":[],"id":2},
    {"jsonrpc":"2.0","method":"eth_chainId","params":[],"id":3}
  ]'

# Health check
curl http://localhost:8545/health
```

## Supported RPC Methods

All standard Ethereum JSON-RPC methods are supported:

- `eth_blockNumber`
- `eth_getBalance`
- `eth_getBlockByNumber`
- `eth_getBlockByHash`
- `eth_getTransactionByHash`
- `eth_call`
- `eth_getLogs`
- `eth_sendRawTransaction`
- `eth_gasPrice`
- `eth_estimateGas`
- And all other standard methods...

## Development

```bash
# Run tests
make test

# Format code
make fmt

# Run with live reload (requires air)
air

# Clean build artifacts
make clean
```



## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

## License

MIT License
