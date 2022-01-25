package request

import (
	"manager-sensin/structs"
)

type ResultRequest struct {
	Season  string          `json:"season,omitempty"`
	Home    string          `json:"home,omitempty"`
	Away    string          `json:"away,omitempty"`
	Score   []int           `json:"score,omitempty"`
	Scorers []ScorerRequest `json:"scorer,omitempty"`
}
type ScorerRequest struct {
	Player  string `json:"player,omitempty"`
	Manager string `json:"manager,omitempty"`
	Count   int    `json:"count,omitempty"`
}
type ManagePlayer struct {
	Manager string `json:"manager,omitempty"`
	Player  string `json:"player,omitempty"`
	Type    int    `json:"type,omitempty"`
}
type ManagePoint struct {
	Manager string `json:"manager,omitempty"`
	Point   int    `json:"point,omitempty"`
	Type    int    `json:"type,omitempty"`
}
type Filter struct {
	Name        string `json:"name,omitempty"`
	Club        string `json:"club,omitempty"`
	League      string `json:"league,omitempty"`
	Nationality string `json:"nationality,omitempty"`
	Age         []int  `json:"age,omitempty"`
	Overall     []int  `json:"overall,omitempty"`
	Potential   []int  `json:"potential,omitempty"`
	Position    string `json:"position,omitempty"`
}

type Response struct {
	Count   int              `json:"count,omitempty"`
	Players []structs.Player `json:"players,omitempty"`
}
