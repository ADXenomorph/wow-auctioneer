package client

//go:generate mkdir -p mocks
//go:generate rm -rf ./mocks/*_minimock.go
//go:generate minimock -i Client -o ./mocks/ -s "_minimock.go"

import (
    "context"
    "crypto/tls"
    "fmt"
    "net/http"
    "time"

    "github.com/pkg/errors"

    "github.com/levigross/grequests"
    "github.com/sirupsen/logrus"
)

const (
    layoutUS = "Mon, 2 Jan 2006 15:04:05 MST"
    timeOut  = time.Second * 10
)

type Client interface {
    GetBlizzRealms() (*BlizzRealmsSearchResult, error)
    MakeBlizzAuth() (*BlizzardToken, error)
    SetToken(token *BlizzardToken)
    GetRealmID(name string) (int, error)
    SearchItem(itemName string) (*SearchItemResult, error)
    GetAuctionData(realmID int) (*AuctionData, error)
    SearchItemsByIds(ids []int) (*SearchItemResult, error)
    GetBonuses() (*Bonuses, error)
}

type client struct {
    token   *BlizzardToken
    cfg     *BlizzApiCfg
    session *grequests.Session
    url     string
    region  string
    ctx     context.Context
    log     *logrus.Logger
}

func NewClient(ctx context.Context, logger *logrus.Logger, blizzCfg *BlizzApiCfg, region string) Client {
    urlsMap := make(map[string]string)
    urlsMap["eu"] = blizzCfg.EuAPIUrl
    urlsMap["us"] = blizzCfg.UsAPIUrl

    session := grequests.NewSession(nil)
    session.HTTPClient.Transport = &http.Transport{
        TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
    }
    session.HTTPClient.Timeout = timeOut

    return &client{
        cfg:     blizzCfg,
        session: session,
        url:     urlsMap[region],
        region:  region,
        ctx:     ctx,
        log:     logger,
    }
}

func (c *client) makeGetRequest(requestURL string, ro *grequests.RequestOptions) (*grequests.Response, error) {
    c.log.Log(logrus.InfoLevel, fmt.Sprintf(
        "Blizz request. Url %s, params: %v, headers: %v",
        requestURL,
        ro.Params,
        ro.Headers,
    ))

    ro.RequestTimeout = 30 * time.Second
    response, err := c.session.Get(requestURL, ro)
    if err != nil {
        return nil, errors.Wrapf(err, "makeGetRequest")
    }
    if !response.Ok {
        return nil, errors.Wrapf(fmt.Errorf(
            "error making get request, status: %v", response.StatusCode,
        ), "makeGetRequest")
    }

    return response, nil
}
