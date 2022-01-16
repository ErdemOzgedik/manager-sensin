package constant

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// User is a struct
type Player struct {
	ID           primitive.ObjectID `bson:"_id,omitempty"`
	LongName     string             `bson:"long_name,omitempty"`
	Name         string             `bson:"short_name,omitempty"`
	Positions    string             `bson:"player_positions,omitempty"`
	ClubPosition string             `bson:"club_position,omitempty"`
	Club         string             `bson:"club_name,omitempty"`
	League       string             `bson:"league_name,omitempty"`
	Nationality  string             `bson:"nationality_name,omitempty"`
	Age          int                `bson:"age,omitempty"`
	Overall      int                `bson:"overall,omitempty"`
	Potential    int                `bson:"potential,omitempty"`
	// Pace        int                `bson:"pace,omitempty"`
	// Passing     int                `bson:"passing,omitempty"`
	// Physic      int                `bson:"physic,omitempty"`
	// Shooting    int                `bson:"shooting,omitempty"`
	// Dribbling   int                `bson:"dribbling,omitempty"`
	// Defending   int                `bson:"defending,omitempty"`
	FaceUrl    string `bson:"player_face_url,omitempty"`
	ClubLogo   string `bson:"club_logo_url,omitempty"`
	NationFlag string `bson:"nation_flag_url,omitempty"`
	WF         int    `bson:"weak_foot,omitempty"`
	SM         int    `bson:"skill_moves,omitempty"`
	WorkRate   string `bson:"work_rate,omitempty"`
	Foot       string `bson:"preferred_foot,omitempty"`
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
	Count   int      `json:"count,omitempty"`
	Players []Player `json:"players,omitempty"`
}
