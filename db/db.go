package db

import (
	"log"
	"os"

	"github.com/jmoiron/sqlx"
)

var db *sqlx.DB

func Init() *sqlx.DB {
	var err error
	db, err = sqlx.Open("postgres", os.Getenv("DATABASE_URL"))
	if err != nil {
		log.Fatalln(err)
		return nil
	}

	createDBSchema(db)

	return db
}

// Schemas --------------------------------------
func createDBSchema(db *sqlx.DB) {
	db.MustExec(`
		CREATE TABLE IF NOT EXISTS sample (
			id serial NOT NULL,
			name varchar(255),
			created integer DEFAULT 0,
		
			PRIMARY KEY (id)
		);

		CREATE TABLE IF NOT EXISTS question (
			id serial NOT NULL,
			content varchar(255) NOT NULL,
			hide smallint DEFAULT 1,
			created integer DEFAULT 0,
		
			PRIMARY KEY (id)
		);

		CREATE TABLE IF NOT EXISTS answer (
			question_id integer NOT NULL,
			content varchar(255) NOT NULL,
			created integer DEFAULT 0,
		
			PRIMARY KEY (question_id)
		);

		CREATE TABLE IF NOT EXISTS cookie_id (
			id varchar(255) NOT NULL,
		
			PRIMARY KEY (id)
		);

		CREATE TABLE IF NOT EXISTS question_good (
			question_id integer REFERENCES question(id),
			cookie_id varchar(255) REFERENCES cookie_id(id),
		
			PRIMARY KEY (question_id, cookie_id)
		);
	`)
}
