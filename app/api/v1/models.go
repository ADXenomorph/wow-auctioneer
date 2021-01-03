package v1

import (
	"auctioneer/app/blizz"
)

type searchQueryParams struct {
	RealmName string `query:"realm_name,required"`
	ItemName  string `query:"item_name,required"`
	Region    string `query:"region,required"`
}

type ResponseV1 struct {
	Success bool                    `json:"success"`
	Message string                  `json:"message,omitempty"`
	Result  []*blizz.AuctionsDetail `json:"result"`
}
