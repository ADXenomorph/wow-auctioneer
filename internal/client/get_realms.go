package client

import (
    "fmt"

    "github.com/levigross/grequests"
    "github.com/pkg/errors"
)

type BlizzRealmsSearchResultResultsDataRealmsName struct {
    RuRU string `json:"ru_RU"`
    EnGB string `json:"en_GB"`
}

type BlizzRealmsSearchResultResultsDataRealms struct {
    Name BlizzRealmsSearchResultResultsDataRealmsName `json:"name"`
}

type BlizzRealmsSearchResultResultsData struct {
    Realms []BlizzRealmsSearchResultResultsDataRealms `json:"realms"`
    ID     int                                        `json:"id"`
}

type BlizzRealmsSearchResultResults struct {
    Data BlizzRealmsSearchResultResultsData `json:"data"`
}

type BlizzRealmsSearchResult struct {
    Results []BlizzRealmsSearchResultResults `json:"results"`
}

func (c *client) GetBlizzRealms() (*BlizzRealmsSearchResult, error) {
    requestURL := c.url + "/data/wow/search/connected-realm"

    ro := &grequests.RequestOptions{
        Params: map[string]string{
            "namespace":    fmt.Sprintf("dynamic-%s", c.region),
            "access_token": c.token.AccessToken,
            "locale":       "en_US",
        },
    }

    response, err := c.makeGetRequest(requestURL, ro)
    if err != nil {
        return nil, errors.Wrapf(err, "getBlizzRealms makeGetRequest")
    }

    realmData := new(BlizzRealmsSearchResult)
    if err := response.JSON(realmData); err != nil {
        return nil, errors.Wrapf(err, "getBlizzRealms JSON")
    }

    return realmData, nil
}

func (c *client) GetRealmID(realmName string) (int, error) {
    realms, err := c.GetBlizzRealms()
    if err != nil {
        return 0, errors.Wrap(err, "GetRealmID GetBlizzRealms")
    }

    for _, connectedRealm := range realms.Results {
        for _, realm := range connectedRealm.Data.Realms {
            if realm.Name.EnGB == realmName {
                return connectedRealm.Data.ID, nil
            }
        }
    }

    return 0, errors.New(fmt.Sprintf("Realm %q not found", realmName))
}
