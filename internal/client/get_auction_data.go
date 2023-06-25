package client

import (
    "fmt"
    "strings"
    "time"

    "github.com/levigross/grequests"
    "github.com/pkg/errors"
)

type ItemAssets struct {
    Key        string `json:"key"`
    Value      string `json:"value"`
    FileDataID int    `json:"file_data_id"`
}

type ItemMedia struct {
    Assets []ItemAssets `json:"assets"`
    ID     int          `json:"id"`
}

type AucItemModifier struct {
    Type  int `json:"type"`
    Value int `json:"value"`
}

type AucItem struct {
    ID           int               `json:"id"`
    Context      int               `json:"context"`
    BonusLists   []int             `json:"bonus_lists"`
    Modifiers    []AucItemModifier `json:"modifiers"`
    PetBreedID   int               `json:"pet_breed_id"`
    PetLevel     int               `json:"pet_level"`
    PetQualityID int               `json:"pet_quality_id"`
    PetSpeciesID int               `json:"pet_species_id"`
}

type AuctionsDetail struct {
    ID       int     `json:"id"`
    Item     AucItem `json:"item"`
    Buyout   int     `json:"buyout"`
    Quantity int     `json:"quantity"`
    TimeLeft string  `json:"time_left"`
    ItemName struct {
        EnUS string `json:"en_US"`
    } `json:"item_name"`
    Quality string `json:"quality"`
    Price   int    `json:"unit_price"`
}

type AuctionData struct {
    Auctions  []*AuctionsDetail `json:"auctions"`
    ExpiresAt int64             `json:"expires_at,omitempty"`
}

func (c *client) GetAuctionData(realmID int) (*AuctionData, error) {
    requestURL := c.url + fmt.Sprintf("/data/wow/connected-realm/%d/auctions", realmID)
    ro := &grequests.RequestOptions{
        Params: map[string]string{
            "namespace":    fmt.Sprintf("dynamic-%s", c.region),
            "access_token": c.token.AccessToken,
        },
    }
    response, err := c.makeGetRequest(requestURL, ro)
    if err != nil {
        return nil, errors.Wrapf(err, "GetAuctionData makeGetRequest")
    }

    auctionData := new(AuctionData)
    if err := response.JSON(auctionData); err != nil {
        return nil, errors.Wrapf(err, "GetAuctionData JSON")
    }

    updatedAt := response.Header.Get("last-modified") // GMT! (-3)
    updatedAtParsed, err := time.Parse(layoutUS, updatedAt)
    if err != nil {
        return nil, errors.Wrapf(err, "GetAuctionData Parse")
    }
    auctionData.ExpiresAt = updatedAtParsed.Add(time.Hour).Unix()
    c.log.Infof("Auction expires at %v", auctionData.ExpiresAt)

    return auctionData, nil
}

func (ad *AuctionData) filter(filterFunc func(auc *AuctionsDetail) bool) *AuctionData {
    res := make([]*AuctionsDetail, 0)
    for _, auction := range ad.Auctions {
        if filterFunc(auction) {
            res = append(res, auction)
        }
    }

    return &AuctionData{Auctions: res, ExpiresAt: ad.ExpiresAt}
}

func (ad *AuctionData) FilterByName(name string) *AuctionData {
    return ad.filter(func(auc *AuctionsDetail) bool {
        return strings.Contains(strings.ToLower(auc.ItemName.EnUS), strings.ToLower(name))
    })
}

func (ad *AuctionData) FilterByItemId(id int) *AuctionData {
    return ad.filter(func(auc *AuctionsDetail) bool {
        return auc.Item.ID == id
    })
}

func (ad *AuctionData) FilterByBuyout(moreOrEq int) *AuctionData {
    return ad.filter(func(auc *AuctionsDetail) bool {
        return auc.Buyout >= moreOrEq
    })
}
