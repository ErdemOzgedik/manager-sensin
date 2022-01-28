package constant

import "go.mongodb.org/mongo-driver/bson"

const (
	DB                = "futManagerDB"
	PLAYERS           = "fut22Collection"
	MANAGERS          = "fut22Managers"
	SEASONS           = "fut22Seasons"
	RESULTS           = "fut22Results"
	TOPPLAYERS        = "topPlayers"
	RANDOMPLAYERLIMIT = 68
)

var OVERALLOPTIONS = bson.D{{Key: "overall", Value: -1}}

const (
	SILVER int = iota
	PREMIUMSILVER
	GOLD
	PREMIUMGOLD
	ULTIMATEGOLD
	PRIMEGOLD
)

var PACK_PRICES = map[int]int{
	SILVER:        500,
	PREMIUMSILVER: 750,
	GOLD:          2000,
	PREMIUMGOLD:   5000,
	ULTIMATEGOLD:  10000,
	PRIMEGOLD:     15000,
}
