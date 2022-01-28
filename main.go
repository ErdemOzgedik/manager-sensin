package main

import (
	"encoding/json"
	"fmt"
	"log"
	"manager-sensin/constant"
	"manager-sensin/helper"
	"manager-sensin/request"
	"manager-sensin/structs"
	"math/rand"
	"net/http"
	"time"

	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	"go.mongodb.org/mongo-driver/bson"
)

// player-start
func home(w http.ResponseWriter, r *http.Request) {
	var players []structs.Player
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
		filter := helper.AddFilterViaFields(&request.Filter{
			Overall: []int{87, 99},
		})
		players, err = helper.SearchPlayerByFilter(filter, constant.OVERALLOPTION, 0)
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
	json.NewEncoder(w).Encode(request.Response{
		Count:   len(players),
		Players: players,
	})
}
func searchPlayer(w http.ResponseWriter, r *http.Request) {
	var f request.Filter
	limit := 0

	err := json.NewDecoder(r.Body).Decode(&f)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if helper.GenerateRedisKey(&f) == constant.ALLPLAYERS {
		limit = constant.ALLPLAYERLIMIT
	}

	filter := helper.AddFilterViaFields(&f)

	players, err := helper.SearchPlayerByFilter(filter, constant.OVERALLOPTION, int64(limit))
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(request.Response{
		Count:   len(players),
		Players: players,
	})
}
func randomPlayer(w http.ResponseWriter, r *http.Request) {
	var f request.Filter
	var players []structs.Player
	var player structs.Player
	limit := 0

	err := json.NewDecoder(r.Body).Decode(&f)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	key := helper.GenerateRedisKey(&f)
	if key == constant.ALLPLAYERS {
		limit = constant.ALLPLAYERLIMIT
	}

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
		players, err = helper.SearchPlayerByFilter(filter, constant.OVERALLOPTION, int64(limit))
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
	}

	random := helper.GetRandom(0, len(players))
	player = players[random]
	players = append(players[:random], players[random+1:]...)

	err = helper.SetRedisData(pool, key, players, 45*time.Minute)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(request.Response{
		Count:   len(players),
		Players: []structs.Player{player},
	})
}

// player-end
// manager-start
func createManager(w http.ResponseWriter, r *http.Request) {
	var m structs.Manager
	err := json.NewDecoder(r.Body).Decode(&m)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	insert, err := helper.CreateManager(m)
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
func getManagers(w http.ResponseWriter, r *http.Request) {
	managers, err := helper.GetManagers()
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(managers)
}
func managePlayers(w http.ResponseWriter, r *http.Request) {
	var mp request.ManagePlayer

	err := json.NewDecoder(r.Body).Decode(&mp)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	manager, err := helper.GetManagerByID(mp.Manager)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	player, err := helper.GetPlayerByID(mp.Player)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if mp.Type == 0 {
		manager.DeletePlayer(player)
	} else {
		manager.AddPlayer(player)
	}

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
func managePoints(w http.ResponseWriter, r *http.Request) {
	var mp request.ManagePoint

	err := json.NewDecoder(r.Body).Decode(&mp)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	manager, err := helper.GetManagerByID(mp.Manager)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if mp.Type == 0 && manager.Points < mp.Point {
		http.Error(w, "Check your balance to precess", http.StatusBadRequest)
		return
	}

	manager.ManagePoint(mp.Point, mp.Type)

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
	var s structs.Season
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

// result-start-end
func resultLogic(w http.ResponseWriter, r *http.Request) {
	var resultRequest request.ResultRequest
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
	homeScorers := &[]structs.Scorer{}
	awayScorers := &[]structs.Scorer{}
	for _, scorer := range resultRequest.Scorers {
		if scorer.Manager == homeManager.ID.Hex() {
			for _, player := range homeManager.Players {
				if scorer.Player == player.ID.Hex() {
					*homeScorers = append(*homeScorers, structs.Scorer{
						Player: player,
						Count:  scorer.Count,
					})
					break
				}
			}
		} else {
			for _, player := range awayManager.Players {
				if scorer.Player == player.ID.Hex() {
					*awayScorers = append(*awayScorers, structs.Scorer{
						Player: player,
						Count:  scorer.Count,
					})
					break
				}
			}
		}
	}

	result := structs.Result{
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

// pack-start-end
func packOpener(w http.ResponseWriter, r *http.Request) {
	var pack request.Pack
	err := json.NewDecoder(r.Body).Decode(&pack)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	packPrice := constant.PACK_PRICES[pack.Type]
	manager, err := helper.GetManagerByID(pack.Manager)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if manager.Points < packPrice {
		http.Error(w, "Check manager point to process this action", http.StatusBadRequest)
		return
	} else {
		manager.ManagePoint(packPrice, 0)
	}

	filter, limit := helper.AddFilterViaType(pack.Type)
	players, err := helper.SearchPlayerByFilter(filter, bson.D{}, int64(limit))
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	rand.Seed(time.Now().UnixNano())
	random := helper.GetRandom(0, len(players))
	player := players[random]

	manager.AddPlayer(player)

	updateResult, err := helper.UpdateManager(&manager)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	fmt.Println("Manager players updated:", updateResult.ModifiedCount)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(request.PackResponse{
		Player:  player,
		Point:   manager.Points,
		Message: fmt.Sprintf("%d Oyuncu arasindan %d. gelen oyuncu %s", len(players), random, player.Name),
	})
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

	router.HandleFunc("/player", home)
	router.HandleFunc("/player/search", searchPlayer).Methods("POST", "OPTIONS")
	router.HandleFunc("/player/random", randomPlayer).Methods("POST", "OPTIONS")

	//manager endpoints
	router.HandleFunc("/manager", createManager).Methods("POST", "OPTIONS")
	router.HandleFunc("/manager", getManagers).Methods("GET", "OPTIONS")
	router.HandleFunc("/manager/player", managePlayers).Methods("POST", "OPTIONS")
	router.HandleFunc("/manager/point", managePoints).Methods("POST", "OPTIONS")

	//redis endpoints
	router.HandleFunc("/cleanRedis", cleanRedis).Methods("GET")

	//season endpoints
	router.HandleFunc("/season", createSeason).Methods("POST", "OPTIONS")

	//result endpoint
	router.HandleFunc("/result", resultLogic).Methods("POST", "OPTIONS")

	//pack endpoins
	router.HandleFunc("/pack", packOpener).Methods("POST", "OPTIONS")

	log.Fatal(http.ListenAndServe(":"+port, handlers.CORS(header, methods, origins)(router)))
}
