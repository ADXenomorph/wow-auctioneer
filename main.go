package main

import (
    "os"

    _ "github.com/joho/godotenv/autoload"

    "github.com/ADXenomorph/wow-auctioneer/internal/cmd"
)

func main() {
    if err := cmd.ExecuteRootCmd(); err != nil {
        os.Exit(1)
    }
}
