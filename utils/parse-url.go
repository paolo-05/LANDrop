package utils

import (
	"log"
	"net/url"
)

// Add below your other imports
func ParseURL(raw string) *url.URL {
	u, err := url.Parse(raw)
	if err != nil {
		log.Println("Invalid URL:", err)
		return &url.URL{}
	}
	return u
}
