package main

import (
	"fmt"
	"io"
	"log"
	"os"
	"strconv"
	"sync"
)

func main() {
	id, err := NewAtomicID()
	if err != nil {
		log.Fatal(err)
	}

	for range 7 {
		fmt.Println(id.Generate())
	}
}

// AtomicID generates increasing integer IDs across multiple threads. It can also
// survive a process crash or machine restart.
type AtomicID struct {
	lastGeneratedID uint64
	flushFrequency  uint64
	file            *os.File
	mu              sync.Mutex
}

func NewAtomicID() (*AtomicID, error) {
	file, err := os.OpenFile("atomic_id", os.O_RDWR|os.O_CREATE, 0644)
	if err != nil {
		return nil, err
	}

	flushFrequency := uint64(10)

	// TODO: immediately write the start ID to the file.
	startID, err := getStartID(file, flushFrequency)
	if err != nil {
		return nil, err
	}

	return &AtomicID{
		file:            file,
		flushFrequency:  flushFrequency,
		lastGeneratedID: startID,
	}, nil
}

// Generate returns a new ID and occasionally writes the last
// generated ID to disk.
func (ai *AtomicID) Generate() (uint64, error) {
	ai.mu.Lock()
	defer ai.mu.Unlock()

	newID := ai.lastGeneratedID + 1
	ai.lastGeneratedID = newID

	if newID%ai.flushFrequency == 0 {
		if err := ai.flushID(); err != nil {
			return 0, err
		}
	}

	return newID, nil
}

// flushID writes the last generated ID to a file.
func (ai *AtomicID) flushID() error {
	if _, err := ai.file.Seek(0, 0); err != nil {
		return err
	}
	if err := ai.file.Truncate(0); err != nil {
		return err
	}
	_, err := ai.file.WriteString(strconv.FormatUint(ai.lastGeneratedID, 10))
	if err != nil {
		return err
	}

	return ai.file.Sync()
}

// Read the last generated ID from the given file, and determine the next safe ID to generate
// based on the flush frequency.
func getStartID(file *os.File, flushFrequency uint64) (uint64, error) {
	if _, err := file.Seek(0, 0); err != nil {
		return 0, err
	}
	lastFlushedIDBytes, err := io.ReadAll(file)
	if err != nil {
		return 0, err
	}
	if len(lastFlushedIDBytes) == 0 {
		return 0, nil
	}
	lastFlushedID, err := strconv.ParseUint(string(lastFlushedIDBytes), 10, 64)
	if err != nil {
		return 0, err
	}

	return lastFlushedID + 2*flushFrequency, nil
}
