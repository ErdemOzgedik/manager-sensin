package helper

import (
	"context"
	"encoding/json"
	"fmt"
	"manager-sensin/constant"
	"manager-sensin/request"
	"manager-sensin/structs"
	"math/rand"
	"net/http"
	"net/url"
	"os"
	"sort"
	"sync"
	"time"

	"github.com/gomodule/redigo/redis"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

/* Used to create a singleton object of MongoDB client.
Initialized and exposed through  GetMongoClient().*/
var clientInstance *mongo.Client

//Used during creation of singleton client object in GetMongoClient().
var clientInstanceError error

//Used to execute client creation procedure only once.
var mongoOnce sync.Once

//GetMongoClient - Return mongodb connection to work with
func GetMongoClient() (*mongo.Client, error) {
	//Perform connection creation operation only once.
	mongoOnce.Do(func() {
		// Set client options
		cs, err := GetEnv("connectionstring")
		if err != nil {
			clientInstanceError = err
		}
		clientOptions := options.Client().ApplyURI(cs)
		// Connect to MongoDB
		client, err := mongo.Connect(context.TODO(), clientOptions)
		if err != nil {
			clientInstanceError = err
		}
		// Check the connection
		err = client.Ping(context.TODO(), nil)
		if err != nil {
			clientInstanceError = err
		}
		clientInstance = client
	})

	return clientInstance, clientInstanceError
}

// redis-start
func GetRedisPool() *redis.Pool {
	var addr string
	var err error
	var pass string

	addr, err = GetEnv("REDISTOGO_URL")
	if err != nil {
		fmt.Println("Server will use default redis")
		addr = "localhost:6379"
	} else {
		u, err := url.Parse(addr)
		if err != nil {
			fmt.Println("url parse error")
		}
		addr = u.Host
		pass, _ = u.User.Password()
	}

	return &redis.Pool{
		MaxIdle:     10,
		IdleTimeout: 240 * time.Second,
		Dial: func() (redis.Conn, error) {
			return redis.Dial("tcp", addr, redis.DialPassword(pass))
		},
		TestOnBorrow: func(c redis.Conn, t time.Time) error {
			_, err := c.Do("PING")
			return err
		},
	}
}
func SetRedisData(pool *redis.Pool, key string, players []structs.Player, duration time.Duration) error {
	conn := pool.Get()
	defer conn.Close()

	value, err := json.Marshal(players)
	if err != nil {
		return err
	}

	if duration.Seconds() == 0 {
		err = conn.Send("SET", key, string(value))
		if err != nil {
			return err
		}
	} else {
		err = conn.Send("SET", key, string(value), "EX", duration.Seconds())
		if err != nil {
			return err
		}
	}

	return nil
}
func GetRedisData(pool *redis.Pool, key string, players *[]structs.Player) error {
	conn := pool.Get()
	defer conn.Close()

	data, err := redis.Bytes(conn.Do("GET", key))
	if err != nil {
		return fmt.Errorf("error getting key %s: %v", key, err)
	}

	return json.Unmarshal(data, &players)
}
func CheckRedisData(pool *redis.Pool, key string) (bool, error) {
	conn := pool.Get()
	defer conn.Close()

	ok, err := redis.Bool(conn.Do("EXISTS", key))
	if err != nil {
		return ok, fmt.Errorf("error checking if key %s exists: %v", key, err)
	}

	return ok, err
}
func GenerateRedisKey(filter *request.Filter) string {
	return fmt.Sprintf("%s-%s-%s-%s-%s-%v-%v-%v", filter.Name,
		filter.Club, filter.Nationality, filter.League,
		filter.Position, filter.Age, filter.Overall, filter.Potential)
}
func DeteleRedisKeys() error {
	pool := GetRedisPool()
	conn := pool.Get()
	defer conn.Close()

	keys, err := redis.Strings(conn.Do("KEYS", "*"))
	if err != nil {
		return err
	}

	for _, key := range keys {
		_, err := conn.Do("DEL", key)
		if err != nil {
			return err
		}
	}
	return nil
}

// redis-end

// mongo
func GetSingleResultByID(id primitive.ObjectID, collectionName string) (*mongo.SingleResult, error) {
	client, err := GetMongoClient()
	if err != nil {
		return nil, err
	}

	return client.Database(constant.DB).Collection(collectionName).FindOne(context.TODO(), bson.M{"_id": id}), nil
}
func GetSingleResultByFirebaseID(id string, collectionName string) (*mongo.SingleResult, error) {
	client, err := GetMongoClient()
	if err != nil {
		return nil, err
	}

	return client.Database(constant.DB).Collection(collectionName).FindOne(context.TODO(), bson.M{"userID": id}), nil
}

// player-start
func GetPlayerByID(id string) (structs.Player, error) {
	player := structs.Player{}

	playerID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return player, err
	}
	result, err := GetSingleResultByID(playerID, constant.PLAYERS)
	if err != nil {
		return player, err
	}

	err = result.Decode(&player)
	if err != nil {
		return player, err
	}

	return player, nil
}
func SearchPlayerByFilter(filter, filterOptions bson.D, limit int64) ([]structs.Player, error) {
	var players []structs.Player
	client, err := GetMongoClient()
	if err != nil {
		return players, err
	}

	cursor, err := client.Database(constant.DB).Collection(constant.PLAYERS).Find(context.TODO(), filter,
		options.Find().SetSort(filterOptions).SetLimit(limit))
	if err != nil {
		return players, err
	}
	if err = cursor.All(context.TODO(), &players); err != nil {
		return players, err
	}

	return players, err
}
func AddFilterViaFields(f *request.Filter) bson.D {
	filter := bson.D{}

	if len(f.Age) == 2 {
		filter = append(filter, bson.E{Key: "age", Value: bson.D{
			{Key: "$gte", Value: f.Age[0]},
			{Key: "$lte", Value: f.Age[1]},
		}},
		)
	} else if len(f.Age) == 1 {
		filter = append(filter, bson.E{Key: "age", Value: f.Age[0]})
	}

	if len(f.Name) > 0 {
		filter = append(filter, bson.E{Key: "long_name", Value: bson.D{
			{Key: "$regex", Value: primitive.Regex{Pattern: f.Name, Options: "i"}},
		}},
		)
	}
	if len(f.Club) > 0 {
		filter = append(filter, bson.E{Key: "club_name", Value: bson.D{
			{Key: "$regex", Value: primitive.Regex{Pattern: f.Club, Options: "i"}},
		}},
		)
	}
	if len(f.Nationality) > 0 {
		filter = append(filter, bson.E{Key: "nationality_name", Value: bson.D{
			{Key: "$regex", Value: primitive.Regex{Pattern: f.Nationality, Options: "i"}},
		}},
		)
	}
	if len(f.League) > 0 {
		filter = append(filter, bson.E{Key: "league_name", Value: bson.D{
			{Key: "$regex", Value: primitive.Regex{Pattern: f.League, Options: "i"}},
		}},
		)
	}
	if len(f.Position) > 0 {
		filter = append(filter, bson.E{Key: "player_positions", Value: bson.D{
			{Key: "$regex", Value: primitive.Regex{Pattern: f.Position, Options: "i"}},
		}},
		)
	}
	if len(f.Overall) == 2 {
		filter = append(filter, bson.E{Key: "overall", Value: bson.D{
			{Key: "$gte", Value: f.Overall[0]},
			{Key: "$lte", Value: f.Overall[1]},
		}},
		)
	} else if len(f.Overall) == 1 {
		filter = append(filter, bson.E{Key: "overall", Value: f.Overall[0]})
	}

	if len(f.Potential) == 2 {
		filter = append(filter, bson.E{Key: "potential", Value: bson.D{
			{Key: "$gte", Value: f.Potential[0]},
			{Key: "$lte", Value: f.Potential[1]},
		}},
		)
	} else if len(f.Potential) == 1 {
		filter = append(filter, bson.E{Key: "potential", Value: f.Potential[0]})
	}

	return filter
}
func AddFilterViaType(packType int) (bson.D, int) {
	filter := bson.D{}
	minOverall := 60
	maxOverall := 64

	randomIndex := GetRandom(1, 11) * 100

	if constant.SILVER == packType {
		minOverall = 65
		maxOverall = 69
	} else if constant.PREMIUMSILVER == packType {
		minOverall = 70
		maxOverall = 74
	} else if constant.GOLD == packType {
		minOverall = 75
		maxOverall = 79
	} else if constant.PREMIUMGOLD == packType {
		minOverall = 77
		maxOverall = 83
	} else if constant.ULTIMATEGOLD == packType {
		minOverall = 81
		maxOverall = 87
	} else if constant.PRIMEGOLD == packType {
		minOverall = 84
		maxOverall = 99
	}

	filter = append(filter, bson.E{Key: "overall", Value: bson.D{
		{Key: "$gte", Value: minOverall},
		{Key: "$lte", Value: maxOverall},
	}},
	)

	return filter, randomIndex
}

