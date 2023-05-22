package helpers

import (
	"os"
	"strings"
)

// EnforceHTTP adds http:// to the url
func EnforceHTTP(url string) string {
	if url[:4] != "http" {
		url = "http://" + url
	}
	return url
}

// RemoveDomainError checks that given url is not current domain
func RemoveDomainError(url string) bool {
	if url == os.Getenv("DOMAIN") {
		return false
	}

	newURL := strings.Replace(url, "http://", "", 1)
	newURL = strings.Replace(newURL, "https://", "", 1)
	newURL = strings.Replace(newURL, "www.", "", 1)
	newURL = strings.Split(newURL, "/")[0]

	if url == os.Getenv("DOMAIN") {
		return false
	}

	return true
}
