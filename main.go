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

	manager, err := helper.GetManagerByID(insert.InsertedID.Hex())
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

	manager, err := helper.GetManagerByID(pt.Manager)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	player, err := helper.GetPlayerByID(pt.Player)
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

	playerID, err := primitive.ObjectIDFromHex(pt.Player)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	manager, err := helper.GetManagerByID(pt.Manager)
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

	season, err := helper.GetSeasonByID(insert.InsertedID.Hex())
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
	var resultRequest constant.ResultRequest
	err := json.NewDecoder(r.Body).Decode(&resultRequest)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	//getByID interface ver parametre refactor taski
	//Concurrency her get icin 3 defa dbyi bekliyoz
	homeManager, err := helper.GetManagerByID(resultRequest.Home)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	awayManager, err := helper.GetManagerByID(resultRequest.Away)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	season, err := helper.GetSeasonByID(resultRequest.Season)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	//method yap refactor taski
	homeScorers := &[]constant.Scorer{}
	awayScorers := &[]constant.Scorer{}
	for _, scorer := range resultRequest.Scorers {
		if scorer.Manager == homeManager.ID.Hex() {
			for _, player := range homeManager.Players {
				if scorer.Player == player.ID.Hex() {
					*homeScorers = append(*homeScorers, constant.Scorer{
						Player: player,
						Count:  scorer.Count,
					})
					break
				}
			}
		} else {
			for _, player := range awayManager.Players {
				if scorer.Player == player.ID.Hex() {
					*awayScorers = append(*awayScorers, constant.Scorer{
						Player: player,
						Count:  scorer.Count,
					})
					break
				}
			}
		}
	}

	result := constant.Result{
		Season:      resultRequest.Season,
		Home:        resultRequest.Home,
		Away:        resultRequest.Away,
		SeasonType:  season.Type,
		SeasonTitle: season.Title,
		HomeManager: homeManager.Name,
		AwayManager: awayManager.Name,
		Score:       resultRequest.Score,
		HomeScorers: *homeScorers,
		AwayScorers: *awayScorers,
	}

	insert, err := helper.CreateResult(result)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	result.ID = insert.InsertedID

	homeManager.AddResult(result)
	_, err = helper.UpdateManager(&homeManager)
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

	season.AddResult(result)
	_, err = helper.UpdateSeason(&season)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(result)
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
