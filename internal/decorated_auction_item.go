package internal

import (
    "fmt"
    "sort"
    "strings"

    "github.com/montanaflynn/stats"

    "github.com/ADXenomorph/wow-auctioneer/internal/client"
)

type DecAucItem struct {
    client.AuctionsDetail
    Name                   string  `json:"name"`
    Ilvl                   int     `json:"ilvl"`
    PriceDiff              int     `json:"price_diff"`
    PriceDiffPercent       float64 `json:"price_diff_percent"`
    ClosestPriceItemIlvl   int     `json:"closest_item_ilvl"`
    ClosestPriceItemBuyout int     `json:"closest_item_buyout"`
}

type DecoratedAuctionData struct {
    Items []*DecAucItem `json:"list"`
}

type ItemGroup []*DecAucItem

// DecAucItem methods

func (dai *DecAucItem) String() string {
    return fmt.Sprintf(
        "%s - %d ilvl - %dg, diff: %dg, diff%%: %.2f, closest item: %d ilvl - %dg",
        dai.Name,
        dai.Ilvl,
        dai.Buyout/10000,
        dai.PriceDiff/10000,
        dai.PriceDiffPercent,
        dai.ClosestPriceItemIlvl,
        dai.ClosestPriceItemBuyout/10000,
    )
}

// DecoratedAuctionData methods

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

func (dad *DecoratedAuctionData) String() string {
    msgs := make([]string, 0)
    for _, auc := range dad.Items {
        msgs = append(msgs, auc.String())
    }
    return strings.Join(msgs, "\n")
}

func (dad *DecoratedAuctionData) GroupItems(hashFn func(item *DecAucItem) string) map[string]ItemGroup {
    groups := make(map[string]ItemGroup, 0)

    for _, item := range dad.Items {
        itemHash := hashFn(item)
        if _, ok := groups[itemHash]; !ok {
            groups[itemHash] = make([]*DecAucItem, 0)
        }

        groups[itemHash] = append(groups[itemHash], item)
    }

    return groups
}

func (dad *DecoratedAuctionData) GroupItemsByNameAndIlvl() map[string]ItemGroup {
    return dad.GroupItems(func(item *DecAucItem) string {
        return fmt.Sprintf("%s-%d", item.Name, item.Ilvl)
    })
}

func (dad *DecoratedAuctionData) GroupItemsByName() map[string]ItemGroup {
    return dad.GroupItems(func(item *DecAucItem) string {
        return item.Name
    })
}

// Item group methods

// Sorting methods

func (g ItemGroup) Len() int {
    return len(g)
}
func (g ItemGroup) Less(i, j int) bool {
    return g[i].Buyout < g[j].Buyout
}
func (g ItemGroup) Swap(i, j int) {
    g[i], g[j] = g[j], g[i]
}

func (g ItemGroup) filterByIlvl(ilvl int) ItemGroup {
    res := make([]*DecAucItem, 0)
    for _, item := range g {
        if item.Ilvl == ilvl {
            res = append(res, item)
        }
    }

    return res
}

func (g ItemGroup) getIlvls() []int {
    unique := make(map[int]bool, 0)
    for _, item := range g {
        unique[item.Ilvl] = true
    }

    res := make([]int, len(unique))
    for k := range unique {
        res = append(res, k)
    }

    sort.Ints(res)
    return res
}

func (g ItemGroup) getOutlierBorder() float64 {
    prices := make([]float64, 0)
    for _, item := range g {
        prices = append(prices, float64(item.Buyout))
    }

    quartiles, _ := stats.Quartile(prices)
    iqr, _ := stats.InterQuartileRange(prices)

    return quartiles.Q1 - 1.5*iqr
}

func (g ItemGroup) findNextPriceHigherThan(price int) int {
    sort.Sort(g)
    nextPrice := 0
    for _, item := range g {
        if item.Buyout > price && (nextPrice == 0 || item.Buyout < nextPrice) {
            nextPrice = item.Buyout
        }
    }

    return nextPrice
}

func (g ItemGroup) getMinPrice() int {
    minPrice := 10_000_000_00_00 // gold cap
    for _, item := range g {
        if item.Price < minPrice {
            minPrice = item.Price
        }
    }

    return minPrice
}

type outlierResult struct {
    ilvl     int
    minPrice int
    outlier  *DecAucItem
}

func (g ItemGroup) FindOutliers() []*DecAucItem {
    outlierResults := make([]outlierResult, 0)

    ilvls := g.getIlvls()
    for _, ilvl := range ilvls {
        ilvlG := g.filterByIlvl(ilvl)
        outlier := ilvlG.findOutlierInIlvlGroup()
        outlierResults = append(outlierResults, outlierResult{
            ilvl:     ilvl,
            minPrice: g.getMinPrice(),
            outlier:  outlier,
        })
    }

    res := make([]*DecAucItem, 0)
    for _, outlierResult := range outlierResults {
        if outlierResult.outlier == nil {
            continue
        }

        closestOtherIlvlPrice := 0
        for _, innerOutlierResult := range outlierResults {
            if closestOtherIlvlPrice == 0 || innerOutlierResult.minPrice < closestOtherIlvlPrice {
                outlierResult.outlier.ClosestPriceItemIlvl = innerOutlierResult.ilvl
                outlierResult.outlier.ClosestPriceItemBuyout = innerOutlierResult.minPrice
                closestOtherIlvlPrice = innerOutlierResult.minPrice
            }
        }
        res = append(res, outlierResult.outlier)
    }

    return res
}

func (g ItemGroup) findOutlierInIlvlGroup() *DecAucItem {
    sort.Sort(g)
    outlierBorder := g.getOutlierBorder()

    outliers := make([]*DecAucItem, 0)
    for _, item := range g {
        if float64(item.Buyout) < outlierBorder {
            outliers = append(outliers, item)
        }
    }

    if len(outliers) != 1 {
        return nil
    }

    res := outliers[0]
    nextPrice := g.findNextPriceHigherThan(res.Buyout)
    if nextPrice != 0 {
        res.PriceDiff = nextPrice - res.Buyout
        res.PriceDiffPercent = 100 - (float64(res.Buyout) * 100 / float64(nextPrice))
    }

    if res.PriceDiffPercent < 5 {
        return nil
    }

    return res
}
