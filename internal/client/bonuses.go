package client

import (
    "github.com/levigross/grequests"
    "github.com/pkg/errors"
)

type Bonuses struct {
    LevelBonuses map[int]int
}

type LevelBonus struct {
    ID    int `json:"id"`
    Level int `json:"level"`
}

func (c *client) GetBonuses() (*Bonuses, error) {
    requestURL := "https://www.raidbots.com/static/data/live/bonuses.json"

    response, err := c.makeGetRequest(requestURL, &grequests.RequestOptions{})
    if err != nil {
        return nil, errors.Wrapf(err, "GetBonuses makeGetRequest")
    }

    var result map[string]LevelBonus
    if err := response.JSON(result); err != nil {
        return nil, errors.Wrapf(err, "getBlizzRealms JSON")
    }

    bonuses := make(map[int]int, 0)
    for _, val := range result {
        if val.Level != 0 {
            bonuses[val.ID] = val.Level
        }
    }

    return &Bonuses{LevelBonuses: bonuses}, nil
}

//func NewBonuses(path string) *Bonuses {
//    jsonStr, err := ioutil.ReadFile(path)
//    if err != nil {
//        panic(err)
//    }
//
//    var result map[string]LevelBonus
//    err = json.Unmarshal(jsonStr, &result)
//    if err != nil {
//        panic(err)
//    }
//
//    bonuses := make(map[int]int, 0)
//    for _, val := range result {
//        if val.Level != 0 {
//            bonuses[val.ID] = val.Level
//        }
//    }
//
//    return &Bonuses{
//        LevelBonuses: bonuses,
//    }
//}

func (b *Bonuses) FindIlvlBonus(ids []int) int {
    for _, id := range ids {
        if level, ok := b.LevelBonuses[id]; ok {
            return level
        }
    }

    return 0
}
