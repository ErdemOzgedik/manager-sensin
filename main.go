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
	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func home(w http.ResponseWriter, r *http.Request) {
	var players []constant.Player
	rdb := helper.GetRedisClient()
	err := helper.GetRedisData(rdb, constant.TOPPLAYERS, &players)
	if err == redis.Nil {
		filter := helper.AddFilterViaFields(&constant.Filter{
			Overall: []int{87, 99},
		})
		players, err = helper.SearchPlayerByFilter(filter)
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

	players, err := helper.SearchPlayerByFilter(filter)
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
		players, err = helper.SearchPlayerByFilter(filter)
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

func createManager(w http.ResponseWriter, r *http.Request) {
	var t constant.Manager
	err := json.NewDecoder(r.Body).Decode(&t)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	helper.CreateManager(t)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	w.Write([]byte("Manager created!!!"))
}

func addPlayer(w http.ResponseWriter, r *http.Request) {
	var pt constant.PlayerTransfer

	err := json.NewDecoder(r.Body).Decode(&pt)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	managerID, err := primitive.ObjectIDFromHex(pt.Manager)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	playerID, err := primitive.ObjectIDFromHex(pt.Player)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	manager, err := helper.GetManagerByID(managerID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	player, err := helper.GetPlayerByID(playerID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	manager.AddPlayer(player)

	updateResult, err := helper.UpdateManager(&manager)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	fmt.Println("Updated Count", updateResult.ModifiedCount)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(manager)
}

func deletePlayer(w http.ResponseWriter, r *http.Request) {
	var pt constant.PlayerTransfer

	err := json.NewDecoder(r.Body).Decode(&pt)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	managerID, err := primitive.ObjectIDFromHex(pt.Manager)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	playerID, err := primitive.ObjectIDFromHex(pt.Player)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	manager, err := helper.GetManagerByID(managerID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	manager.DeletePlayer(playerID)

	updateResult, err := helper.UpdateManager(&manager)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	fmt.Println("Updated Count", updateResult.ModifiedCount)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(manager)
}

func main() {
	var port string
	var err error

	port, err = helper.GetEnv("PORT")
	if err != nil {
		fmt.Println("Server will use default port")
		port = "8080"
	}

	router := mux.NewRouter().StrictSlash(true)
	header := handlers.AllowedHeaders([]string{"X-Requested-With", "Content-Type", "Authorization"})
	methods := handlers.AllowedMethods([]string{"GET", "POST", "PUT", "HEAD", "OPTIONS"})
	origins := handlers.AllowedOrigins([]string{"*"})

	router.HandleFunc("/", home)
	router.HandleFunc("/search", searchPlayer).Methods("POST", "OPTIONS")
	router.HandleFunc("/random", randomPlayer).Methods("POST", "OPTIONS")

	//manager endpoints
	router.HandleFunc("/createManager", createManager).Methods("POST", "OPTIONS")
	router.HandleFunc("/addPlayer", addPlayer).Methods("POST", "OPTIONS")
	router.HandleFunc("/deletePlayer", deletePlayer).Methods("POST", "OPTIONS")

	log.Fatal(http.ListenAndServe(":"+port, handlers.CORS(header, methods, origins)(router)))
}
