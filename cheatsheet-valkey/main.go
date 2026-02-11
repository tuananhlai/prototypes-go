package main

import (
	"context"
	"fmt"
	"log"

	"github.com/valkey-io/valkey-go"
)

// Valkey Cheatsheet in Go
// This example uses the official valkey-go client: github.com/valkey-io/valkey-go

func main() {
	// 1. Connection
	// Connect to a local Valkey instance.
	// You can also use valkey.NewClient(valkey.ClientOption{InitAddress: []string{"127.0.0.1:6379"}})
	client, err := valkey.NewClient(valkey.ClientOption{
		InitAddress: []string{"127.0.0.1:6379"},
	})
	if err != nil {
		log.Fatalf("failed to connect to valkey: %v", err)
	}
	defer client.Close()

	ctx := context.Background()

	// Ensure connection is clean for demo purposes (Optional)
	// client.Do(ctx, client.B().Flushall().Build())

	fmt.Println("--- Generic Key Operations ---")
	genericOperations(ctx, client)

	fmt.Println("\n--- String Operations ---")
	stringOperations(ctx, client)

	fmt.Println("\n--- List Operations ---")
	listOperations(ctx, client)

	fmt.Println("\n--- Hash Operations ---")
	hashOperations(ctx, client)

	fmt.Println("\n--- Set Operations ---")
	setOperations(ctx, client)

	fmt.Println("\n--- Sorted Set (ZSet) Operations ---")
	sortedSetOperations(ctx, client)
}

func genericOperations(ctx context.Context, client valkey.Client) {
	key := "generic_key"

	// SET a key (helper for generic ops)
	client.Do(ctx, client.B().Set().Key(key).Value("test_value").Build())

	// EXISTS: Check if a key exists
	exists, _ := client.Do(ctx, client.B().Exists().Key(key).Build()).AsInt64()
	fmt.Printf("EXISTS %s: %v\n", key, exists > 0)

	// EXPIRE: Set a TTL on a key (10 seconds)
	client.Do(ctx, client.B().Expire().Key(key).Seconds(10).Build())

	// TTL: Get remaining time to live
	ttl, _ := client.Do(ctx, client.B().Ttl().Key(key).Build()).AsInt64()
	fmt.Printf("TTL %s: %d seconds\n", key, ttl)

	// DEL: Delete a key
	client.Do(ctx, client.B().Del().Key(key).Build())
	fmt.Printf("DEL %s\n", key)
}

func stringOperations(ctx context.Context, client valkey.Client) {
	key := "string_key"

	// SET: Set a string value
	client.Do(ctx, client.B().Set().Key(key).Value("hello valkey").Build())
	fmt.Printf("SET %s 'hello valkey'\n", key)

	// GET: Get a string value
	val, _ := client.Do(ctx, client.B().Get().Key(key).Build()).ToString()
	fmt.Printf("GET %s: %s\n", key, val)

	// INCR: Increment a number
	numKey := "counter_key"
	client.Do(ctx, client.B().Set().Key(numKey).Value("10").Build())
	newVal, _ := client.Do(ctx, client.B().Incr().Key(numKey).Build()).AsInt64()
	fmt.Printf("INCR %s: %d\n", numKey, newVal) // 11

	// DECR: Decrement a number
	newVal, _ = client.Do(ctx, client.B().Decr().Key(numKey).Build()).AsInt64()
	fmt.Printf("DECR %s: %d\n", numKey, newVal) // 10

	// SETNX: Set only if not exists
	setnx, _ := client.Do(ctx, client.B().Setnx().Key(key).Value("should not work").Build()).AsInt64()
	fmt.Printf("SETNX %s: %v (0=failed, 1=set)\n", key, setnx)
}

func listOperations(ctx context.Context, client valkey.Client) {
	key := "list_key"
	client.Do(ctx, client.B().Del().Key(key).Build()) // Clean up

	// LPUSH: Push to head (left)
	client.Do(ctx, client.B().Lpush().Key(key).Element("world", "hello").Build())
	fmt.Printf("LPUSH %s 'world', 'hello'\n", key)

	// RPUSH: Push to tail (right)
	client.Do(ctx, client.B().Rpush().Key(key).Element("!").Build())
	fmt.Printf("RPUSH %s '!'\n", key)

	// LRANGE: Get range of elements (0 -1 means all)
	items, _ := client.Do(ctx, client.B().Lrange().Key(key).Start(0).Stop(-1).Build()).AsStrSlice()
	fmt.Printf("LRANGE %s 0 -1: %v\n", key, items)

	// LPOP: Remove and return first element
	pop, _ := client.Do(ctx, client.B().Lpop().Key(key).Build()).ToString()
	fmt.Printf("LPOP %s: %s\n", key, pop)

	// LLEN: Get length
	len, _ := client.Do(ctx, client.B().Llen().Key(key).Build()).AsInt64()
	fmt.Printf("LLEN %s: %d\n", key, len)
}

