package postgres

import (
	"database/sql"
	"fmt"

	// driver for postgres
	_ "github.com/lib/pq"
)

type Db struct {
	*sql.DB
}

// User shape
type User struct {
	ID         int
	Name       string
	Age        int
	Profession string
	Friendly   bool
}

// New makes a new database using the connection string
func New(connString string) (*Db, error) {
	db, err := sql.Open("postgres", connString)
	if err != nil {
		return nil, err
	}

	err = db.Ping()
	if err != nil {
		return nil, err
	}

	return &Db{db}, nil
}

// ConnString returns a connection string for the
// postgres database
func ConnString(port int, host, user, dbName string) string {
	return fmt.Sprintf(
		"host=%s port=%d user=%s password=password dbname=%s sslmode=disable",
		host, port, user, dbName,
	)
}

// GetUsersByName is clled within our users query for graphql
func (d *Db) GetUsersByName(name string) []User {
	stmt, err := d.Prepare("SELECT * FROM users WHERE name=$1")
	if err != nil {
		fmt.Println("GetUsersByName Preparation Err: ", err)
	}

	rows, err := stmt.Query(name)
	if err != nil {
		fmt.Println("GetUsersByName Query Err: ", err)
	}

	var r User

	users := []User{}

	for rows.Next() {
		err = rows.Scan(
			&r.ID,
			&r.Name,
			&r.Age,
			&r.Profession,
			&r.Friendly,
		)

		if err != nil {
			fmt.Println("Error scanning rows: ", err)
		}

		users = append(users, r)
	}

	return users
}
