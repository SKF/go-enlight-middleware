package extractor

import (
	"fmt"
	"net/http"
	"strings"

	custom_problems "github.com/SKF/go-enlight-middleware/v1/client-id/problems"
)

// Default extractor is using the request header X-Client-ID
var Default Extractor = HeaderExtractor{"X-Client-ID"}

// Extractor extracts an client id identifier from an HTTP request
type Extractor interface {
	ExtractClientID(*http.Request) (string, error)
}

type HeaderExtractor []string

func (e HeaderExtractor) ExtractClientID(r *http.Request) (string, error) {
	for _, header := range e {
		if value := r.Header.Get(header); value != "" {
			return value, nil
		}
	}

	return "", custom_problems.NoClientID(fmt.Sprintf(
		"Should be provided in the request header(s): %s.",
		strings.Join(e, ", "),
	))
}
