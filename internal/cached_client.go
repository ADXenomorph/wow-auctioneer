package internal

import (
    "fmt"
    "strconv"
    "time"

    "github.com/pkg/errors"
    "github.com/sirupsen/logrus"

    "github.com/ADXenomorph/wow-auctioneer/internal/client"
    "github.com/ADXenomorph/wow-auctioneer/internal/pcache"
)

type cachedClient struct {
    client client.Client
    cache  *pcache.PCache
    log    *logrus.Logger
    region string
}

func NewCachedClient(client client.Client, cache *pcache.PCache, log *logrus.Logger, region string) client.Client {
    return &cachedClient{client, cache, log, region}
}

func (c *cachedClient) GetBlizzRealms() (*client.BlizzRealmsSearchResult, error) {
    cacheKey := fmt.Sprintf("GetBlizzRealms:%s", c.region)

    var cache client.BlizzRealmsSearchResult
    err := c.cache.PGetStruct(cacheKey, &cache)
    if err == nil {
        c.log.Info("Got realms from cache")
        return &cache, nil
    }

    realms, err := c.client.GetBlizzRealms()
    if err != nil {
        return nil, errors.Wrap(err, "cached client GetBlizzRealms")
    }

    err = c.cache.PSetStruct(cacheKey, realms)
    if err != nil {
        return nil, errors.Wrap(err, "cached client GetBlizzRealms cache set")
    }

    return realms, nil
}

func (c *cachedClient) GetRealmID(name string) (int, error) {
    realmIdCacheKey := fmt.Sprintf("GetRealmID:%s:%s", c.region, name)

    data, err := c.cache.PGet(realmIdCacheKey)
    if err == nil {
        res, err := strconv.Atoi(string(data))
        if err == nil {
            c.log.Infof("Got realm id for %q from cache", name)
            return res, nil
        }
    }

    id, err := c.client.GetRealmID(name)
    if err != nil {
        return 0, errors.Wrap(err, "cached client GetBlizzRealms")
    }

    c.cache.PSet(realmIdCacheKey, []byte(strconv.Itoa(id)))

    return id, nil
}

func (c *cachedClient) SetToken(token *client.BlizzardToken) {
    c.client.SetToken(token)
}

func (c *cachedClient) MakeBlizzAuth() (*client.BlizzardToken, error) {
    cacheKey := "MakeBlizzAuth"

    var res client.BlizzardToken
    err := c.cache.PGetStruct(cacheKey, &res)
    if err == nil {
        if res.ExpiresAt < time.Now().Unix() {
            c.log.Info("Token from cache is expired")
        } else {
            c.log.Info("Got token from cache")
            return &res, nil
        }
    }

    token, err := c.client.MakeBlizzAuth()
    if err != nil {
        return nil, errors.Wrap(err, "cached client MakeBlizzAuth")
    }

    err = c.cache.PSetStruct(cacheKey, token)
    if err != nil {
        return nil, errors.Wrap(err, "cached client MakeBlizzAuth cache set")
    }

    return token, nil
}

func (c *cachedClient) SearchItem(itemName string) (*client.SearchItemResult, error) {
    return c.client.SearchItem(itemName)
}

func (c *cachedClient) GetAuctionData(realmID int) (*client.AuctionData, error) {
    cacheKey := fmt.Sprintf("GetAuctionData:%d:%s", realmID, c.region)

    var res client.AuctionData
    err := c.cache.PGetStruct(cacheKey, &res)
    if err == nil {
        if res.ExpiresAt < time.Now().Unix() {
            c.log.Infof("AH data for realm %d region %s from cache is expired", realmID, c.region)
        } else {
            c.log.Infof("Got AH data for realm %d region %s from cache", realmID, c.region)
            return &res, nil
        }
    }

    ahData, err := c.client.GetAuctionData(realmID)
    if err != nil {
        return nil, errors.Wrap(err, "cached client GetAuctionData")
    }

    err = c.cache.PSetStruct(cacheKey, ahData)
    if err != nil {
        return nil, errors.Wrap(err, "cached client GetAuctionData cache set")
    }

    return ahData, nil
}

func itemIdCacheKey(itemId int, region string) string {
    return fmt.Sprintf("ItemId:%d:%s", itemId, region)
}

func (c *cachedClient) getItemFromCache(id int) (*client.Item, error) {
    var res client.Item
    err := c.cache.PGetStruct(itemIdCacheKey(id, c.region), &res)
    if err == nil {
        return &res, nil
    }

    return nil, errors.Wrap(err, "item not found in cache")
}

func getItemFromCache(cache client.SearchItemResult, id int) *client.SearchItemResultData {
    for _, item := range cache.Results {
        if item.Data.Id == id {
            return &item
        }
    }

    return nil
}

func (c *cachedClient) SearchItemsByIds(ids []int) (*client.SearchItemResult, error) {
    // get whole cahce
    // for each id
    // check if this id is in cache
    // if not, add to "to request list"
    // if it's in the cache, add it to result list

    // after iterating check if request list is not empty
    // request items from api
    // add each item from api response to result list
    // cache result list
    // return result

    res := make([]client.SearchItemResultData, 0)
    notCachedIds := make([]int, 0)

    cacheKey := "SearchItemsByIds"
    var cache client.SearchItemResult
    err := c.cache.PGetStruct(cacheKey, &cache)
    if err == nil {
        for _, id := range ids {
            cachedItem := getItemFromCache(cache, id)
            if cachedItem != nil {
                res = append(res, *cachedItem)
            } else {
                notCachedIds = append(notCachedIds, id)
            }
        }
    } else {
        notCachedIds = ids
    }

    if len(notCachedIds) > 0 {
        apiResponse, err := c.client.SearchItemsByIds(notCachedIds)
        if err != nil {
            return nil, errors.Wrap(err, "cached client SearchItemsByIds")
        }

        for _, apiItem := range apiResponse.Results {
            res = append(res, client.SearchItemResultData{Data: apiItem.Data})
        }
    }

    err = c.cache.PSetStruct(cacheKey, client.SearchItemResult{Results: res})
    if err != nil {
        return nil, errors.Wrap(err, "cached client SearchItemsByIds cache set")
    }

    return &client.SearchItemResult{Results: res}, nil
}

func (c *cachedClient) GetBonuses() (*client.Bonuses, error) {
    cacheKey := "GetBonuses"

    var cached client.Bonuses
    err := c.cache.PGetStruct(cacheKey, &cached)
    if err == nil {
        c.log.Infof("loaded bonuses from cache")
        return &cached, nil
    }

    res, err := c.client.GetBonuses()
    if err != nil {
        return nil, errors.Wrap(err, "cached client GetBonuses")
    }

    err = c.cache.PSetStruct(cacheKey, res)
    if err != nil {
        return nil, errors.Wrap(err, "cached client GetBonuses cache set")
    }

    return res, nil
}
