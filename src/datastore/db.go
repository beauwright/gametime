package datastore

import (
	"context"
	"errors"
	"fmt"
	"time"

	"gametime/internal/utils"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
	"go.mongodb.org/mongo-driver/v2/mongo/readpref"
)

type ClockEventType string
const (
	START ClockEventType = "START"
	STOP        = "STOP"
	ADD        = "ADD"
	SUB        = "SUB"
)

type ClockState string
const (
    RUNNING ClockState = "RUNNING"
    STOPPED  = "STOPPED"
)

type ClockEvent struct {
    EventType ClockEventType
    Timestamp time.Time
    // The amount of time remaining on the clock when emitted
    TimeRemaining time.Duration
    Detail interface{}
}

type Clock struct {
	ID            string
	Name          string
        EventLog      []ClockEvent
	Increment     time.Duration

        // TODO: Either remove these fields, replacing with computations from the EventLog, or ensure we update these always
	TimeRemaining time.Duration
	InitialTime time.Duration
}

func (c *Clock) latestStateChange() ClockEvent {
    relevant := utils.Filter(c.EventLog, func(ce ClockEvent) bool {
        return ce.EventType == START || ce.EventType == STOP
    })

    return relevant[len(relevant)-1]
}

func (c *Clock) State() ClockState {
    latest := c.latestStateChange()

    if latest.EventType == START {
        return RUNNING
    } else {
        return STOPPED
    }
}

func (c *Clock) getStopEvent() (*ClockEvent, error) {
    if c.State() != RUNNING {
        return nil, errors.New("clock is not running")
    }

    latest := c.latestStateChange()

    elapsed := time.Now().Sub(latest.Timestamp)
    newRemaining := latest.TimeRemaining - elapsed


    return &ClockEvent{
        EventType: STOP,
        Timestamp: time.Now(),
        TimeRemaining: newRemaining,
    }, nil
}

func (c *Clock) getStartEvent() (*ClockEvent, error) {
    if c.State() != STOPPED {
        return nil, errors.New("clock is not stopped")
    }

    latest := c.latestStateChange()

    elapsed := time.Now().Sub(latest.Timestamp)
    newRemaining := latest.TimeRemaining - elapsed


    return &ClockEvent{
        EventType: START,
        Timestamp: time.Now(),
        TimeRemaining: newRemaining,
    }, nil
}


type GameState struct {
	ActiveClockID string
	NextClockID   string
        Running bool
	Clocks        []Clock
}

type GameConfig struct{}

type Lobby struct {
	ID     string
	State  GameState
	Config GameConfig
}

func (l *Lobby) ClockByID(clockID string) (int, *Clock) {
    for i, c := range l.State.Clocks {
        if c.ID == clockID {
            return i, &c
        }
    }

    return -1, nil
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

func (db *GametimeDB) AdvanceLobby(ctx context.Context, clockID string) (*Lobby, error) {
    lobby, err := db.GetLobbyByClock(ctx, clockID)
    if err != nil {
        return nil, err
    }

    stopIndex, stopClock := lobby.ClockByID(clockID)
    stopEvent, err := stopClock.getStopEvent()
    if err != nil {
        return nil, err
    }
    stopPath := fmt.Sprintf("state.clocks.%d.eventlog", stopIndex)

    startIndex, startClock := lobby.ClockByID(lobby.State.NextClockID)
    startEvent, err := startClock.getStartEvent()
    if err != nil {
        return  nil, err
    }
    startPath := fmt.Sprintf("state.clocks.%d.eventlog", startIndex)

    upcomingIndex := startIndex+1
    if upcomingIndex >= len(lobby.State.Clocks) {
        upcomingIndex = 0
    }
    upcomingClock := lobby.State.Clocks[upcomingIndex]


    filter := bson.D{{"id", lobby.ID}}
    update := bson.D{
        {
            "$push", bson.D{
                bson.E{stopPath, stopEvent},
                bson.E{startPath, startEvent},
            },
        },
        {
            "$set", bson.D{
                bson.E{"state.activeclockid", lobby.State.NextClockID},
                bson.E{"state.nextclockid", upcomingClock.ID},
            },
        },
    }

    asdf := options.FindOneAndUpdate().SetReturnDocument(options.After)

    var result Lobby
    err = db.db.Database(database).Collection(lobbies).FindOneAndUpdate(ctx, filter,update, asdf).Decode(&result)
    if err != nil {
        return nil, err
    }

    return &result, nil
}

func (db *GametimeDB) StartLobby(ctx context.Context, lobbyID string) error {
    lobby, err := db.GetLobby(ctx, lobbyID)
    if err != nil {
        return err
    }

    if lobby.State.Running {
        return errors.New("lobby already running")
    }

    index, clock := lobby.ClockByID(lobby.State.ActiveClockID)
    if clock == nil {
        return errors.New("active clock does not exist")
    }
    path := fmt.Sprintf("state.clocks.%d.eventlog", index)

    startEvent, err := clock.getStartEvent()
    if err != nil {
        return err
    }


    filter := bson.D{{"id", lobby.ID}}
    update := bson.D{
        {
            "$push", bson.D{
                bson.E{path, startEvent},
            },
        },
        {
            "$set", bson.D{
                bson.E{"state.running", true},
            },
        },
    }

    _, err = db.db.Database(database).Collection(lobbies).UpdateOne(ctx, filter,update)
    if err != nil {
        return  err
    }

    return nil

}
