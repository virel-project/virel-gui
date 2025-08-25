//go:build js
// +build js

package lang

import (
	"errors"
	"syscall/js"
)

// DetectIETF attempts to detect the browser's preferred language
// Returns IETF language tag (e.g., "en-US", "fr-FR") or error
func DetectIETF() (string, error) {
	// Get the global window object
	window := js.Global()

	// Try navigator.languages first (most modern approach)
	if languages := window.Get("navigator").Get("languages"); !languages.IsUndefined() {
		length := languages.Length()
		if length > 0 {
			// Return the first preferred language
			return languages.Index(0).String(), nil
		}
	}

	// Fallback to navigator.language
	if language := window.Get("navigator").Get("language"); !language.IsUndefined() {
		return language.String(), nil
	}

	return "", errors.New("could not detect browser language")
}
