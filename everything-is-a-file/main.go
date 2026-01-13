package main

import (
	"bytes"
	"fmt"
	"io"
	"net"
	"os"
	"path/filepath"
	"syscall"
)

func main() {
	// Starting input data
	inputData := "Hello, Unix! Everything is a file.\n"

	fmt.Fprintf(os.Stderr, "Starting with input: %q\n\n", inputData)

	// 1. /dev/null - The null device (discards all data)
	// A device file that acts like a black hole - not a traditional file!
	devNull, err := os.OpenFile("/dev/null", os.O_WRONLY, 0)
	if err != nil {
		panic(err)
	}
	defer devNull.Close()
	devNull.WriteString(inputData)
	fmt.Fprintf(os.Stderr, "1. Written to /dev/null (null device - data discarded)\n")

	// 2. /dev/zero - Produces infinite zeros
	// A device that generates endless zeros - definitely not a traditional file!
	devZero, err := os.OpenFile("/dev/zero", os.O_RDONLY, 0)
	if err != nil {
		panic(err)
	}
	defer devZero.Close()
	zeroData := make([]byte, len(inputData))
	devZero.Read(zeroData)
	fmt.Fprintf(os.Stderr, "2. Read zeros from /dev/zero (zero device): %x\n", zeroData)

	// 3. Named pipe (FIFO) - Created as a file but acts like a pipe
	// A file that's actually a communication channel between processes!
	fifoPath := filepath.Join(os.TempDir(), "everything-is-a-file-fifo")
	os.Remove(fifoPath) // Clean up if exists
	if err := syscall.Mkfifo(fifoPath, 0666); err != nil {
		panic(err)
	}
	defer os.Remove(fifoPath)

	// Write to FIFO in a goroutine (FIFOs block until both ends are open)
	var fifoData bytes.Buffer
	go func() {
		fifoWrite, err := os.OpenFile(fifoPath, os.O_WRONLY, 0)
		if err != nil {
			panic(err)
		}
		defer fifoWrite.Close()
		fifoWrite.WriteString(inputData)
	}()

	fifoRead, err := os.OpenFile(fifoPath, os.O_RDONLY, 0)
	if err != nil {
		panic(err)
	}
	defer fifoRead.Close()
	io.Copy(&fifoData, fifoRead)
	fmt.Fprintf(os.Stderr, "3. Passed through named pipe (FIFO): %s\n", filepath.Base(fifoPath))

	// 4. /dev/urandom - Random number generator device
	// A device that produces cryptographically secure random data!
	devUrandom, err := os.OpenFile("/dev/urandom", os.O_RDONLY, 0)
	if err != nil {
		panic(err)
	}
	defer devUrandom.Close()
	randomData := make([]byte, 8)
	devUrandom.Read(randomData)
	fmt.Fprintf(os.Stderr, "4. Read random bytes from /dev/urandom (random device): %x\n", randomData)

	// 5. Network socket - TCP connection as a file descriptor!
	// Network sockets are file descriptors that can be read/written like files!
	// This is mind-blowing: network connections are "just files"!
	listener, err := net.Listen("tcp", "127.0.0.1:0") // Listen on random port
	if err != nil {
		panic(err)
	}
	defer listener.Close()

	addr := listener.Addr().String()
	var socketData bytes.Buffer

	// Accept connection and write in a goroutine
	go func() {
		conn, err := listener.Accept()
		if err != nil {
			panic(err)
		}
		// Write data to the socket (treating it like a file)
		conn.Write(fifoData.Bytes())
		conn.Close() // Close connection after writing to signal EOF
	}()

	// Connect to the socket
	clientConn, err := net.Dial("tcp", addr)
	if err != nil {
		panic(err)
	}
	defer clientConn.Close()

	// Read from the socket (treating it like a file)
	io.Copy(&socketData, clientConn)
	fmt.Fprintf(os.Stderr, "5. Passed through network socket (TCP connection as file)\n")

	// 6. /dev/tty - Current terminal device
	// The terminal you're using is also just a file!
	devTty, err := os.OpenFile("/dev/tty", os.O_WRONLY, 0)
	if err != nil {
		// If /dev/tty not available, try /dev/console or just use stderr
		devTty, err = os.OpenFile("/dev/console", os.O_WRONLY, 0)
		if err != nil {
			// Fallback: use stderr which is also a file descriptor
			devTty = os.Stderr
			fmt.Fprintf(os.Stderr, "6. Using os.Stderr (standard error as file)\n")
		} else {
			defer devTty.Close()
			fmt.Fprintf(os.Stderr, "6. Opened /dev/console (console device as file)\n")
		}
	} else {
		defer devTty.Close()
		fmt.Fprintf(os.Stderr, "6. Opened /dev/tty (terminal as file)\n")
	}
	// Write to terminal/console/stderr
	devTty.WriteString("Written directly to terminal device!\n")

	// Finally: Write to stdout
	fmt.Fprintf(os.Stderr, "\nFinal output:\n")
	os.Stdout.WriteString(socketData.String())
}
