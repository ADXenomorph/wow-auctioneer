package client

import (
    "fmt"
    "strings"
    "time"

    "github.com/pkg/errors"
    "github.com/sirupsen/logrus"

    "github.com/levigross/grequests"
)

type BlizzardToken struct {
    AccessToken string `json:"access_token"`
    TokenType   string `json:"token_type"`
    ExpiresIn   int    `json:"expires_in"`
    ExpiresAt   int64  `json:"expires_at,omitempty"`
}

func (c *client) SetToken(token *BlizzardToken) {
    c.token = token
}

func (c *client) MakeBlizzAuth() (*BlizzardToken, error) {
    body := strings.NewReader("grant_type=client_credentials")
    ro := &grequests.RequestOptions{
        RequestBody: body,
        Auth:        []string{c.cfg.ClientID, c.cfg.ClientSecret},
        Headers: map[string]string{
            "Content-Type": "application/x-www-form-urlencoded",
        },
    }

    c.log.Log(logrus.InfoLevel, fmt.Sprintf(
        "Blizz POST request. Url %s, params: %v, headers: %v", c.cfg.AUTHUrl, ro.Params, ro.Headers,
    ))

    response, err := c.session.Post(c.cfg.AUTHUrl, ro)
    if err != nil {
        return nil, errors.Wrapf(err, "MakeBlizzAuth POST")
    }

    tokenData := new(BlizzardToken)
    if err := response.JSON(tokenData); err != nil {
        return nil, errors.Wrapf(err, "MakeBlizzAuth Parse JSON")
    }

    duration, err := time.ParseDuration(fmt.Sprintf("%ds", tokenData.ExpiresIn))
    if err != nil {
        return nil, errors.Wrapf(err, "MakeBlizzAuth ParseDuration")
    }
    tokenData.ExpiresAt = time.Now().Add(duration).Unix()

    return tokenData, nil
}
