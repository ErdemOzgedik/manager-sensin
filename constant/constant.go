package constant

import "go.mongodb.org/mongo-driver/bson"

const (
	DB                = "futManagerDB"
	PLAYERS           = "fut22Collection"
	MANAGERS          = "fut22Managers"
	SEASONS           = "fut22Seasons"
	RESULTS           = "fut22Results"
	TOPPLAYERS        = "topPlayers"
	ALLPLAYERS        = "-----[]-[]-[]"
	ALLPLAYERLIMIT    = 3000
	RANDOMPLAYERLIMIT = 68
)

// error messages
const (
	CHECKREDISERROR   = "Check redis error"
	GETREDISERROR     = "Get redis error"
	SETREDISERROR     = "Set redis error"
	SEARCHPLAYERERROR = "Search player error"
	DECODEERROR       = "Decode error"
	GETMANAGERERROR   = "Get manager error"
	GETPLAYERERROR    = "Get player error"
	GETSEASONERROR    = "Get season error"
	UPDATEERROR       = "Update error"
)

var OVERALLOPTION = bson.D{{Key: "overall", Value: -1}}

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
