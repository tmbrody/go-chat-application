package database

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type User struct {
	ID        primitive.ObjectID `bson:"_id,omitempty"`
	Name      string             `bson:"name"`
	Email     string             `bson:"email"`
	Password  string             `bson:"password"`
	CreatedAt time.Time          `bson:"created_at"`
	UpdatedAt time.Time          `bson:"updated_at"`
}

type Conversation struct {
	ID        primitive.ObjectID   `bson:"_id,omitempty"`
	Name      string               `bson:"name"`
	Users     []primitive.ObjectID `bson:"users"`
	CreatedAt time.Time            `bson:"created_at"`
	UpdatedAt time.Time            `bson:"updated_at"`
}

type Message struct {
	ID             primitive.ObjectID `bson:"_id,omitempty"`
	ConversationID primitive.ObjectID `bson:"conversation_id"`
	SenderID       primitive.ObjectID `bson:"sender_id"`
	Content        string             `bson:"content"`
	CreatedAt      time.Time          `bson:"created_at"`
}
