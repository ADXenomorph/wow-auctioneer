package client

type BlizzApiCfg struct {
    RegionList   []string
    EuAPIUrl     string `envconfig:"BLIZZARD_EU_API_URL" default:"https://eu.api.blizzard.com"`
    UsAPIUrl     string `envconfig:"BLIZZARD_US_API_URL" default:"https://us.api.blizzard.com"`
    AUTHUrl      string `envconfig:"BLIZZARD_AUTH_URL" default:"https://us.battle.net/oauth/token"`
    ClientSecret string `envconfig:"BLIZZARD_CLIENT_SECRET" required:"true"`
    ClientID     string `envconfig:"BLIZZARD_CLIENT_ID" required:"true"`
    AuthTimeOut  int    `envconfig:"BLIZZARD_AUTH_TIMEOUT" default:"3"`
}
