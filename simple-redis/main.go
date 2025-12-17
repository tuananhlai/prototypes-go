package main

import (
	"context"
	"log"

	"github.com/redis/go-redis/v9"
)

func main() {
	rdb := redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
	})

	defer rdb.Close()

	sampleStrKey := "mystring"
	sampleValue := "hello, world"
	ctx := context.Background()
	err := rdb.Set(ctx, sampleStrKey, sampleValue, 0).Err()
	if err != nil {
		log.Fatalf("SET %s failed: %v", sampleStrKey, err)
	}

	val, err := rdb.Get(ctx, sampleStrKey).Result()
	if err != nil {
		log.Fatalf("GET %s failed: %v", sampleStrKey, err)
	}
	log.Printf("GET %s: %s\n", sampleStrKey, val)

	sampleHashKey := "myhash"

	err = rdb.HSet(ctx, sampleHashKey, "field1", "value1").Err()
	if err != nil {
		log.Fatalf("HSET %s failed: %v", sampleHashKey, err)
	}

	err = rdb.HSet(ctx, sampleHashKey, "field2", "value2").Err()
	if err != nil {
		log.Fatalf("HSET %s (field2) failed: %v", sampleHashKey, err)
	}

	hashValue, err := rdb.HGetAll(ctx, sampleHashKey).Result()
	if err != nil {
		log.Fatalf("HGETALL %s: %v", sampleHashKey, err)
	}
	log.Printf("HGETALL %s: %v\n", sampleHashKey, hashValue)

	sampleListKey := "mylist"
	err = rdb.RPush(ctx, sampleListKey, "elem1", "elem2", "elem3").Err()
	if err != nil {
		log.Fatalf("RPUSH %s failed: %v", sampleListKey, err)
	}

	// Why is `listValue` []string?
	listValue, err := rdb.LRange(ctx, sampleListKey, 0, -1).Result()
	if err != nil {
		log.Fatalf("LRANGE %s failed: %v", sampleListKey, err)
	}
	log.Printf("LRANGE %s: %v\n", sampleListKey, listValue)

	sampleSetKey := "myset"
	err = rdb.SAdd(ctx, sampleSetKey, "member1", "member2", "member1").Err()
	if err != nil {
		log.Fatalf("SADD %s failed: %v", sampleSetKey, err)
	}

	setValue, err := rdb.SMembers(ctx, sampleSetKey).Result()
	if err != nil {
		log.Fatalf("SMEMBERS %s failed: %v", sampleSetKey, err)
	}
	log.Printf("SMEMBERS %s: %v", sampleSetKey, setValue)
}
