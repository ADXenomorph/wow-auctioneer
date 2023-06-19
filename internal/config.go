package internal

import (
    "github.com/kelseyhightower/envconfig"

    "github.com/ADXenomorph/wow-auctioneer/internal/client"
)

type Config struct {
    LogLvl         string `envconfig:"LOG_LEVEL" default:"INFO"`
    TelegramToken  string `envconfig:"TELEGRAM_TOKEN" required:"true"`
    TelegramChatId string `envconfig:"TELEGRAM_CHAT_ID" required:"true"`
    client.BlizzApiCfg
}

func NewConfig() (*Config, error) {
    cfg := new(Config)
    cfg.RegionList = []string{"eu" /*, "us"*/}

    if err := envconfig.Process("AUCTIONEER", cfg); err != nil {
        return nil, err
    }

    return cfg, nil
}
