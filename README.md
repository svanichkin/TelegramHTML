# Telegram HTML

`telehtml` — Go utility for cleaning and formatting HTML to match Telegram message formatting requirements.

## Features

- Removes invisible characters and junk spaces
- Converts unsupported or inconsistent HTML to Telegram-safe tags
- Normalizes line breaks, spaces, and links
- Protects against XSS using `bluemonday` sanitizer

## Installation

```bash
go get github.com/svanichkin/TelegramHTML
```

## Usage

```go
import "github.com/svanichkin/TelegramHTML"

cleanedHTML := telehtml.CleanTelegramHTML(rawHTML)
splittedHTML := telehtml.SplitTelegramHTML(cleaned)
```

# Splitter:

The splitter divides a long Telegram HTML message into smaller chunks, each no longer than 4000 characters. It attempts to cut at optimal points to avoid breaking HTML tags or words awkwardly. It looks for closing tags like `</a>, </b>, </code>`, and natural breaks like newlines or punctuation to split cleanly. If no suitable break point is found, it cuts at the maximum length.

It also ensures that any opened HTML tags before the cut are properly closed, so each chunk remains valid HTML.

This allows sending long formatted Telegram messages without exceeding Telegram’s message length limits, while preserving formatting.

# Invisible Int Tag Encoder/Decoder

This package provides functions to encode and decode integers using invisible Unicode characters.
It can be used for hiding numeric data inside text without visible changes.

## Features

- EncodeIntInvisible(n int) string — converts an integer to a string of invisible Unicode characters.
- DecodeIntInvisible(s string) int — decodes a string of invisible characters back to the original integer.
- findInvisibleUidSequences(s string) []string — extracts sequences of invisible characters from any input string.

## How It Works

- Uses a predefined set of zero-width and invisible Unicode runes to represent digits 0-9.
- Maps runes to digits and vice versa for encoding/decoding.
- Efficient initialization of rune maps at package load time.

## Use Cases

- Steganography in text.
- Hidden tags or metadata embedding.
- Invisible markers for text processing.
