package internal

import (
    "bytes"
    "encoding/json"
    "fmt"
    "io/ioutil"
    "net/http"
    "time"

    "github.com/pkg/errors"
    "github.com/sirupsen/logrus"

    "github.com/ADXenomorph/wow-auctioneer/internal/client"
)

type App struct {
    bclient client.Client
    cfg     *Config
    logger  *logrus.Logger
}

func NewApp(bclient client.Client, cfg *Config, logger *logrus.Logger) *App {
    return &App{
        bclient: bclient,
        cfg:     cfg,
        logger:  logger,
    }
}

func (app *App) Setup() error {
    token, err := app.bclient.MakeBlizzAuth()
    if err != nil {
        return errors.Wrap(err, "blizzClient.MakeBlizzAuth")
    }

    app.bclient.SetToken(token)

    return nil
}

func (app *App) GetRealmId(realmName string) (int, error) {
    return app.bclient.GetRealmID(realmName)
}

func (app *App) GetAuctions(realmName string) (*client.AuctionData, error) {
    defer app.timeTrack(time.Now(), "GetAuctions")

    realmId, err := app.GetRealmId(realmName)

    if err != nil {
        err = errors.Wrap(err, "app GetAuctionData")
        app.logger.Error(err)
        return nil, err
    }

    return app.bclient.GetAuctionData(realmId)
}

func (app *App) SearchItems(name string) error {
    _, err := app.bclient.SearchItem(name)
    return err
}

func (app *App) DecorateAuctionData(data *client.AuctionData) (*DecoratedAuctionData, error) {
    defer app.timeTrack(time.Now(), "DecorateAuctionData")

    res := make([]*DecAucItem, 0)

    itemIds := make([]int, 0)
    for _, auc := range data.Auctions {
        itemIds = append(itemIds, auc.Item.ID)
    }

    items, err := app.bclient.SearchItemsByIds(itemIds)
    if err != nil {
        err = errors.Wrap(err, "app DecorateAuctionData")
        app.logger.Error(err)
        return nil, err
    }

    bonuses, err := app.bclient.GetBonuses()
    if err != nil {
        err = errors.Wrap(err, "app DecorateAuctionData GetBonuses")
        app.logger.Error(err)
        return nil, err
    }

    for _, auc := range data.Auctions {
        aucItem, _ := items.FindById(auc.Item.ID)
        if aucItem == nil {
            continue
        }

        res = append(res, &DecAucItem{
            AuctionsDetail: *auc,
            Name:           aucItem.Name.EnUS,
            Ilvl:           aucItem.Level + bonuses.FindIlvlBonus(auc.Item.BonusLists),
        })
    }

    return &DecoratedAuctionData{Items: res}, nil
}

func (app *App) FindBOEOutliers(data *DecoratedAuctionData) *DecoratedAuctionData {
    res := make([]*DecAucItem, 0)

    groups := data.GroupItemsByNameAndIlvl()

    for _, group := range groups {
        outliers := group.FindOutliers()
        for _, outlier := range outliers {
            res = append(res, outlier)
        }
    }

    return &DecoratedAuctionData{Items: res}
}

func (app *App) timeTrack(start time.Time, name string) {
    elapsed := time.Since(start)
    app.logger.Infof("%s took %s", name, elapsed)
}

func (app *App) SendMessage(text string) error {
    var err error
    var response *http.Response

    url := fmt.Sprintf("https://api.telegram.org/bot%s/sendMessage", app.cfg.TelegramToken)
    body, _ := json.Marshal(map[string]string{
        "chat_id": app.cfg.TelegramChatId,
        "text":    text,
    })
    response, err = http.Post(
        url,
        "application/json",
        bytes.NewBuffer(body),
    )
    if err != nil {
        return err
    }

    defer response.Body.Close()

    body, err = ioutil.ReadAll(response.Body)
    if err != nil {
        return err
    }

    app.logger.Infof("Telegram message sent")

    return nil
}

func (app *App) ScanForOutliers(server string) (*DecoratedAuctionData, error) {
    auctions, err := app.GetAuctions(server)
    if err != nil {
        return nil, errors.Wrap(err, "app.GetAuctions")
    }

    auctions = auctions.FilterByBuyout(30_000_00_00)
    decoratedAuctions, err := app.DecorateAuctionData(auctions)
    if err != nil {
        return nil, errors.Wrap(err, "app.DecorateAuctionData")
    }
    // decoratedAuctions = decoratedAuctions.FilterByName("Devoted Warden")
    decoratedAuctions = decoratedAuctions.FilterByIlvl(400, 450)
    decoratedAuctions = app.FindBOEOutliers(decoratedAuctions)

    return decoratedAuctions, nil
}
