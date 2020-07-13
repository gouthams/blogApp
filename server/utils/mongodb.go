package utils

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var db *mongo.Database
var ctx context.Context

const dbName = "blogDB"
const blogUserCollection = "blogUser"
const blogPostCollection = "blogPost"

func ConnectToDatabase() *mongo.Database {
	logEntry := Log()
	// Set client options
	client, err := mongo.NewClient(options.Client().ApplyURI("mongodb://localhost:27017"))
	if err != nil {
		logEntry.Fatalf("Db client get failed, %v", err)
	}

	ctx, _ := context.WithTimeout(context.Background(), 60*time.Second)
	client.Connect(ctx)
	// Connect to MongoDB
	if err != nil {
		logEntry.Fatalf("Db connection failed, %v", err)
	}
	//defer client.Disconnect(ctx)

	// Check the connection
	err = client.Ping(ctx, nil)
	if err != nil {
		logEntry.Fatalf("Db ping failed, %v", err)
	}

	// Create the DB if does not exist
	db = client.Database(dbName)
	logEntry.Infof("Created Db: %s -> %v ", db.Name(), dbName)

	return db
}

func GetDb() (*mongo.Database, context.Context) {
	if db == nil {
		db = ConnectToDatabase()
	}
	return db, ctx
}

func GetUserCollection() (*mongo.Collection, context.Context) {
	if db == nil {
		db = ConnectToDatabase()
	}

	return db.Collection(blogUserCollection), ctx
}

func GetPostCollection() (*mongo.Collection, context.Context) {
	if db == nil {
		db = ConnectToDatabase()
	}

	return db.Collection(blogPostCollection), ctx
}

func FlushCollections() error {
	logEntry := Log()
	userCollection, ctx := GetUserCollection()
	err := userCollection.Drop(ctx)
	if err != nil {
		logEntry.Errorf("Drop on user collection failed %v", err)
		return err
	}

	blogPostCollection, ctx := GetPostCollection()
	err = blogPostCollection.Drop(ctx)
	if err != nil {
		logEntry.Errorf("Drop on blogPost collection failed %v", err)
		return err
	}

	return nil
}
