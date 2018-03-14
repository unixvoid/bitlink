package main

import (
	"crypto/rand"
	"fmt"
	"time"

	"github.com/unixvoid/glogger"
	"golang.org/x/crypto/sha3"
	"gopkg.in/redis.v5"
)

func tokenGen(strSize int, client *redis.Client) string {
	// generate new token
	token := randStr(strSize, &config.Bitlink.TokenDictionary)

	// hash token
	hash := sha3.Sum512([]byte(token))
	hashstr := fmt.Sprintf("%x", hash)
	_, err := client.Get(hashstr).Result()
	for err != redis.Nil {
		glogger.Debug.Println("DEBUG :: COLLISION")
		token = randStr(strSize, &config.Bitlink.TokenDictionary)
		hash := sha3.Sum512([]byte(token))
		hashstr := fmt.Sprintf("%x", hash)
		_, err = client.Get(hashstr).Result()
		// sleep inbetween retries
		time.Sleep(time.Second * 1)
	}
	return token
}

func randStr(strSize int, dictionary *string) string {
	uberDictionary := *dictionary
	var bytes = make([]byte, strSize)
	rand.Read(bytes)
	for k, v := range bytes {
		bytes[k] = uberDictionary[v%byte(len(uberDictionary))]
	}
	return string(bytes)
}
