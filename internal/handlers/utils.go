package handlers

import (
	"fmt"
	"net/http"
	"strconv"
	"time"

	paste "github.com/infinage/pastebin/pkg"
)

type modelFormFields struct {
	content    string
	expiry     time.Duration
	visibility paste.Visibility
}

func parseForm(r *http.Request) (modelFormFields, error) {
	if err := r.ParseForm(); err != nil {
		return modelFormFields{}, fmt.Errorf("Error parsing form")
	}

	visibilityRaw := r.FormValue("visibility")
	visibility, err := strconv.Atoi(visibilityRaw)
	if err != nil || visibility != 1 && visibility != 2 {
		return modelFormFields{}, fmt.Errorf("Visibility field must be 1 (public) or 2 (unlisted)")
	}

	expiryMinutesRaw := r.FormValue("expiry")
	expiryMinutes, err := strconv.Atoi(expiryMinutesRaw)
	if err != nil || expiryMinutes < 1 || expiryMinutes > 110376000 {
		return modelFormFields{}, fmt.Errorf("Expiry field must be between [1, 110376000]")
	}

	content := r.FormValue("content")
	expiry := time.Minute * time.Duration(expiryMinutes)
	res := modelFormFields{content: content, expiry: expiry, visibility: paste.Visibility(visibility)}
	return res, nil
}
