package main

import (
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"

	"github.com/liuminhaw/yatijapp-tui/internal/authclient"
	"github.com/liuminhaw/yatijapp-tui/internal/data"
	flag "github.com/spf13/pflag"
	"github.com/spf13/viper"
	"gopkg.in/natefinch/lumberjack.v2"
)

var defaultConfigFile = "config.toml"

type config struct {
	apiEndpoint string // http://yatijapp.server.url

	displayMode string // light | dark | auto

	preferences *data.Preferences

	logger     *slog.Logger
	authClient *authclient.AuthClient
}

func configSetup(
	conf *viper.Viper,
) (config, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return config{}, err
	}

	rotater := &lumberjack.Logger{
		Filename:   fmt.Sprintf("%s/.yatijapp/tui.log", homeDir),
		MaxSize:    100,
		MaxBackups: 3,
		MaxAge:     30,
		Compress:   false,
	}
	logger := slog.New(slog.NewJSONHandler(rotater, nil))

	conf.SetDefault("api.endpoint", "https://api.yatij.app")
	conf.SetDefault("preference.displayMode", "auto")

	conf.SetConfigName(defaultConfigFile)
	conf.SetConfigType("toml")
	conf.AddConfigPath(filepath.Join(homeDir, ".yatijapp"))

	if err := conf.ReadInConfig(); err != nil {
		logger.Info("No configuration file found, using defaults")
	}

	conf.BindPFlag("api.endpoint", flag.Lookup("api-endpoint"))
	conf.BindPFlag("preference.displayMode", flag.Lookup("display-mode"))

	client := &authclient.AuthClient{
		Client:    http.DefaultClient,
		Refresh:   authclient.RefreshToken(conf.GetString("api.endpoint")),
		TokenPath: filepath.Join(homeDir, ".yatijapp", "creds", "token.json"),
	}

	return config{
		apiEndpoint: conf.GetString("api.endpoint"),
		displayMode: conf.GetString("preference.displayMode"),
		logger:      logger,
		authClient:  client,
	}, nil
}
