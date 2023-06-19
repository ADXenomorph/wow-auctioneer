package cmd

import (
    "context"
    "fmt"

    "github.com/pkg/errors"
    "github.com/spf13/cobra"

    "github.com/ADXenomorph/wow-auctioneer/internal"
    "github.com/ADXenomorph/wow-auctioneer/internal/client"
    "github.com/ADXenomorph/wow-auctioneer/internal/pcache"
)

var (
    region       string
    server       string
    withTelegram bool
    cachePath    string

    scanCmd = &cobra.Command{
        Use:   "scan",
        Short: "Run scan",
        RunE: func(cmd *cobra.Command, args []string) error {
            if server == "" {
                return errors.New("server parameter cannot be empty")
            }

            ctx := context.Background()
            cfg, err := internal.NewConfig()
            if err != nil {
                return errors.Wrap(err, "config error")
            }

            logger, err := internal.NewLogger(cfg.LogLvl)
            if err != nil {
                return errors.Wrap(err, "logger error")
            }

            blizzClient := client.NewClient(ctx, logger, &cfg.BlizzApiCfg, region)
            blizzClient = internal.NewCachedClient(blizzClient, pcache.NewPCache(cachePath, 0, 5000), logger, region)

            app := internal.NewApp(blizzClient, cfg, logger)

            if app.Setup() != nil {
                return errors.Wrap(err, "app.Setup")
            }

            decoratedAuctions, err := app.ScanForOutliers(server)
            if err != nil {
                return errors.Wrap(err, "app.ScanForOutliers")
            }

            if len(decoratedAuctions.Items) == 0 {
                fmt.Println("No auctions found")
                return nil
            }

            msg := decoratedAuctions.String()
            fmt.Println(msg)

            if withTelegram {
                err = app.SendMessage(msg)
                if err != nil {
                    return errors.Wrap(err, "app.SendMessage")
                }
            }

            return nil
        },
    }
)

func init() {
    rootCmd.AddCommand(scanCmd)

    scanCmd.PersistentFlags().StringVar(&region, "region", "eu", "Server region. Can be 'eu' or 'us'")
    scanCmd.PersistentFlags().StringVar(&server, "server", "", "Server name. E.g 'Argent Dawn'")
    scanCmd.PersistentFlags().BoolVarP(&withTelegram, "telegram", "t", false, "Send results to Telegram")
    scanCmd.PersistentFlags().StringVar(&cachePath, "cachePath", "./.cache", "Path to cache dir. Default is './.cache'")
}
