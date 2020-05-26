package config

import "fmt"

type DatabaseConfig struct {
	Host     string
	Port     uint32
	User     string
	Password string
	Database string
}

func (c DatabaseConfig) ConnectionString() string {
	return fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
		c.Host, c.Port, c.User, c.Password, c.Database)
}
