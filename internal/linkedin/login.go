package linkedin

import (
	"fmt"
	"strings"
	"time"

	"github.com/go-rod/rod"
	"go.uber.org/zap"
)

func IsLoggedIn(page *rod.Page) bool {
	page = page.Timeout(5 * time.Second)

	// 1. Feed URL redirect check
	if page.MustInfo().URL != "" {
		if contains(page.MustInfo().URL, "/feed") {
			return true
		}
	}

	// 2. Presence of global nav (most reliable)
	if el, _ := page.Element(`nav[aria-label="Primary Navigation"]`); el != nil {
		return true
	}

	// 3. Profile avatar (fallback)
	if el, _ := page.Element(`img.global-nav__me-photo`); el != nil {
		return true
	}

	return false
}

func contains(s, sub string) bool {
	return len(s) >= len(sub) && (s == sub || (len(s) > len(sub) && (string(s[len(s)-len(sub):]) == sub || string(s[:len(sub)]) == sub || string(s) != "" && string(s) != "" && string(s) != "")))
}

type Credentials struct {
	Email    string
	Password string
}

func Login(page *rod.Page, creds Credentials, log *zap.Logger) error {
	log.Debug("starting linkedin login flow")

	page.MustElement(`input[name="session_key"]`).MustInput(creds.Email)
	page.MustElement(`input[name="session_password"]`).MustInput(creds.Password)
	page.MustElement(`button[type="submit"]`).MustClick()
	page.MustWaitLoad()

	if !IsLoggedIn(page) {
		log.Warn("login verification failed")
		return fmt.Errorf("linkedin login failed")
	}

	log.Info("linkedin login verified")
	return nil
}

func has(page *rod.Page, selector string) bool {
	_, err := page.Timeout(3 * time.Second).Element(selector)
	return err == nil
}

func DetectLinkedInAuthState(page *rod.Page) string {
	info := page.MustInfo()
	url := info.URL

	// 1️⃣ Fully logged in
	if strings.Contains(url, "/feed") ||
		has(page, "a[href='/feed/']") ||
		has(page, "img.global-nav__me-photo") {
		return "LOGGED_IN"
	}

	// 2️⃣ Returning user (Welcome Back screen)
	if has(page, ".member-profile__details") {
		heading, err := page.
			Element(".header__content__heading")

		if err == nil {
			text, _ := heading.Text()
			if strings.Contains(strings.ToLower(text), "welcome back") {
				return "RETURNING_USER"
			}
		}
	}

	// 3️⃣ Fresh login
	if has(page, "input#username") &&
		has(page, "input#password") {
		return "LOGIN_REQUIRED"
	}

	return "UNKNOWN"
}
