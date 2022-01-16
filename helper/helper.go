package helper

import (
	"context"
	"encoding/json"
	"fmt"
	"manager-sensin/constant"
	"sync"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/spf13/viper"
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

//I have used below constants just to hold required database config's.
const (
	DB     = "futManagerDB"
	ISSUES = "fut22Collection"
)

//GetMongoClient - Return mongodb connection to work with
func GetMongoClient() (*mongo.Client, error) {
	//Perform connection creation operation only once.
	mongoOnce.Do(func() {
		// Set client options
		clientOptions := options.Client().ApplyURI(viper.GetString("database.connectionstring"))
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
	rdb := redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "", // no password set
		DB:       0,  // use default DB
	})

	return rdb
}
func SetRedisData(rdb *redis.Client, key string, value []constant.Player) error {
	v, err := json.Marshal(value)
	if err != nil {
		return err
	}
	return rdb.Set(context.TODO(), key, v, 45*time.Minute).Err()
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
	cursor, err := client.Database(DB).Collection(ISSUES).Find(context.TODO(), filter)
	if err != nil {
		return players, err
	}
	if err = cursor.All(context.TODO(), &players); err != nil {
		return players, err
	}

	return players, err
}
