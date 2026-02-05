/*
Package database is the middleware between the app database and the code. All data (de)serialization (save/load) from a
persistent database are handled here. Database specific logic should never escape this package.

To use this package you need to apply migrations to the database if needed/wanted, connect to it (using the database
data source name from config), and then initialize an instance of AppDatabase from the DB connection.

For example, this code adds a parameter in `webapi` executable for the database data source name (add it to the
main.WebAPIConfiguration structure):

	DB struct {
		Filename string `conf:""`
	}

This is an example on how to migrate the DB and connect to it:

	// Start Database
	logger.Println("initializing database support")
	db, err := sql.Open("sqlite3", "./foo.db")
	if err != nil {
		logger.WithError(err).Error("error opening SQLite DB")
		return fmt.Errorf("opening SQLite: %w", err)
	}
	defer func() {
		logger.Debug("database stopping")
		_ = db.Close()
	}()

Then you can initialize the AppDatabase and pass it to the api package.
*/
package database

import (
	"database/sql"
	"errors"
	"fmt"

	"github.com/SickRickX2/wasapp/service/api/schemas"

	_ "github.com/mattn/go-sqlite3"
)

// AppDatabase is the high level interface for the DB
type AppDatabase interface {
	// Login operatiobs
	CreateUser(u schemas.User) error
	FindUserByName(name string) (schemas.User, error)
	GetUserById(id schemas.UserId) (schemas.User, error)
	SetUserName(id schemas.UserId, newName string) error
	SearchUsers(query string) ([]schemas.User, error)
	CreateConversation(userA schemas.UserId, userB schemas.UserId) (schemas.PrivateConversation, error)
	CreateMessage(convId schemas.ConversationId, msg schemas.Message) error
	GetConversationMessages(convId schemas.ConversationId, limit int, beforeId string) ([]schemas.Message, error)
	DeleteMessage(convId schemas.ConversationId, messageId schemas.MessageId, userId schemas.UserId) (schemas.Message, error)
	CreateGroup(creator schemas.UserId, name string, participants []schemas.UserId) (schemas.Group, error)
	AddGroupMember(convId schemas.ConversationId, userId schemas.UserId) error
	RemoveGroupMember(convId schemas.ConversationId, userId schemas.UserId) error
	GetConversations(userId schemas.UserId) ([]schemas.Conversation, error)
	SaveMedia(media schemas.Media, mediaId string) error
	SetUserPhoto(userId schemas.UserId, photoUrl string) error
	SetGroupPhoto(convId schemas.ConversationId, photoUrl string) error
	IsUserInConversation(convId schemas.ConversationId, userId schemas.UserId) (bool, error)

	Ping() error
}

type appdbimpl struct {
	c *sql.DB
}

