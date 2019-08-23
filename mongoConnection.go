package main

import (
	"context"
	"log"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func getClient() *mongo.Client {
	clientOptions := options.Client().ApplyURI("mongodb://localhost:27017")
	client, err := mongo.NewClient(clientOptions)
	if err != nil {
		log.Fatal(err)
	}
	err = client.Connect(context.Background())
	if err != nil {
		log.Fatal(err)
	}
	return client
}

func insertFoodProduct(client *mongo.Client, food Product) interface{} {
	collection := client.Database("OpenFood").Collection("FoodSchema")
	insertResult, err := collection.InsertOne(context.TODO(), food)
	if err != nil {
		log.Fatalln("Error on inserting Food", err)
	}
	return insertResult.InsertedID
}
