# telehtml

`telehtml` is a Go utility for cleaning and formatting HTML content to match Telegram message formatting requirements.

## Features

- Removes invisible characters and junk spaces
- Converts unsupported or inconsistent HTML to Telegram-safe tags
- Normalizes line breaks, spaces, and links
- Protects against XSS using `bluemonday` sanitizer

## Installation

```bash
go get github.com/yourusername/telehtml
```

## Usage

```go
import "github.com/svanichkin/telehtml"

cleaned := telehtml.CleanTelegramHTML(rawHTML)
```