package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

// UserModel type
type UserModel struct {
	Name      string             `json:"name" bson:"name"`
	Mail      string             `json:"mail" bson:"mail"`
	CreatedAt time.Time          `json:"createdAt" bson:"createdAt"`
	ID        primitive.ObjectID `json:"id" bson:"_id,omitempty"`
}

// UserService type
type UserService struct {
	col *mongo.Collection
}

func (s *UserService) Get(mail string) (UserModel, error) {
	var result UserModel
	err := s.col.FindOne(context.Background(), bson.M{"mail": mail}).Decode(&result)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Found a single document: %+v\n", result)
	return result, err
}

func (s *UserService) Add(user UserModel) error {
	user.ID = primitive.NewObjectID()
	user.CreatedAt = time.Now()
	insertResult, err := s.col.InsertOne(context.Background(), user)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("Inserted a single document: ", insertResult.InsertedID)
	return err
}

func (s *UserService) Update(mail string, user UserModel) error {
	updateResult, err := s.col.UpdateOne(context.Background(), bson.M{"mail": mail}, user)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Matched %v documents and updated %v documents.\n", updateResult.MatchedCount, updateResult.ModifiedCount)
	return err
}

func (s *UserService) Delete(mail string) error {
	deleteResult, err := s.col.DeleteMany(context.Background(), bson.M{"mail": mail})
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Deleted %v documents in the trainers collection\n", deleteResult.DeletedCount)
	return err
}
