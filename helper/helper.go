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
func SetRedisData(pool *redis.Pool, key string, players []constant.Player, duration time.Duration) error {
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
func GetRedisData(pool *redis.Pool, key string, players *[]constant.Player) error {
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
func GenerateRedisKey(filter *constant.Filter) string {
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

// player-start
func GetPlayerByID(id primitive.ObjectID) (constant.Player, error) {
	player := constant.Player{}

	result, err := GetSingleResultByID(id, constant.PLAYERS)
	if err != nil {
		return player, err
	}

	err = result.Decode(&player)
	if err != nil {
		return player, err
	}

	return player, nil
}
func SearchPlayerByFilter(filter bson.D) ([]constant.Player, error) {
	var players []constant.Player
	client, err := GetMongoClient()
	if err != nil {
		return players, err
	}

	cursor, err := client.Database(constant.DB).Collection(constant.PLAYERS).Find(context.TODO(), filter,
		options.Find().SetSort(bson.D{{Key: "overall", Value: -1}}))
	if err != nil {
		return players, err
	}
	if err = cursor.All(context.TODO(), &players); err != nil {
		return players, err
	}

	return players, err
}
func AddFilterViaFields(f *constant.Filter) bson.D {
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

// player-end

func GetEnv(key string) (string, error) {
	val, ok := os.LookupEnv(key)
	if !ok {
		return "", fmt.Errorf("key didn't set before key:%s", key)
	}
	return val, nil
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
func CreateManager(manager constant.Manager) error {
	client, err := GetMongoClient()
	if err != nil {
		return err
	}

	db := client.Database(constant.DB)
	err = CreateCollection(db, constant.MANAGERS)
	if err != nil {
		return err
	}

	doc, err := db.Collection(constant.MANAGERS).InsertOne(context.TODO(), bson.D{
		{Key: "name", Value: manager.Name},
		{Key: "players", Value: bson.A{}},
		{Key: "points", Value: 0},
		{Key: "results", Value: bson.A{}},
	})
	if err != nil {
		return err
	}
	fmt.Println("Manager added to collection", doc.InsertedID)
	return nil
}
func GetManagerByID(id primitive.ObjectID) (constant.Manager, error) {
	manager := constant.Manager{}

	result, err := GetSingleResultByID(id, constant.MANAGERS)
	if err != nil {
		return manager, err
	}

	err = result.Decode(&manager)
	if err != nil {
		return manager, err
	}

	return manager, nil
}
func UpdateManager(man *constant.Manager) (*mongo.UpdateResult, error) {
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

//
