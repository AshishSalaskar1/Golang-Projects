package helpers

import (
	"os"
	"strings"
)

// RemveDomainError checks if user has given the application domain name itself to be shortened
// if yes, it returns false. but you also need to check all possibilities
// dont allow localhost to be a domain
func RemoveDomainError(url string) bool {
	if url == os.Getenv("DOMAIN") {
		return false
	}

	cleanURL := strings.Replace(url, "http://", "", 1)
	cleanURL = strings.Replace(cleanURL, "https://", "", 1)
	cleanURL = strings.Replace(cleanURL, "www.", "", 1)
	cleanURL = strings.Split(cleanURL, "/")[0]

	if cleanURL == os.Getenv("DOMAIN") {
		return false
	}

	return true

}

func EnforceHTTPS(s string) string {
	return strings.Replace(s, "http://", "https://", 1)
}
