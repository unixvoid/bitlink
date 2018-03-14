package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/gorilla/mux"
	"github.com/unixvoid/glogger"
	"gopkg.in/gcfg.v1"
	"gopkg.in/redis.v5"
)

type Config struct {
	Bitlink struct {
		Loglevel        string
		Port            int
		TokenSize       int
		TokenDictionary string
		BootstrapDelay  time.Duration
	}

	Redis struct {
		Host     string
		Password string
	}
}

var (
	config = Config{}
)

func main() {
	// init config file and logger
	readConf()
	initLogger()

	// init redis connection
	// allow the bootstrap delay time if needed
	// this allows redis to start before the app connects
	// valuable when deploying in a container

	redisClient, redisErr := initRedisConnection()
	if redisErr != nil {
		glogger.Debug.Printf("redis connection cannot be made, trying again in %s second(s)\n", config.Bitlink.BootstrapDelay*time.Second)
		time.Sleep(config.Bitlink.BootstrapDelay * time.Second)
		redisClient, redisErr = initRedisConnection()
		if redisErr != nil {
			glogger.Error.Println("redis connection cannot be made, exiting.")
			panic(redisErr)
		}
	} else {
		glogger.Debug.Println("connection to redis succeeded.")
	}

	router := mux.NewRouter().StrictSlash(true)
	router.HandleFunc("/compress", func(w http.ResponseWriter, r *http.Request) {
		linkcompressor(w, r, redisClient)
	})
	router.HandleFunc("/{dataId}", func(w http.ResponseWriter, r *http.Request) {
		linkhandler(w, r, redisClient)
	}).Methods("GET")

	glogger.Info.Println("started server on", config.Bitlink.Port)
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", config.Bitlink.Port), router))
}

func readConf() {
	// init config file
	err := gcfg.ReadFileInto(&config, "config.gcfg")
	if err != nil {
		panic(fmt.Sprintf("Could not load config.gcfg, error: %s\n", err))
	}
}

func initLogger() {
	// init logger
	if config.Bitlink.Loglevel == "debug" {
		glogger.LogInit(os.Stdout, os.Stdout, os.Stdout, os.Stderr)
	} else if config.Bitlink.Loglevel == "cluster" {
		glogger.LogInit(os.Stdout, os.Stdout, ioutil.Discard, os.Stderr)
	} else if config.Bitlink.Loglevel == "info" {
		glogger.LogInit(os.Stdout, ioutil.Discard, ioutil.Discard, os.Stderr)
	} else {
		glogger.LogInit(ioutil.Discard, ioutil.Discard, ioutil.Discard, os.Stderr)
	}
}

func initRedisConnection() (*redis.Client, error) {
	// init redis connection
	redisClient := redis.NewClient(&redis.Options{
		Addr:     config.Redis.Host,
		Password: config.Redis.Password,
		DB:       0,
	})

	_, redisErr := redisClient.Ping().Result()
	return redisClient, redisErr
}
