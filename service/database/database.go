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

	_ "github.com/mattn/go-sqlite3"
)

// AppDatabase is the high level interface for the DB
type AppDatabase interface {
	GetName() (string, error)
	SetName(name string) error
	FindUserByName(name string) (string, error)
	CreateUser(identifier string, name string) error

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
	// TABLE USERS
	err := db.QueryRow(`SELECT name FROM sqlite_master WHERE type='table' AND name='users';`).Scan(&tableName)
	if errors.Is(err, sql.ErrNoRows) {
		// La tabella 'users' non esiste, creala.
		sqlStmt := `
            CREATE TABLE "Users" (
				"userName" TEXT NOT NULL UNIQUE,
                "identifier" TEXT NOT NULL PRIMARY KEY,
                "created_at" DATETIME DEFAULT CURRENT_TIMESTAMP,
				"PFPURL" TEXT
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
	// --------------------------------------------------------
	// TABLE CONVERSATIONS
	err = db.QueryRow(`SELECT name FROM sqlite_master WHERE type='table' AND name='conversations';`).Scan(&tableName)
	if errors.Is(err, sql.ErrNoRows) {
		sqlStmt := `CREATE TABLE conversations (
                -- ID: L'identificatore principale.
                convId VARCHAR(12) NOT NULL PRIMARY KEY,
                
                -- CAMPI GRUPPO (Usati solo se kind='group')
                groupName TEXT DEFAULT NULL, 
                groupPhoto TEXT DEFAULT NULL, 
                
                -- DISCRIMINATORE: Permette di distinguere tra chat 1-a-1 e gruppi.
                kind TEXT NOT NULL DEFAULT 'private', 
                
                -- Data di creazione automatica
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
	err = db.QueryRow(`SELECT name FROM sqlite_master WHERE type='table' AND name='Messages';`).Scan(&tableName)
	if errors.Is(err, sql.ErrNoRows) {
		sqlStmt := `CREATE TABLE Messages (
				messageId VARCHAR(12) NOT NULL PRIMARY KEY,
				convId VARCHAR(12) NOT NULL,	
				senderId VARCHAR(12) NOT NULL,
				content TEXT NOT NULL,
				sentAt DATETIME DEFAULT CURRENT_TIMESTAMP,

				FOREIGN KEY (convId) REFERENCES conversations(convId) ON DELETE CASCADE,
				FOREIGN KEY (senderId) REFERENCES users(userId) ON DELETE CASCADE
			);`
		_, err = db.Exec(sqlStmt)
		if err != nil {
			return nil, fmt.Errorf("error creating 'Messages' table: %w", err)
		}
	} else if err != nil {
		// Se c'è un errore nella query (es. connessione), fallo risalire
		return nil, fmt.Errorf("error querying 'Messages' table existence: %w", err)
	}
	// --------------------------------------------------------
	// TABLE CONVERSATION_PARTICIPANTS

	err = db.QueryRow(`SELECT name FROM sqlite_master WHERE type='table' AND name='conversation_participants';`).Scan(&tableName)
	if errors.Is(err, sql.ErrNoRows) {
		sqlStmt := `
            CREATE TABLE conversation_participants (
                convId VARCHAR(12) NOT NULL,
                userId VARCHAR(12) NOT NULL,
                
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
	err = db.QueryRow(`SELECT name FROM sqlite_master WHERE type='table' AND name='example_table';`).Scan(&tableName)
	if errors.Is(err, sql.ErrNoRows) {
		sqlStmt := `CREATE TABLE example_table (id INTEGER NOT NULL PRIMARY KEY, name TEXT);`
		_, err = db.Exec(sqlStmt)
		if err != nil {
			return nil, fmt.Errorf("error creating database structure: %w", err)
		}
	}

	return &appdbimpl{
		c: db,
	}, nil
}

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
