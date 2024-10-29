package datastore

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
	"go.mongodb.org/mongo-driver/v2/mongo/readpref"
)

type Clock struct {
	ID            string
	Name          string
	Position      int
	Increment     time.Duration
	TimeRemaining time.Duration
}


type GameState struct {
	ActiveClockID string
	NextClockID   string
	Clocks        []Clock
}

type GameConfig struct{}

type Lobby struct {
	ID     string
	State  GameState
	Config GameConfig
}

type GametimeDB struct {
    db mongo.Client
}

var (
    database = "gametime"
    lobbies = "lobbies"
)

func New(connString string) (*GametimeDB, error) {
    dbClient, err := mongo.Connect(options.Client().ApplyURI(connString))
    if err != nil {
        return nil, err
    }
    ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
    defer cancel()

    // TODO: Rather than ping, just try to create our indexes. Which accomplishes the same thing.
    err = dbClient.Ping(ctx, readpref.Primary())
    if err != nil {
        return nil, err
    }

    return &GametimeDB{
        db: *dbClient,

    }, nil


}

func (db *GametimeDB) SaveLobby(ctx context.Context, lobby Lobby) error {
    _, err := db.db.Database(database).Collection(lobbies).InsertOne(ctx, lobby)
    return err
}

func (db *GametimeDB) GetLobby(ctx context.Context, lobbyID string) (*Lobby, error) {
    filter := bson.D{{"id", lobbyID}}

    var result Lobby
    err := db.db.Database(database).Collection(lobbies).FindOne(ctx, filter).Decode(&result)
    if err != nil {
        return nil, err
    }

    return &result, err
}

func (db *GametimeDB) GetLobbyByClock(ctx context.Context, clockID string) (*Lobby, error) {
    filter := bson.D{{"state.clocks.id", clockID}}

    var result Lobby
    err := db.db.Database(database).Collection(lobbies).FindOne(ctx, filter).Decode(&result)
    if err != nil {
        return nil, err
    }

    return &result, err
}
