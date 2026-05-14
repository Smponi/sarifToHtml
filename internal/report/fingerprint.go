package report

import (
	"crypto/sha1"
	"encoding/hex"
	"sort"
	"strings"
)

func stableFingerprint(parts ...string) string {
	hash := sha1.New()
	for _, part := range parts {
		hash.Write([]byte(part))
		hash.Write([]byte{0})
	}
	return hex.EncodeToString(hash.Sum(nil))
}

func firstFingerprint(values map[string]string, preferred ...string) string {
	for _, key := range preferred {
		if value := values[key]; value != "" {
			return value
		}
	}

	keys := make([]string, 0, len(values))
	for key := range values {
		keys = append(keys, key)
	}
	sort.Strings(keys)

	for _, key := range keys {
		if strings.TrimSpace(values[key]) != "" {
			return values[key]
		}
	}

	return ""
}