// New returns a new instance of AppDatabase based on the SQLite connection `db`.
// `db` is required - an error will be returned if `db` is `nil`.
func New(db *sql.DB) (AppDatabase, error) {
	if db == nil {
		return nil, errors.New("database is required when building a AppDatabase")
	}

	// Check if table exists. If not, the database is empty, and we need to create the structure
	var tableName string
	// TABLE users
	err := db.QueryRow(`SELECT name FROM sqlite_master WHERE type='table' AND name='users';`).Scan(&tableName)
	if errors.Is(err, sql.ErrNoRows) {
		// La tabella 'users' non esiste, creala.
		sqlStmt := `
            CREATE TABLE "users" (
				userId TEXT NOT NULL PRIMARY KEY,
                userName TEXT NOT NULL UNIQUE,
                createdAt DATETIME DEFAULT CURRENT_TIMESTAMP,
                pfpUrl TEXT
          );
        `
		_, err = db.Exec(sqlStmt)
		if err != nil {
			return nil, fmt.Errorf("error creating 'users' table: %w", err)
		}
	} else if err != nil {
		// Se c'è un errore nella query (es. connessione), fallo risalire
		return nil, fmt.Errorf("error querying 'users' table existence: %w", err)
	}

	// TABLE SESSIONS
	err = db.QueryRow(`SELECT name FROM sqlite_master WHERE type='table' AND name='sessions';`).Scan(&tableName)
	if errors.Is(err, sql.ErrNoRows) {
		sqlStmt := `
            CREATE TABLE sessions (
                token TEXT NOT NULL PRIMARY KEY,
                userId TEXT NOT NULL,
                createdAt DATETIME DEFAULT CURRENT_TIMESTAMP,
                expiresAt DATETIME DEFAULT NULL,

                FOREIGN KEY (userId) REFERENCES users(userId) ON DELETE CASCADE
            );
        `
		_, err = db.Exec(sqlStmt)
		if err != nil {
			return nil, fmt.Errorf("error creating 'sessions' table: %w", err)
		}
	} else if err != nil {
		return nil, fmt.Errorf("error querying 'sessions' table existence: %w", err)
	}
	// --------------------------------------------------------
	// TABLE CONVERSATIONS
	err = db.QueryRow(`SELECT name FROM sqlite_master WHERE type='table' AND name='conversations';`).Scan(&tableName)
	if errors.Is(err, sql.ErrNoRows) {
		sqlStmt := `CREATE TABLE conversations (
				convId TEXT NOT NULL PRIMARY KEY,
                kind TEXT NOT NULL DEFAULT 'private', -- private | group
                groupName TEXT DEFAULT NULL, 
                groupPhoto TEXT DEFAULT NULL, 
                createdAt DATETIME DEFAULT CURRENT_TIMESTAMP
            );`
		_, err = db.Exec(sqlStmt)
		if err != nil {
			return nil, fmt.Errorf("error creating 'conversations' table: %w", err)
		}
	} else if err != nil {
		// Se c'è un errore nella query (es. connessione), fallo risalire
		return nil, fmt.Errorf("error querying 'conversations' table existence: %w", err)
	}
	// --------------------------------------------------------
	//TABLE MESSAGES
	err = db.QueryRow(`SELECT name FROM sqlite_master WHERE type='table' AND name='messages';`).Scan(&tableName)
	if errors.Is(err, sql.ErrNoRows) {
		sqlStmt := `CREATE TABLE messages (
				messageId TEXT NOT NULL PRIMARY KEY,
                convId TEXT NOT NULL,	
                senderId TEXT NOT NULL,
                text TEXT DEFAULT NULL,
                mediaId TEXT DEFAULT NULL,
                status TEXT NOT NULL DEFAULT 'sent', -- sent|delivered|seen|deleted
                kind TEXT NOT NULL DEFAULT 'normal', -- normal|forwarded
                replyToId TEXT DEFAULT NULL,
                sentAt DATETIME DEFAULT CURRENT_TIMESTAMP,

                FOREIGN KEY (convId) REFERENCES conversations(convId) ON DELETE CASCADE,
                FOREIGN KEY (senderId) REFERENCES users(userId) ON DELETE CASCADE,
                FOREIGN KEY (replyToId) REFERENCES messages(messageId) ON DELETE SET NULL
			);`
		_, err = db.Exec(sqlStmt)
		if err != nil {
			return nil, fmt.Errorf("error creating 'messages' table: %w", err)
		}
	} else if err != nil {
		// Se c'è un errore nella query (es. connessione), fallo risalire
		return nil, fmt.Errorf("error querying 'messages' table existence: %w", err)
	}
	// --------------------------------------------------------
	// TABLE CONVERSATION_PARTICIPANTS

	err = db.QueryRow(`SELECT name FROM sqlite_master WHERE type='table' AND name='conversation_participants';`).Scan(&tableName)
	if errors.Is(err, sql.ErrNoRows) {
		sqlStmt := `
            CREATE TABLE conversation_participants (
                convId TEXT NOT NULL,
                userId TEXT NOT NULL,
                
                PRIMARY KEY (convId, userId),
                FOREIGN KEY (convId) REFERENCES conversations(convId) ON DELETE CASCADE,
                FOREIGN KEY (userId) REFERENCES users(userId) ON DELETE CASCADE
            );
        `
		_, err = db.Exec(sqlStmt)
		if err != nil {
			return nil, fmt.Errorf("error creating 'conversation_participants' table: %w", err)
		}
	} else if err != nil {
		// Se c'è un errore nella query (es. connessione), fallo risalire
		return nil, fmt.Errorf("error querying 'conversation_participants' table existence: %w", err)
	}

	// --------------------------------------------------------
	// TABLE MEDIA
	err = db.QueryRow(`SELECT name FROM sqlite_master WHERE type='table' AND name='media';`).Scan(&tableName)
	if errors.Is(err, sql.ErrNoRows) {
		sqlStmt := `
            CREATE TABLE media (
                mediaId TEXT NOT NULL PRIMARY KEY,
                url TEXT NOT NULL,
                filename TEXT DEFAULT NULL,
                mimeType TEXT NOT NULL,
                size INTEGER DEFAULT NULL,
                createdAt DATETIME DEFAULT CURRENT_TIMESTAMP
            );
        `
		_, err = db.Exec(sqlStmt)
		if err != nil {
			return nil, fmt.Errorf("error creating 'media' table: %w", err)
		}
	} else if err != nil {
		return nil, fmt.Errorf("error querying 'media' table existence: %w", err)
	}
	// --------------------------------------------------------
	// TABLE REACTIONS
	err = db.QueryRow(`SELECT name FROM sqlite_master WHERE type='table' AND name='reactions';`).Scan(&tableName)
	if errors.Is(err, sql.ErrNoRows) {
		sqlStmt := `
            CREATE TABLE reactions (
                messageId TEXT NOT NULL,
                userId TEXT NOT NULL,
                emoji TEXT NOT NULL,
                createdAt DATETIME DEFAULT CURRENT_TIMESTAMP,

                PRIMARY KEY (messageId, userId),
                FOREIGN KEY (messageId) REFERENCES messages(messageId) ON DELETE CASCADE,
                FOREIGN KEY (userId) REFERENCES users(userId) ON DELETE CASCADE
            );
        `
		_, err = db.Exec(sqlStmt)
		if err != nil {
			return nil, fmt.Errorf("error creating 'reactions' table: %w", err)
		}
	} else if err != nil {
		return nil, fmt.Errorf("error querying 'reactions' table existence: %w", err)
	}

	return &appdbimpl{
		c: db,
	}, nil
}

// --------------------------------------------------------
// err = db.QueryRow(`SELECT name FROM sqlite_master WHERE type='table' AND name='example_table';`).Scan(&tableName)
// if errors.Is(err, sql.ErrNoRows) {
// sqlStmt := `CREATE TABLE example_table (id INTEGER NOT NULL PRIMARY KEY, name TEXT);`
// _, err = db.Exec(sqlStmt)
// if err != nil {
// return nil, fmt.Errorf("error creating database structure: %w", err)
// }
// }
func (db *appdbimpl) Ping() error {
	return db.c.Ping()
}

/* TODO
--Tabelle--
 |Users|-> fatto
- |Conversations|-> fatto
- |ConversationParticipants|-> fatto
- Groups(?) -> in teoria fatto
- |Messages|-> fatto
- Comments
- Reactions
- Media
*/
