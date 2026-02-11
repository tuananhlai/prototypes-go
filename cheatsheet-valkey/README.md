# Simple Valkey Cheatsheet in Go

This project demonstrates how to perform common operations with [Valkey](https://valkey.io/) using the official Go client `github.com/valkey-io/valkey-go`.

## Prerequisites

- Go 1.22+
- A running Valkey instance

## Getting Started

### 1. Start Valkey (using Docker Compose)

Start the Valkey instance defined in `compose.yml`:

```bash
docker compose up -d
```

To stop:

```bash
docker compose down
```

Alternatively, if you want to run a one-off container without compose:

```bash
docker run -d --name valkey-server -p 6379:6379 valkey/valkey:latest
```

### 2. Run the Cheatsheet

Run the Go program to see the operations in action:

```bash
go mod tidy
go run main.go
```

## Key Concepts & Operations

### 1. Generic Key Operations

These commands work on any key, regardless of its data type.

- **SET**: Assigns a value to a key.
- **GET**: Retrieves the value stored at a key.
- **DEL**: Removes a key and its associated value.
- **EXISTS**: Checks if a key exists in the database.
- **EXPIRE**: Sets a time-to-live (TTL) for a key in seconds.
- **TTL**: Returns the remaining time-to-live of a key.

### 2. String Operations

Strings are the most basic Valkey value. Binary-safe and can contain any data.

- **INCR**: Increments the integer value of a key by one. Atomic operation.
- **DECR**: Decrements the integer value of a key by one. Atomic operation.
- **SETNX** (SET if Not eXists): Sets a key only if it does not already exist. Useful for distributed locks.

### 3. List Operations

Lists are linked lists of strings. Order is based on insertion.

- **LPUSH**: Inserts one or more values at the head (left) of the list.
- **RPUSH**: Inserts one or more values at the tail (right) of the list.
- **LPOP**: Removes and returns the first element of the list.
- **LRANGE**: Returns a range of elements. `0 -1` returns all elements.
- **LLEN**: Returns the length of the list.

### 4. Hash Operations

Hashes are maps between string fields and string values, perfect for representing objects.

- **HSET**: Sets the string value of a hash field.
- **HGET**: Gets the value of a hash field.
- **HGETALL**: Returns all fields and values in a hash.
- **HDEL**: Removes one or more fields from a hash.

### 5. Set Operations

Sets are unordered collections of unique strings.

- **SADD**: Adds one or more members to a set.
- **SMEMBERS**: Returns all members of the set.
- **SISMEMBER**: Checks if a value is a member of the set.
- **SREM**: Removes one or more members from a set.

### 6. Sorted Set (ZSet) Operations

Sorted Sets are like Sets but every member has an associated score (float) used for ordering.

- **ZADD**: Adds one or more members, or updates the score if it already exists.
- **ZRANGE**: Returns members in a specified range, sorted by score (low to high).
- **ZRANK**: Returns the rank (index) of a member, with scores ordered from low to high.

## Code Structure

The `main.go` file contains a focused function for each data structure type.

```go
func main() {
    // ... connection setup ...
    genericOperations(ctx, client)
    stringOperations(ctx, client)
    listOperations(ctx, client)
    hashOperations(ctx, client)
    setOperations(ctx, client)
    sortedSetOperations(ctx, client)
}
```