func hashOperations(ctx context.Context, client valkey.Client) {
	key := "user:1001"

	// HSET: Set hash fields
	client.Do(ctx, client.B().Hset().Key(key).FieldValue().FieldValue("name", "Alice").FieldValue("role", "admin").Build())
	fmt.Printf("HSET %s name='Alice' role='admin'\n", key)

	// HGET: Get a specific field
	name, _ := client.Do(ctx, client.B().Hget().Key(key).Field("name").Build()).ToString()
	fmt.Printf("HGET %s name: %s\n", key, name)

	// HGETALL: Get all fields and values
	all, _ := client.Do(ctx, client.B().Hgetall().Key(key).Build()).AsStrMap()
	fmt.Printf("HGETALL %s: %v\n", key, all)

	// HDEL: Delete a field
	client.Do(ctx, client.B().Hdel().Key(key).Field("role").Build())
	fmt.Printf("HDEL %s role\n", key)
}

func setOperations(ctx context.Context, client valkey.Client) {
	key := "set_key"
	client.Do(ctx, client.B().Del().Key(key).Build()) // Clean up

	// SADD: Add members
	client.Do(ctx, client.B().Sadd().Key(key).Member("apple", "banana", "cherry").Build())
	fmt.Printf("SADD %s apple banana cherry\n", key)

	// SMEMBERS: Get all members
	members, _ := client.Do(ctx, client.B().Smembers().Key(key).Build()).AsStrSlice()
	fmt.Printf("SMEMBERS %s: %v\n", key, members)

	// SISMEMBER: Check membership
	isMember, _ := client.Do(ctx, client.B().Sismember().Key(key).Member("banana").Build()).AsInt64()
	fmt.Printf("SISMEMBER %s banana: %v\n", key, isMember > 0)

	// SREM: Remove member
	client.Do(ctx, client.B().Srem().Key(key).Member("apple").Build())
	fmt.Printf("SREM %s apple\n", key)
}

func sortedSetOperations(ctx context.Context, client valkey.Client) {
	key := "leaderboard"
	client.Do(ctx, client.B().Del().Key(key).Build()) // Clean up

	// ZADD: Add members with scores
	client.Do(ctx, client.B().Zadd().Key(key).ScoreMember().ScoreMember(100, "Player1").ScoreMember(
		200, "Player2").ScoreMember(50, "Player3").Build())
	fmt.Printf("ZADD %s 100 Player1, 200 Player2, 50 Player3\n", key)

	// ZRANGE: Get range by index (low to high score)
	// For high to low, use ZREVRANGE (or ZRANGE with REV argument in newer Valkey/Redis)
	// valkey-go might use ZRange with ZRANGE options, or ZRevRange command builder if available locally or server version dependent.
	// Standard ZRANGE returns low to high.
	top, _ := client.Do(ctx, client.B().Zrange().Key(key).Min("0").Max("-1").Build()).AsStrSlice()
	fmt.Printf("ZRANGE %s 0 -1 (Low to High): %v\n", key, top)

	// ZRANGEBYSCORE: Get members within score range
	// Note: explicit min/max strings required for bounds like "(50", "+inf" etc if needed, but here simple numbers.
	// client.B().Zrangebyscore().Key(key).Min("50").Max("200").Build()
	// However, valkey-go has specific builders. Let's stick to simple ZRange which is often preferred in RESP3/newer servers,
	// but ZRangeByScore is classic.
	byScore, _ := client.Do(ctx, client.B().Zrangebyscore().Key(key).Min("50").Max("150").Build()).AsStrSlice()
	fmt.Printf("ZRANGEBYSCORE %s 50 150: %v\n", key, byScore)

	// ZRANK: Get rank (index)
	rank, _ := client.Do(ctx, client.B().Zrank().Key(key).Member("Player2").Build()).AsInt64()
	fmt.Printf("ZRANK %s Player2: %d\n", key, rank)
}
