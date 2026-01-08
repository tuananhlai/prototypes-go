package main

import (
	"fmt"
	"io"
	"log"
	"os"
	"strconv"
	"strings"
	"sync"
)

func main() {
	id, err := NewAtomicID()
	if err != nil {
		log.Fatal(err)
	}
	defer func() {
		if err := id.Close(); err != nil {
			log.Fatal(err)
		}
	}()

	for range 7 {
		value, err := id.Generate()
		if err != nil {
			log.Fatal(err)
		}
		fmt.Println(value)
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

	startID, err := getStartID(file, flushFrequency)
	if err != nil {
		return nil, err
	}
	// persist the start id immediately in case the process crashed before
	// the next flush.
	if err := writeID(file, startID); err != nil {
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

func (ai *AtomicID) Close() error {
	return ai.file.Close()
}

// flushID writes the last generated ID to a file.
func (ai *AtomicID) flushID() error {
	return writeID(ai.file, ai.lastGeneratedID)
}

func writeID(file *os.File, id uint64) error {
	if _, err := file.Seek(0, io.SeekStart); err != nil {
		return err
	}
	if err := file.Truncate(0); err != nil {
		return err
	}
	if _, err := file.WriteString(strconv.FormatUint(id, 10)); err != nil {
		return err
	}
	return file.Sync()
}

// Read the last generated ID from the given file, and determine the next safe ID to generate
// based on the flush frequency.
func getStartID(file *os.File, flushFrequency uint64) (uint64, error) {
	if _, err := file.Seek(0, io.SeekStart); err != nil {
		return 0, err
	}
	lastFlushedIDBytes, err := io.ReadAll(file)
	if err != nil {
		return 0, err
	}
	trimmed := strings.TrimSpace(string(lastFlushedIDBytes))
	if trimmed == "" {
		return 0, nil
	}
	lastFlushedID, err := strconv.ParseUint(trimmed, 10, 64)
	if err != nil {
		return 0, err
	}

	return lastFlushedID + 2*flushFrequency, nil
}
