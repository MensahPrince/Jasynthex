# Jasynthex

Jasynthex is a powerful Go package for searching arbitrarily nested JSON payloads for specific field names, returning every match with its full structured JSONPath and corresponding value.

Unlike standard JSON unmarshaling, Jasynthex preserves the document order, handles duplicate keys correctly, and provides support for memory-efficient streaming.

## Installation

```bash
go get github.com/MensahPrince/Jasynthex
```

## Features
- **Duplicate Key Support**: Uses a custom ordered map underneath to ensure every key is seen, even duplicates.
- **Document Order**: Returns results strictly in the order they appear (top-to-bottom) in the JSON payload.
- **Flexible Matching**: Find fields using **Exact**, **Case-Insensitive**, or **Contains** (substring) matching modes.
- **Standard JSONPath Notation**: Nested array paths and object keys produce standard JSONPath representations (e.g., `$.data.users[0].profile.contact.email`).
- **Memory-Efficient Streaming**: Process large payload streams directly from HTTP responses without buffering huge files into memory.
- **Structured Paths**: Every returned result gives you access to the individual path segments for programmatic node traversal.

## Quick Start

```go
package main

import (
	"fmt"
	"log"

	jasynthex "github.com/MensahPrince/Jasynthex"
)

func main() {
	rawJSON := []byte(`{
		"organization": {
			"id": 1234,
			"contact": {
				"email": "info@example.org"
			}
		}
	}`)

	// Search for the "email" field with case-insensitive matching
	results, err := jasynthex.Find(rawJSON, "email", jasynthex.CaseInsensitive)
	if err != nil {
		log.Fatalf("Error finding path: %v", err)
	}

	for _, r := range results {
		fmt.Printf("Path: %s\n", r.PathString)
		fmt.Printf("Value: %v\n", r.Value)
		fmt.Printf("Depth: %d\n", r.Depth())
	}
}
```

## API Interception / JSON Processing Example

This package is incredibly useful when intercepting deep, undocumented API responses where you just want to extract a nested key (like an email) without writing huge nested structs.

Here is an example of intercepting a deeply nested JSON payload and extracting paths:

```go
package main

import (
	"fmt"
	"log"

	jasynthex "github.com/MensahPrince/Jasynthex"
)

func main() {
	interceptedResponse := []byte(`{
		"meta": { "status": "ok", "version": "2.1" },
		"data": {
			"users": [
				{
					"id": "u001",
					"profile": {
						"name": "Ama Boateng",
						"contact": {
							"email": "ama@clinic.com",
							"phone": "0244000001"
						}
					},
					"billing": { "email": "ama-billing@clinic.com" }
				},
				{
					"id": "u002",
					"profile": {
						"name": "Kofi Asante",
						"contact": {
							"email": "kofi@clinic.com",
							"phone": "0244000002"
						}
					},
					"billing": { "email": "kofi-billing@clinic.com" }
				}
			]
		}
	}`)

	// Search matching any field containing "email"
	results, err := jasynthex.Find(interceptedResponse, "email", jasynthex.Contains)
	if err != nil {
		log.Fatalf("Failed to parse JSON: %v", err)
	}

	fmt.Println("Extracted Email Paths:")
	for _, r := range results {
		fmt.Printf("%-55s -> %v\n", r.PathString, r.Value)
	}
}
```

## Large Files and Streaming
For massive JSON payloads, use `FindStream` to process directly from an `io.Reader` without fully loading the buffer:

```go
import "encoding/json"
import jasynthex "github.com/MensahPrince/Jasynthex"

func ProcessLargeResponse(body io.Reader) {
	decoder := json.NewDecoder(body)
	results, err := jasynthex.FindStream(decoder, "email", jasynthex.CaseInsensitive)
	// ...
}
```


## Matching Modes

1. **`jasynthex.Exact`** (Default): The key must exactly match the search string character-for-character.
2. **`jasynthex.CaseInsensitive`**: Bypasses casing difference (e.g. `Email` == `email`).
3. **`jasynthex.Contains`**: Will return successful matches if the search key substring exists inside any JSON key (e.g. `contactEmail`).

## For any encoutered difficulties, be it bugs or questions, kindly leave behind an issue.
### Thank you.
