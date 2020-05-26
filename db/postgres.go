package db

import "database/sql"

func LocalConnection(connectionString string) (*sql.DB, error) {
	postgres, err := sql.Open("postgres", connectionString)
	if err != nil {
		return nil, err
	}

	if err := postgres.Ping(); err != nil {
		return nil, err
	}

	return postgres, err
}