// player-end

func GetEnv(key string) (string, error) {
	val, ok := os.LookupEnv(key)
	if !ok {
		return "", fmt.Errorf("key didn't set before key:%s", key)
	}
	return val, nil
}
func GetRandom(min, max int) int {
	rand.Seed(time.Now().UnixNano())
	return rand.Intn(max-min) + min
}

func ReturnError(w http.ResponseWriter, code int, err error, message string) {
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(request.ErrorResponse{
		Code:    code,
		Message: message,
		Error:   err.Error(),
	})
}

func CreateCollection(db *mongo.Database, collectionName string) error {
	collNames, err := db.ListCollectionNames(context.TODO(), bson.D{})
	if err != nil {
		return err
	}

	exist := false
	for _, coll := range collNames {
		if coll == collectionName {
			exist = true
			break
		}
	}

	if !exist {
		db.CreateCollection(context.TODO(), collectionName)
	}

	return nil
}

//crud opt for manager
func CreateManager(manager structs.Manager) (structs.Insert, error) {
	insert := structs.Insert{}
	client, err := GetMongoClient()
	if err != nil {
		return insert, err
	}

	db := client.Database(constant.DB)
	err = CreateCollection(db, constant.MANAGERS)
	if err != nil {
		return insert, err
	}

	doc, err := db.Collection(constant.MANAGERS).InsertOne(context.TODO(), bson.D{
		{Key: "name", Value: manager.Name},
		{Key: "userID", Value: manager.UserID},
		{Key: "email", Value: manager.Email},
		{Key: "players", Value: bson.A{}},
		{Key: "points", Value: 0},
		{Key: "results", Value: bson.A{}},
	})
	if err != nil {
		return insert, err
	}

	docByte, err := json.Marshal(doc)
	if err != nil {
		return insert, err
	}

	err = json.Unmarshal(docByte, &insert)
	if err != nil {
		return insert, err
	}

	return insert, nil
}
func GetManagers() ([]structs.Manager, error) {
	var managers []structs.Manager
	client, err := GetMongoClient()
	if err != nil {
		return managers, err
	}

	cursor, err := client.Database(constant.DB).Collection(constant.MANAGERS).Find(context.TODO(), bson.M{}, nil)
	if err != nil {
		return managers, err
	}
	if err = cursor.All(context.TODO(), &managers); err != nil {
		return managers, err
	}

	return managers, err
}
func GetManager(id string) (structs.Manager, error) {
	manager := structs.Manager{}

	result, err := GetSingleResultByFirebaseID(id, constant.MANAGERS)
	if err != nil {
		return manager, err
	}

	err = result.Decode(&manager)
	if err != nil {
		return manager, err
	}

	return manager, nil
}
func GetManagerByID(id string) (structs.Manager, error) {
	manager := structs.Manager{}

	managerID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return manager, err
	}
	result, err := GetSingleResultByID(managerID, constant.MANAGERS)
	if err != nil {
		return manager, err
	}

	err = result.Decode(&manager)
	if err != nil {
		return manager, err
	}

	return manager, nil
}
func UpdateManager(man *structs.Manager) (*mongo.UpdateResult, error) {
	result := &mongo.UpdateResult{}
	client, err := GetMongoClient()
	if err != nil {
		return result, err
	}

	result, err = client.Database(constant.DB).Collection(constant.MANAGERS).ReplaceOne(context.TODO(), bson.M{"_id": man.ID}, man)
	if err != nil {
		return result, err
	}

	return result, nil
}

