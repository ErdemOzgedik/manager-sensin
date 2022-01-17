package main

import (
	"encoding/json"
	"fmt"
	"log"
	"manager-sensin/constant"
	"manager-sensin/helper"
	"math/rand"
	"net/http"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/gorilla/mux"
	"github.com/joho/godotenv"
)

func home(w http.ResponseWriter, r *http.Request) {
	var players []constant.Player
	rdb := helper.GetRedisClient()
	err := helper.GetRedisData(rdb, constant.TOPPLAYERS, &players)
	if err == redis.Nil {
		filter := helper.AddFilterViaFields(&constant.Filter{
			Overall: []int{87, 99},
		})
		players, err = helper.SearchByFilter(filter)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		helper.SetRedisData(rdb, constant.TOPPLAYERS, players, 0)
	} else if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(constant.Response{
		Count:   len(players),
		Players: players,
	})
}
func searchPlayer(w http.ResponseWriter, r *http.Request) {
	var f constant.Filter

	err := json.NewDecoder(r.Body).Decode(&f)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	filter := helper.AddFilterViaFields(&f)

	players, err := helper.SearchByFilter(filter)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(constant.Response{
		Count:   len(players),
		Players: players,
	})
}
func randomPlayer(w http.ResponseWriter, r *http.Request) {
	var f constant.Filter
	var players []constant.Player
	var player constant.Player

	err := json.NewDecoder(r.Body).Decode(&f)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	key := helper.GenerateRedisKey(&f)
	rdb := helper.GetRedisClient()
	err = helper.GetRedisData(rdb, key, &players)
	if err == redis.Nil {
		filter := helper.AddFilterViaFields(&f)
		players, err = helper.SearchByFilter(filter)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
	} else if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if len(players) > 0 {
		rand.Seed(time.Now().UnixNano())
		min := 0
		max := len(players)
		random := rand.Intn(max-min) + min
		player = players[random]
		players = append(players[:random], players[random+1:]...)
		helper.SetRedisData(rdb, key, players, 45*time.Minute)
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(constant.Response{
		Count:   len(players),
		Players: []constant.Player{player},
	})
}

func main() {
	var port string
	var err error

	err = godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	port, err = helper.GetEnv("PORT")
	if err != nil {
		fmt.Println("Server will use default port")
		port = "8080"
	}

	router := mux.NewRouter().StrictSlash(true)
	router.HandleFunc("/", home)
	router.HandleFunc("/search", searchPlayer).Methods("POST")
	router.HandleFunc("/random", randomPlayer).Methods("POST")

	log.Fatal(http.ListenAndServe(":"+port, router))
}
