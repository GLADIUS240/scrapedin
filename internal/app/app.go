package app

import (
	"github.com/GLADIUS240/scrapedin/internal/browser"
	"github.com/GLADIUS240/scrapedin/internal/config"
	"github.com/GLADIUS240/scrapedin/internal/database"
	"github.com/GLADIUS240/scrapedin/internal/linkedin"
	"github.com/GLADIUS240/scrapedin/internal/logger"
	"github.com/go-rod/rod"
	"go.uber.org/zap"
)

type App struct {
	cfg     *config.Config
	logger  *zap.Logger
	browser *rod.Browser
}

func New(cfg *config.Config) (*App, error) {
	log, err := logger.New(
		cfg.Logging.Level,
		cfg.Logging.Format,
	)
	if err != nil {
		return nil, err
	}

	br, err := browser.New(browser.Browser{
		Headless:  cfg.Browser.Headless,
		UserAgent: cfg.Browser.UserAgent,
		SlowMs:    cfg.Browser.SlowMoMillis,
		UserData:  cfg.Browser.SessionDir,
	}, log)
	if err != nil {
		return nil, err
	}

	return &App{
		cfg:     cfg,
		logger:  log,
		browser: br,
	}, nil
}

func (a *App) Run() error {
	page := a.browser.MustPage("https://www.linkedin.com/login")
	page.MustWaitLoad()

	state := linkedin.DetectLinkedInAuthState(page)

	switch state {
	case "RETURNING_USER":
		a.logger.Info("Returning user detected")
		linkedin.HandleReturningUser(page)

	case "LOGIN_REQUIRED":
		if err := linkedin.Login(
			page,
			linkedin.Credentials{
				Email:    a.cfg.LinkedIn.Email,
				Password: a.cfg.LinkedIn.Password,
			},
			a.logger,
		); err != nil {
			return err
		}

	case "LOGGED_IN":
		a.logger.Info("Already logged in")
	}

	a.logger.Info("login successful, starting search")

	db, err := database.NewSQLite(a.cfg.DB.DSN)
	if err != nil {
		return err
	}

	err = linkedin.RunPeopleSearch(
		page,
		a.cfg.Search,
		a.logger,
		func(profile string) error {
			return database.Save(db, profile)
		},
	)
	if err != nil {
		return err
	}

	db.Close()
	a.browser.MustClose()

	a.logger.Info("browser closed, app finished")
	return nil
}