//manager-end

//crud opt for season
func CreateSeason(season structs.Season) (structs.Insert, error) {
	insert := structs.Insert{}
	client, err := GetMongoClient()
	if err != nil {
		return insert, err
	}

	db := client.Database(constant.DB)
	err = CreateCollection(db, constant.SEASONS)
	if err != nil {
		return insert, err
	}

	doc, err := db.Collection(constant.SEASONS).InsertOne(context.TODO(), bson.D{
		{Key: "type", Value: season.Type},
		{Key: "title", Value: season.Title},
		{Key: "results", Value: bson.A{}},
	})
	if err != nil {
		return insert, err
	}

	docByte, err := json.Marshal(doc)
	if err != nil {
		return insert, err
	}

	err = json.Unmarshal(docByte, &insert)
	if err != nil {
		return insert, err
	}

	return insert, nil
}
func GetSeasons() ([]request.SeasonResponse, error) {
	var seasons []request.SeasonResponse
	client, err := GetMongoClient()
	if err != nil {
		return seasons, err
	}

	cursor, err := client.Database(constant.DB).Collection(constant.SEASONS).Find(context.TODO(), bson.M{}, nil)
	if err != nil {
		return seasons, err
	}
	if err = cursor.All(context.TODO(), &seasons); err != nil {
		return seasons, err
	}

	return seasons, err
}
func GetSeasonByID(id string) (structs.Season, error) {
	season := structs.Season{}

	seasonID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return season, err
	}
	result, err := GetSingleResultByID(seasonID, constant.SEASONS)
	if err != nil {
		return season, err
	}

	err = result.Decode(&season)
	if err != nil {
		return season, err
	}

	return season, nil
}
func UpdateSeason(season *structs.Season) (*mongo.UpdateResult, error) {
	result := &mongo.UpdateResult{}
	client, err := GetMongoClient()
	if err != nil {
		return result, err
	}

	result, err = client.Database(constant.DB).Collection(constant.SEASONS).ReplaceOne(context.TODO(), bson.M{"_id": season.ID}, season)
	if err != nil {
		return result, err
	}

	return result, nil
}
func GetStanding(results []structs.Result) []structs.Standing {
	standingMap := make(map[string]*structs.Standing)
	for _, result := range results {
		if result.Score[0] > result.Score[1] {
			_, found := standingMap[result.HomeManager]
			if found {
				standingMap[result.HomeManager].Set(structs.Standing{
					Manager: result.HomeManager,
					Points:  3,
					Played:  1,
					Won:     1,
					GF:      result.Score[0],
					GA:      result.Score[1],
					GD:      result.Score[0] - result.Score[1],
					Form:    []string{"W"},
				})
			} else {
				standingMap[result.HomeManager] = &structs.Standing{
					Manager: result.HomeManager,
					Points:  3,
					Played:  1,
					Won:     1,
					GF:      result.Score[0],
					GA:      result.Score[1],
					GD:      result.Score[0] - result.Score[1],
					Form:    []string{"W"},
				}
			}
			_, foundAway := standingMap[result.AwayManager]
			if foundAway {
				standingMap[result.AwayManager].Set(structs.Standing{
					Manager: result.AwayManager,
					Played:  1,
					Lost:    1,
					GA:      result.Score[0],
					GF:      result.Score[1],
					GD:      result.Score[1] - result.Score[0],
					Form:    []string{"L"},
				})
			} else {
				standingMap[result.AwayManager] = &structs.Standing{
					Manager: result.AwayManager,
					Played:  1,
					Lost:    1,
					GA:      result.Score[0],
					GF:      result.Score[1],
					GD:      result.Score[1] - result.Score[0],
					Form:    []string{"L"},
				}
			}
		} else if result.Score[1] > result.Score[0] {
			_, found := standingMap[result.HomeManager]
			if found {
				standingMap[result.HomeManager].Set(structs.Standing{
					Manager: result.HomeManager,
					Played:  1,
					Lost:    1,
					GF:      result.Score[0],
					GA:      result.Score[1],
					GD:      result.Score[0] - result.Score[1],
					Form:    []string{"L"},
				})
			} else {
				standingMap[result.HomeManager] = &structs.Standing{
					Manager: result.HomeManager,
					Played:  1,
					Lost:    1,
					GF:      result.Score[0],
					GA:      result.Score[1],
					GD:      result.Score[0] - result.Score[1],
					Form:    []string{"L"},
				}
			}
			_, foundAway := standingMap[result.AwayManager]
			if foundAway {
				standingMap[result.AwayManager].Set(structs.Standing{
					Manager: result.AwayManager,
					Played:  1,
					Points:  3,
					Won:     1,
					GA:      result.Score[0],
					GF:      result.Score[1],
					GD:      result.Score[1] - result.Score[0],
					Form:    []string{"W"},
				})
			} else {
				standingMap[result.AwayManager] = &structs.Standing{
					Manager: result.AwayManager,
					Played:  1,
					Points:  3,
					Won:     1,
					GA:      result.Score[0],
					GF:      result.Score[1],
					GD:      result.Score[1] - result.Score[0],
					Form:    []string{"W"},
				}
			}
		} else {
			_, found := standingMap[result.HomeManager]
			if found {
				standingMap[result.HomeManager].Set(structs.Standing{
					Manager: result.HomeManager,
					Played:  1,
					Draw:    1,
					Points:  1,
					GF:      result.Score[0],
					GA:      result.Score[1],
					GD:      result.Score[0] - result.Score[1],
					Form:    []string{"D"},
				})
			} else {
				standingMap[result.HomeManager] = &structs.Standing{
					Manager: result.HomeManager,
					Played:  1,
					Points:  1,
					Draw:    1,
					GF:      result.Score[0],
					GA:      result.Score[1],
					GD:      result.Score[0] - result.Score[1],
					Form:    []string{"D"},
				}
			}
			_, foundAway := standingMap[result.AwayManager]
			if foundAway {
				standingMap[result.AwayManager].Set(structs.Standing{
					Manager: result.AwayManager,
					Played:  1,
					Points:  1,
					Draw:    1,
					GA:      result.Score[0],
					GF:      result.Score[1],
					GD:      result.Score[1] - result.Score[0],
					Form:    []string{"D"},
				})
			} else {
				standingMap[result.AwayManager] = &structs.Standing{
					Manager: result.AwayManager,
					Played:  1,
					Points:  1,
					Draw:    1,
					GA:      result.Score[0],
					GF:      result.Score[1],
					GD:      result.Score[1] - result.Score[0],
					Form:    []string{"D"},
				}
			}
		}
	}
	standing := []structs.Standing{}
	for _, s := range standingMap {
		standing = append(standing, *s)
	}

	sort.Slice(standing, func(i, j int) bool {
		return standing[i].Points > standing[j].Points
	})
	return standing
}
func GetStats(results []structs.Result) []structs.Stats {
	statsMap := make(map[string]*structs.Stats)
	for _, result := range results {
		for _, scorer := range result.HomeScorers {
			_, found := statsMap[scorer.Player.Name]
			if found {
				statsMap[scorer.Player.Name].Count += scorer.Count
			} else {
				statsMap[scorer.Player.Name] = &structs.Stats{
					Manager: result.HomeManager,
					Player:  scorer.Player.Name,
					FaceUrl: scorer.Player.FaceUrl,
					Count:   scorer.Count,
				}
			}
		}

		for _, scorer := range result.AwayScorers {
			_, found := statsMap[scorer.Player.Name]
			if found {
				statsMap[scorer.Player.Name].Count += scorer.Count
			} else {
				statsMap[scorer.Player.Name] = &structs.Stats{
					Manager: result.AwayManager,
					Player:  scorer.Player.Name,
					FaceUrl: scorer.Player.FaceUrl,
					Count:   scorer.Count,
				}
			}
		}
	}

	stats := []structs.Stats{}
	for _, s := range statsMap {
		stats = append(stats, *s)
	}

	sort.Slice(stats, func(i, j int) bool {
		return stats[i].Count > stats[j].Count
	})

	return stats
}

//season-end

//result-locig-start
func CreateResult(result structs.Result) (structs.Insert, error) {
	insert := structs.Insert{}
	client, err := GetMongoClient()
	if err != nil {
		return insert, err
	}

	db := client.Database(constant.DB)
	err = CreateCollection(db, constant.RESULTS)
	if err != nil {
		return insert, err
	}

	doc, err := db.Collection(constant.RESULTS).InsertOne(context.TODO(), bson.D{
		{Key: "season", Value: result.Season},
		{Key: "home", Value: result.Home},
		{Key: "away", Value: result.Away},
		{Key: "seasonType", Value: result.SeasonType},
		{Key: "seasonTitle", Value: result.SeasonTitle},
		{Key: "homeManager", Value: result.HomeManager},
		{Key: "awayManager", Value: result.AwayManager},
		{Key: "score", Value: result.Score},
		{Key: "homescorers", Value: result.HomeScorers},
		{Key: "awayscorers", Value: result.AwayScorers},
	})
	if err != nil {
		return insert, err
	}

	docByte, err := json.Marshal(doc)
	if err != nil {
		return insert, err
	}

	err = json.Unmarshal(docByte, &insert)
	if err != nil {
		return insert, err
	}

	return insert, nil
}
