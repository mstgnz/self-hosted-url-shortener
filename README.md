# Self-Hosted URL Shortener

A simple, self-hosted URL shortener built with Go and SQLite. Create short, easy-to-share links with a clean web interface, API, and CLI support.

## Features

- **Simple Web Interface**: Easily create, manage, and track shortened URLs
- **Custom Short Codes**: Create memorable, branded short links
- **QR Code Generation**: Generate QR codes for your shortened URLs
- **Click Tracking**: Track how many times your shortened URLs have been clicked
- **API Support**: Programmatically create and manage shortened URLs
- **CLI Support**: Command-line interface for URL shortening
- **Self-Hosted**: All your data stays on your server with SQLite
- **Single Binary**: Easy to deploy with no external dependencies

## Installation

### Download Binary

Download the latest release from the [releases page](https://github.com/mstgnz/self-hosted-url-shortener/releases).

### Build from Source

```bash
git clone https://github.com/mstgnz/self-hosted-url-shortener.git
cd self-hosted-url-shortener
go build -o url-shortener ./cmd
```

## Usage

### Web Interface

Start the server:

```bash
./url-shortener --port 8080 --base-url "https://your-domain.com"
```

Then open your browser and navigate to `http://localhost:8080` (or your custom domain).

### API

The URL shortener provides a RESTful API:

#### Create a shortened URL

```bash
curl -X POST http://localhost:8080/api/shorten \
  -H "Content-Type: application/json" \
  -d '{"url": "https://example.com/very/long/url", "custom_code": "my-link"}'
```

#### List all URLs

```bash
curl -X GET http://localhost:8080/api/urls
```

#### Get URL details

```bash
curl -X GET http://localhost:8080/api/url/my-link
```

#### Delete a URL

```bash
curl -X DELETE http://localhost:8080/api/url/my-link
```

### CLI

The URL shortener also provides a command-line interface:

#### Shorten a URL

```bash
./url-shortener --cli shorten https://example.com/very/long/url
```

With a custom code:

```bash
./url-shortener --cli shorten https://example.com/very/long/url --code my-link
```

#### List all URLs

```bash
./url-shortener --cli list
```

#### Get URL details

```bash
./url-shortener --cli get my-link
```

#### Generate a QR code

```bash
./url-shortener --cli qr my-link --output qr.png
```

#### Delete a URL

```bash
./url-shortener --cli delete my-link
```

## Configuration

The URL shortener can be configured using command-line flags:

- `--port`: HTTP server port (default: 8080)
- `--db`: SQLite database path (default: data.db)
- `--base-url`: Base URL for shortened URLs (default: http://localhost:8080)
- `--templates`: Templates directory (default: templates)
- `--cli`: Run in CLI mode

## Development

### Prerequisites

- Go 1.18 or higher
- SQLite

### Setup

1. Clone the repository:

```bash
git clone https://github.com/mstgnz/self-hosted-url-shortener.git
cd self-hosted-url-shortener
```

2. Install dependencies:

```bash
go mod download
```

3. Run the application:

```bash
go run ./cmd
```

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.
