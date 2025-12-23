package browser

import (
	"time"

	"github.com/go-rod/rod"
	"github.com/go-rod/rod/lib/launcher"
	"go.uber.org/zap"
)

type Browser struct {
	Headless  bool
	UserAgent string
	SlowMs    int
	UserData  string
}

func New(cfg Browser, logger *zap.Logger) (*rod.Browser, error) {
	logger.Info("launching browser",
		zap.Bool("headless", cfg.Headless),
		zap.String("user_data_dir", cfg.UserData),
	)

	l := launcher.New().
		Headless(cfg.Headless).
		UserDataDir(cfg.UserData).
		Leakless(false). // avoids zombie chrome
		NoSandbox(true)

	// Stealth / anti-detection flags
	l.Set("disable-blink-features", "AutomationControlled")
	l.Set("disable-infobars")
	l.Set("disable-dev-shm-usage")

	if cfg.UserAgent != "" {
		l.Set("user-agent", cfg.UserAgent)
	}

	url, err := l.Launch()
	if err != nil {
		return nil, err
	}

	browser := rod.New().
		ControlURL(url).
		SlowMotion(time.Duration(cfg.SlowMs) * time.Millisecond)

	if err := browser.Connect(); err != nil {
		return nil, err
	}

	return browser, nil
}

func Close(br *rod.Browser) {
	if br != nil {
		_ = br.Close()
	}
}
