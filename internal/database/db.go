package database

import (
	"context"

	"go.mongodb.org/mongo-driver/mongo"
)

type DBTX interface {
	Database(name string) *mongo.Database
	StartSession() (mongo.SessionContext, error)
}

type MongoDBClient struct {
	*mongo.Client
	DBName string
}

func (c *MongoDBClient) Database(name string) *mongo.Database {
	return c.Client.Database(name)
}

func (c *MongoDBClient) StartSession() (mongo.SessionContext, error) {
	session, err := c.Client.StartSession()
	if err != nil {
		return nil, err
	}
	return mongo.NewSessionContext(context.Background(), session), nil
}
