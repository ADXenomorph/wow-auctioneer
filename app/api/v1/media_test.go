package v1_test

import (
	"fmt"
	"io/ioutil"
	"testing"

	"auctioneer/app/api/v1"
	server "auctioneer/app/auctioneer"
	"auctioneer/app/blizz"
	"auctioneer/app/conf"
	logging "auctioneer/app/logger"
	"encoding/json"
	"github.com/stretchr/testify/assert"
	"net/http/httptest"
)

func Test_SearchItemMedia(t *testing.T) {
	cfg, _ := conf.NewConfig()
	logger, _ := logging.NewLogger("DEBUG")
	app := server.NewApp(logger, cfg)
	app.BaseHandler = &mockHandler{
		v1:     newV1handler(),
		system: newSystemHandler(),
	}
	app.SetupRoutes()

	testCases := []struct {
		name      string
		itemID    int
		expStatus int
		exp       interface{}
	}{
		{
			name:      "OK request",
			itemID:    200,
			expStatus: 200,
			exp: v1.ResponseV1ItemMedia{
				Success: true,
				ItemMedia: &blizz.ItemMedia{
					Assets: []blizz.ItemAssets{
						blizz.ItemAssets{
							Key:        "hello",
							Value:      "world",
							FileDataID: 100,
						},
					},
					ID: 200,
				},
			},
		},
		{
			name:      "404 request",
			itemID:    404,
			expStatus: 404,
			exp: v1.ResponseV1{
				Success: false,
				Message: "Item with ID 404 not found",
			},
		},
		{
			name:      "400 request",
			itemID:    400,
			expStatus: 400,
			exp: v1.ResponseV1{
				Success: false,
				Message: "error making get item mdeia request, status: 400",
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			reqURI := fmt.Sprintf("/api/v1/item_media/%d", tc.itemID)
			req := httptest.NewRequest("GET", reqURI, nil)

			resp, err := app.Fib.Test(req)
			assert.NoError(t, err)
			assert.Equal(t, tc.expStatus, resp.StatusCode)

			bodyBytes, err := ioutil.ReadAll(resp.Body)
			assert.NoError(t, err)

			if tc.expStatus == 200 {
				var respV1 v1.ResponseV1ItemMedia
				err = json.Unmarshal(bodyBytes, &respV1)
				assert.NoError(t, err)

				assert.EqualValues(t, tc.exp, respV1)
				return
			}

			var respV1 v1.ResponseV1
			err = json.Unmarshal(bodyBytes, &respV1)
			assert.NoError(t, err)

			assert.EqualValues(t, tc.exp, respV1)
		})
	}
}
