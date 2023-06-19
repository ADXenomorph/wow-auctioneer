package client

import (
    "fmt"
    "strings"

    "github.com/levigross/grequests"
    "github.com/pkg/errors"
)

type Item struct {
    Level         int `json:"level"`
    RequiredLevel int `json:"required_level"`
    SellPrice     int `json:"sell_price"`
    ItemSubclass  struct {
        Name struct {
            EnUS string `json:"en_US"`
        } `json:"name"`
        Id int `json:"id"`
    } `json:"item_subclass"`
    IsEquippable     bool `json:"is_equippable"`
    PurchaseQuantity int  `json:"purchase_quantity"`
    Media            struct {
        Id int `json:"id"`
    } `json:"media"`
    ItemClass struct {
        Name struct {
            EnUS string `json:"en_US"`
        } `json:"name"`
        Id int `json:"id"`
    } `json:"item_class"`
    Quality struct {
        Name struct {
            EnUS string `json:"en_US"`
        } `json:"name"`
        Type string `json:"type"`
    } `json:"quality"`
    MaxCount    int  `json:"max_count"`
    IsStackable bool `json:"is_stackable"`
    Name        struct {
        EnUS string `json:"en_US"`
    } `json:"name"`
    PurchasePrice int `json:"purchase_price"`
    Id            int `json:"id"`
    InventoryType struct {
        Name struct {
            EnUS string `json:"en_US"`
        } `json:"name"`
        Type string `json:"type"`
    } `json:"inventory_type"`
}

type SearchItemResultData struct {
    Data Item `json:"data"`
}

type SearchItemResult struct {
    Results []SearchItemResultData `json:"results"`
}

func (s *SearchItemResult) FindById(id int) (*Item, error) {
    for _, item := range s.Results {
        if item.Data.Id == id {
            return &item.Data, nil
        }
    }

    return nil, errors.New("item not found in SearchItemResult")
}

func (c *client) SearchItem(itemName string) (*SearchItemResult, error) {
    requestURL := c.url + "/data/wow/search/item"

    ro := &grequests.RequestOptions{
        Params: map[string]string{
            "namespace":    fmt.Sprintf("static-%s", c.region),
            "access_token": c.token.AccessToken,
            "locale":       "en_US",
            "_page":        "1",
            "_pageSize":    "1000",
            "name.en_US":   itemName,
        },
    }

    response, err := c.makeGetRequest(requestURL, ro)
    if err != nil {
        return nil, errors.Wrapf(err, "SearchItem makeGetRequest")
    }

    itemData := new(SearchItemResult)
    if err := response.JSON(itemData); err != nil {
        return nil, errors.Wrapf(err, "SearchItem JSON")
    }

    return itemData, nil
}

func (c *client) SearchItemsByIds(ids []int) (*SearchItemResult, error) {
    chunkSize := 100
    if len(ids) <= chunkSize {
        return c.searchItemsByIds(ids)
    } else {
        res := new(SearchItemResult)
        chunks := chunkSlice(ids, chunkSize)
        for _, chunk := range chunks {
            chunkResult, err := c.searchItemsByIds(chunk)
            if err != nil {
                return nil, errors.Wrapf(err, "SearchItemsByIds")
            }
            res.Results = append(res.Results, chunkResult.Results...)
        }

        return res, nil
    }
}

func (c *client) searchItemsByIds(ids []int) (*SearchItemResult, error) {
    requestURL := c.url + "/data/wow/search/item"

    strIds := make([]string, len(ids))
    for i, id := range ids {
        strIds[i] = fmt.Sprint(id)
    }

    ro := &grequests.RequestOptions{
        Params: map[string]string{
            "namespace":    fmt.Sprintf("static-%s", c.region),
            "access_token": c.token.AccessToken,
            "locale":       "en_US",
            "_page":        "1",
            "_pageSize":    "1000",
            "id":           strings.Join(strIds, "||"),
        },
    }

    response, err := c.makeGetRequest(requestURL, ro)
    if err != nil {
        return nil, errors.Wrapf(err, "SearchItem makeGetRequest")
    }

    itemData := new(SearchItemResult)
    if err := response.JSON(itemData); err != nil {
        return nil, errors.Wrapf(err, "SearchItem JSON")
    }

    return itemData, nil
}

func chunkSlice(slice []int, chunkSize int) [][]int {
    var chunks [][]int
    for i := 0; i < len(slice); i += chunkSize {
        end := i + chunkSize

        if end > len(slice) {
            end = len(slice)
        }

        chunks = append(chunks, slice[i:end])
    }

    return chunks
}
