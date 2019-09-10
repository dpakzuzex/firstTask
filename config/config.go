package config

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"time"

	jwt "github.com/dgrijalva/jwt-go"
	"github.com/gomodule/redigo/redis"
	"github.com/pkg/errors"
	"github.com/spf13/viper"
)

var (
	postgres *sql.DB
)

//GenerateJWT generating jwt
func GenerateJWT() (string, error) {
	token := jwt.New(jwt.SigningMethodHS256)

	claims := token.Claims.(jwt.MapClaims)

	claims["authorized"] = true
	claims["user"] = "user"
	claims["exp"] = time.Now().Add(time.Minute * 30).Unix()

	tokenString, err := token.SignedString([]byte(viper.GetString(`keyJWT`)))
	if err != nil {
		fmt.Errorf("Error GenerateJWT: %s", err.Error())
		return "", err
	}
	return tokenString, nil
}

//PostgresConnect Connect database postgres
func PostgresConnect() (err error) {
	postgres, err = sql.Open("postgres", fmt.Sprintf(
		"host=%s user=%s dbname=%s password=%s port=%d sslmode=%s",
		viper.GetString(`postgres.host`),
		viper.GetString(`postgres.user`),
		viper.GetString(`postgres.dbname`),
		viper.GetString(`postgres.password`),
		viper.GetInt(`postgres.port`),
		"disable",
	))
	if err != nil {
		return errors.Wrap(err, "sql open connect")
	}
	postgres.SetMaxOpenConns(viper.GetInt(`postgres.maxConnect`))
	postgres.SetMaxIdleConns(viper.GetInt(`postgres.maxConnect`))
	postgres.SetConnMaxLifetime(time.Duration(50) * time.Second)

	if err = postgres.Ping(); err != nil {
		return errors.Wrap(err, "db ping")
	}

	return nil
}

//DB ..
var DB *Pool

// Options contain settings for redis DB connect.
type Options struct {
	Host     string
	Port     int
	Password string
	DBName   int
	Timeout  time.Duration
}

// Pool wrapper of redis client
type Pool struct {
	*redis.Pool
}

// Connect database redis
func RedisConnect() (*Pool, error) {
	ro := &Options{
		Host:     viper.GetString(`redis.host`),
		Port:     viper.GetInt(`redis.port`),
		Password: viper.GetString(`redis.password`),
		DBName:   viper.GetInt(`redis.dbname`),
		Timeout:  viper.GetDuration(`redis.idleTimeout`),
	}
	return NewPool(ro)
}

func ping(conn redis.Conn) error {
	_, err := redis.String(conn.Do("PING"))
	if err != nil {
		log.Printf("ERROR: fail ping redis conn: %s", err.Error())
		os.Exit(1)
	}
	return err
}

// NewPool to database redis.
func NewPool(o *Options) (*Pool, error) {
	//create pool
	pool := redis.NewPool(func() (redis.Conn, error) {
		conn, err := redis.Dial("tcp", fmt.Sprintf("%s:%d", o.Host, o.Port), redis.DialDatabase(o.DBName))
		if err != nil {
			log.Printf("ERROR: fail init redis: %s", err.Error())
		}
		return conn, err
	}, 30000)

	// adding parameters.
	pool.MaxActive = 29500
	pool.Wait = true

	//ping
	if err := ping(pool.Get()); err != nil {
		return nil, errors.Wrap(err, "ping redis db error")
	}

	return &Pool{pool}, nil
}
