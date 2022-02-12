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
type StatisticRequest struct {
	Season string `json:"season,omitempty"`
}
type StatisticResponse struct {
	Standing []structs.Standing `json:"standing"`
	Stats    []structs.Stats    `json:"stats"`
}
type SeasonRequest struct {
	ID       string `json:"id,omitempty"`
	IsActive bool   `json:"isActive,omitempty"`
}
type SeasonResponse struct {
	ID       string `json:"id,omitempty" bson:"_id,omitempty"`
	Title    string `json:"title,omitempty"`
	Type     string `json:"type,omitempty"`
	IsActive bool   `json:"isActive"`
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
type Pack struct {
	Type    int    `json:"type,omitempty"`
	Manager string `json:"manager,omitempty"`
}
type PackResponse struct {
	Point   int            `json:"point,omitempty"`
	Player  structs.Player `json:"player,omitempty"`
	Message string         `json:"message,omitempty"`
}
type Response struct {
	Count   int              `json:"count,omitempty"`
	Players []structs.Player `json:"players,omitempty"`
}
type ErrorResponse struct {
	Code    int    `json:"code,omitempty"`
	Message string `json:"message,omitempty"`
	Error   string `json:"error,omitempty"`
}
