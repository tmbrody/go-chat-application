package database

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"strings"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"golang.org/x/crypto/bcrypt"
)

// CreateUser creates a new user in the database.
func (client *MongoDBClient) CreateUser(name, email, password string) (User, error) {
	// Get the users collection from the database.
	collection := client.Database(client.DBName).Collection("users")
	if collection == nil {
		return User{}, fmt.Errorf("database is nil")
	}

	// Hash the password using bcrypt.
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return User{}, err
	}

	// Create a new user.
	user := User{
		Name:      name,
		Email:     email,
		Password:  string(hashedPassword),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	// Check if a user with the same email already exists.
	filter := bson.M{"email": email}
	err = collection.FindOne(context.Background(), filter).Err()
	if err != mongo.ErrNoDocuments {
		if err != nil {
			return User{}, err
		}
		return User{}, errors.New("email already in use")
	}

	// Insert the new user into the database.
	response, err := collection.InsertOne(context.Background(), user)
	if err != nil {
		return User{}, nil
	}

	// Create a response with the inserted ID.
	userResponse := User{
		ID:        response.InsertedID.(primitive.ObjectID),
		Name:      user.Name,
		Email:     user.Email,
		Password:  user.Password,
		CreatedAt: user.CreatedAt,
		UpdatedAt: user.UpdatedAt,
	}

	return userResponse, nil
}

// GetAllUsers retrieves all users from the database.
func (client *MongoDBClient) GetAllUsers() ([]User, error) {
	// Get the users collection from the database.
	collection := client.Database(client.DBName).Collection("users")
	if collection == nil {
		return []User{}, fmt.Errorf("database is nil")
	}

	var allUsers []User

	// Find all users in the collection.
	cursor, err := collection.Find(context.Background(), bson.M{})
	if err != nil {
		return []User{}, err
	}
	defer cursor.Close(context.Background())

	// Decode the cursor into the allUsers slice.
	err = cursor.All(context.Background(), &allUsers)
	if err != nil {
		return []User{}, err
	}

	return allUsers, nil
}

// UpdateUser updates the user with the given ID in the MongoDB database.
func (client *MongoDBClient) UpdateUser(id, name, email, password string) error {
	// Get the users collection from the database.
	collection := client.Database(client.DBName).Collection("users")
	if collection == nil {
		return fmt.Errorf("database is nil")
	}

	// Convert the string ID to MongoDB ObjectID.
	originalID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return err
	}

	// Define the filter and update operation for the update query.
	filter := bson.M{"_id": originalID}
	update := bson.M{
		"$set": bson.M{
			"name":       name,
			"email":      email,
			"password":   password,
			"updated_at": time.Now(),
		},
	}

	// Execute the update query.
	_, err = collection.UpdateOne(context.Background(), filter, update)
	if err != nil {
		return err
	}

	return nil
}

// DeleteUser deletes the user with the given ID from the MongoDB database.
func (client *MongoDBClient) DeleteUser(id string) error {
	// Get the users collection from the database.
	collection := client.Database(client.DBName).Collection("users")
	if collection == nil {
		return fmt.Errorf("database is nil")
	}

	// Convert the string ID to MongoDB ObjectID.
	originalID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return err
	}

	// Define the filter for the delete query.
	filter := bson.M{"_id": originalID}

	// Execute the delete query.
	result, err := collection.DeleteOne(context.Background(), filter)
	if err != nil {
		return err
	}

	// Check if a user was deleted.
	if result.DeletedCount == 0 {
		return errors.New("no user found with the given ID")
	}

	return nil
}

// GetDatabaseNAmeFromURL extracts the database name from the given MongoDB URL.
func GetDatabaseNAmeFromURL(dbURL string) (string, error) {
	// Parse the URL.
	u, err := url.Parse(dbURL)
	if err != nil {
		return "", err
	}

	// Extract the database name from the URL path.
	dbName := strings.TrimPrefix(u.Path, "/")
	return dbName, nil
}
