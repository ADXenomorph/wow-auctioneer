package internal

import (
    "github.com/sirupsen/logrus"
)

func NewLogger(logLvl string) (*logrus.Logger, error) {
    lvl, err := logrus.ParseLevel(logLvl)
    if err != nil {
        return nil, err
    }

    logger := logrus.New()
    logger.SetLevel(lvl)
    //logger.SetFormatter(&logrus.JSONFormatter{PrettyPrint: true})
    //logger.SetFormatter(&logrus.JSONFormatter{})
    logger.SetFormatter(&logrus.TextFormatter{})

    return logger, nil
}
