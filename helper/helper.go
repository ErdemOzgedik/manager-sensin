package helper

import (
	"context"
	"encoding/json"
	"fmt"
	"manager-sensin/constant"
	"net/url"
	"os"
	"sync"
	"time"

	"github.com/go-redis/redis/v8"
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
func GetRedisClient() *redis.Client {
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

	rdb := redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: pass, // no password set
		DB:       0,    // use default DB
	})

	return rdb
}
func SetRedisData(rdb *redis.Client, key string, value []constant.Player, duration time.Duration) error {
	v, err := json.Marshal(value)
	if err != nil {
		return err
	}
	return rdb.Set(context.TODO(), key, v, duration).Err()
}
func GetRedisData(rdb *redis.Client, key string, players *[]constant.Player) error {
	val, err := rdb.Get(context.TODO(), key).Result()
	if err != nil {
		return err
	}
	return json.Unmarshal([]byte(val), &players)
}
func GenerateRedisKey(filter *constant.Filter) string {
	return fmt.Sprintf("%s-%s-%s-%s-%s-%v-%v-%v", filter.Name,
		filter.Club, filter.Nationality, filter.League,
		filter.Position, filter.Age, filter.Overall, filter.Potential)
}
func AddFilterViaFields(f *constant.Filter) bson.D {
	filter := bson.D{}

	if len(f.Age) == 2 {
		filter = append(filter, bson.E{Key: "age", Value: bson.D{
			{Key: "$gte", Value: f.Age[0]},
			{Key: "$lte", Value: f.Age[1]},
		}},
		)
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
	}
	if len(f.Potential) == 2 {
		filter = append(filter, bson.E{Key: "potential", Value: bson.D{
			{Key: "$gte", Value: f.Potential[0]},
			{Key: "$lte", Value: f.Potential[1]},
		}},
		)
	}
	return filter
}
func SearchByFilter(filter bson.D) ([]constant.Player, error) {
	var players []constant.Player
	client, err := GetMongoClient()
	if err != nil {
		return players, err
	}

	cursor, err := client.Database(constant.DB).Collection(constant.ISSUES).Find(context.TODO(), filter,
		options.Find().SetSort(bson.D{{Key: "overall", Value: -1}}))
	if err != nil {
		return players, err
	}
	if err = cursor.All(context.TODO(), &players); err != nil {
		return players, err
	}

	return players, err
}
func GetEnv(key string) (string, error) {
	val, ok := os.LookupEnv(key)
	if !ok {
		return "", fmt.Errorf("key didn't set before key:%s", key)
	}
	return val, nil
}
