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

	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// player-start
func home(w http.ResponseWriter, r *http.Request) {
	var players []constant.Player
	pool := helper.GetRedisPool()

	exists, err := helper.CheckRedisData(pool, constant.TOPPLAYERS)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if exists {
		err := helper.GetRedisData(pool, constant.TOPPLAYERS, &players)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
	} else {
		filter := helper.AddFilterViaFields(&constant.Filter{
			Overall: []int{87, 99},
		})
		players, err = helper.SearchPlayerByFilter(filter)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		err = helper.SetRedisData(pool, constant.TOPPLAYERS, players, 0)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
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
	w.WriteHeader(http.StatusOK)
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
	pool := helper.GetRedisPool()
	exists, err := helper.CheckRedisData(pool, key)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if exists {
		err = helper.GetRedisData(pool, key, &players)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
	} else {
		filter := helper.AddFilterViaFields(&f)
		players, err = helper.SearchPlayerByFilter(filter)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
	}

	if len(players) > constant.RANDOMPLAYERLIMIT {
		rand.Seed(time.Now().UnixNano())
		min := 0
		max := len(players)
		random := rand.Intn(max-min) + min
		player = players[random]
		players = append(players[:random], players[random+1:]...)

		err = helper.SetRedisData(pool, key, players, 45*time.Minute)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
	} else {
		http.Error(w, "Change filter to use random api", http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(constant.Response{
		Count:   len(players),
		Players: []constant.Player{player},
	})
}

// player-end
// manager-start
func createManager(w http.ResponseWriter, r *http.Request) {
	var t constant.Manager
	err := json.NewDecoder(r.Body).Decode(&t)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	insert, err := helper.CreateManager(t)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	manager, err := helper.GetManagerByID(insert.InsertedID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(manager)
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

// manager-end
// redis-start-end
func cleanRedis(w http.ResponseWriter, r *http.Request) {
	err := helper.DeteleRedisKeys()
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	http.Error(w, "Redis data removed!!!", http.StatusOK)
}

// season-start-en
func createSeason(w http.ResponseWriter, r *http.Request) {
	var s constant.Season
	err := json.NewDecoder(r.Body).Decode(&s)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	insert, err := helper.CreateSeason(s)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	season, err := helper.GetSeasonByID(insert.InsertedID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(season)
}

// result-start-en
func resultLogic(w http.ResponseWriter, r *http.Request) {
	var result constant.Result
	err := json.NewDecoder(r.Body).Decode(&result)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	insert, err := helper.CreateResult(result)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	result.ID = insert.InsertedID

	//managerlerin result update et duzelt bunu kotu kod
	homeID, err := primitive.ObjectIDFromHex(result.Home)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	homeManager, err := helper.GetManagerByID(homeID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	homeManager.AddResult(result)
	_, err = helper.UpdateManager(&homeManager)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	awayID, err := primitive.ObjectIDFromHex(result.Away)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	awayManager, err := helper.GetManagerByID(awayID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	awayManager.AddResult(result)
	_, err = helper.UpdateManager(&awayManager)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	//season result update et
	seasonID, err := primitive.ObjectIDFromHex(result.Season)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	season, err := helper.GetSeasonByID(seasonID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	season.AddResult(result)
	_, err = helper.UpdateSeason(&season)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

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

	//redis endpoints
	router.HandleFunc("/cleanRedis", cleanRedis).Methods("GET")

	//season endpoints
	router.HandleFunc("/createSeason", createSeason).Methods("POST", "OPTIONS")

	//result endpoint
	router.HandleFunc("/result", resultLogic).Methods("POST", "OPTIONS")

	log.Fatal(http.ListenAndServe(":"+port, handlers.CORS(header, methods, origins)(router)))
}
