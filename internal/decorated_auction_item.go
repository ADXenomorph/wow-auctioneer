package internal

import (
    "fmt"
    "strings"

    "github.com/montanaflynn/stats"

    "github.com/ADXenomorph/wow-auctioneer/internal/client"
)

type DecAucItem struct {
    client.AuctionsDetail
    Name string `json:"name"`
    Ilvl int    `json:"ilvl"`
}

type DecoratedAuctionData struct {
    Items []*DecAucItem `json:"list"`
}

type ItemGroup []*DecAucItem

func (dad *DecoratedAuctionData) filter(filterFunc func(auc *DecAucItem) bool) *DecoratedAuctionData {
    res := make([]*DecAucItem, 0)
    for _, auction := range dad.Items {
        if filterFunc(auction) {
            res = append(res, auction)
        }
    }

    return &DecoratedAuctionData{Items: res}
}

func (dad *DecoratedAuctionData) FilterByName(name string) *DecoratedAuctionData {
    return dad.filter(func(auc *DecAucItem) bool {
        return strings.Contains(strings.ToLower(auc.Name), strings.ToLower(name))
    })
}

func (dad *DecoratedAuctionData) FilterByItemId(id int) *DecoratedAuctionData {
    return dad.filter(func(auc *DecAucItem) bool {
        return auc.AuctionsDetail.Item.ID == id
    })
}

func (dad *DecoratedAuctionData) FilterByIlvl(from int, to int) *DecoratedAuctionData {
    return dad.filter(func(auc *DecAucItem) bool {
        return from <= auc.Ilvl && auc.Ilvl <= to
    })
}

func (dad *DecoratedAuctionData) ToString() string {
    msgs := make([]string, 0)
    for _, auc := range dad.Items {
        msgs = append(msgs, fmt.Sprintf("%s - %d ilvl - %dg", auc.Name, auc.Ilvl, auc.Buyout/10000))
    }
    return strings.Join(msgs, "\n")
}

func (dad *DecoratedAuctionData) GroupItemsByNameAndIlvl() map[string]ItemGroup {
    groups := make(map[string]ItemGroup, 0)

    for _, item := range dad.Items {
        itemHash := fmt.Sprintf("%s-%d", item.Name, item.Ilvl)
        if _, ok := groups[itemHash]; !ok {
            groups[itemHash] = make([]*DecAucItem, 0)
        }

        groups[itemHash] = append(groups[itemHash], item)
    }

    return groups
}

func (g ItemGroup) FindOutlier() *DecAucItem {
    prices := make([]float64, 0)
    for _, item := range g {
        prices = append(prices, float64(item.Buyout))
    }

    quartiles, _ := stats.Quartile(prices)
    iqr, _ := stats.InterQuartileRange(prices)

    outlierBorder := quartiles.Q1 - 1.5*iqr

    var res *DecAucItem
    for _, item := range g {
        if float64(item.Buyout) < outlierBorder && (res == nil || res.Buyout > item.Buyout) {
            res = item
            break
        }
    }

    return res
}
