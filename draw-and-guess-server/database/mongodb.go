package database

import (
	"context"
	"log"
	"time"

	"draw-and-guess-server/config"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var (
	Client   *mongo.Client
	Database *mongo.Database
)

func Connect() error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	clientOptions := options.Client().ApplyURI(config.MongoURI)

	client, err := mongo.Connect(ctx, clientOptions)
	if err != nil {
		return err
	}

	err = client.Ping(ctx, nil)
	if err != nil {
		return err
	}

	Client = client
	Database = client.Database(config.DatabaseName)

	log.Println("Connected to MongoDB!")
	return nil
}

func Disconnect() error {
	if Client == nil {
		return nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	return Client.Disconnect(ctx)
}

func GetCollection(name string) *mongo.Collection {
	if Database == nil {
		return nil
	}
	return Database.Collection(name)
}
