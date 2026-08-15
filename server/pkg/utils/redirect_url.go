package utils

import (
	"net/url"
	"strings"
)

// AppendRedirectQuery adds query parameters to a redirect URL. Hash-router
// callbacks keep their query string inside the URL fragment, so those
// parameters must be added to the fragment instead of the outer URL query.
func AppendRedirectQuery(rawURL string, values url.Values) (string, error) {
	parsed, err := url.Parse(rawURL)
	if err != nil {
		return "", err
	}

	if hashIndex := strings.IndexByte(rawURL, '#'); hashIndex >= 0 && hashIndex < len(rawURL)-1 {
		rawFragment := rawURL[hashIndex+1:]
		fragmentPath := rawFragment
		fragmentQuery := ""
		if index := strings.IndexByte(fragmentPath, '?'); index >= 0 {
			fragmentQuery = fragmentPath[index+1:]
			fragmentPath = fragmentPath[:index]
		}

		query, err := url.ParseQuery(fragmentQuery)
		if err != nil {
			return "", err
		}
		mergeQuery(query, values)
		target := rawURL[:hashIndex+1] + fragmentPath
		if encoded := query.Encode(); encoded != "" {
			target += "?" + encoded
		}
		return target, nil
	}

	query := parsed.Query()
	mergeQuery(query, values)
	parsed.RawQuery = query.Encode()
	return parsed.String(), nil
}

func mergeQuery(target, values url.Values) {
	for key, entries := range values {
		target.Del(key)
		for _, entry := range entries {
			target.Add(key, entry)
		}
	}
}
