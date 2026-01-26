package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
)

const (
	addr = ":8080"
)

func main() {
	http.HandleFunc("GET /", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "index.html")
	})
	http.HandleFunc("GET /stream", streamAudio)

	log.Println("start server on", addr)
	if err := http.ListenAndServe(addr, nil); err != nil {
		log.Fatalln("error starting http server")
	}
}

// streamAudio ...
func streamAudio(w http.ResponseWriter, r *http.Request) {
	f, err := os.Open("sample.mp3")
	if err != nil {
		http.Error(w, "cannot open file", http.StatusInternalServerError)
		return
	}
	defer f.Close()

	fileStat, err := f.Stat()
	if err != nil {
		http.Error(w, fmt.Sprintf("cannot get file stat: %v", err), http.StatusInternalServerError)
		return
	}
	fileSize := fileStat.Size()

	rangeHeader := r.Header.Get("range")

	// If this is not a range request, return the whole file.
	if rangeHeader == "" {
		w.Header().Set("content-type", "audio/mpeg")
		w.Header().Set("content-length", strconv.FormatInt(fileSize, 10))
		io.Copy(w, f)
		return
	}

	rangeStart, rangeEnd, err := parseRangeHeader(rangeHeader, fileSize-1)
	if err != nil {
		http.Error(w, fmt.Sprintf("error parsing range header: %v", err), http.StatusBadRequest)
		return
	}
	if rangeStart > rangeEnd || rangeStart < 0 || rangeEnd >= fileSize {
		http.Error(w, "invalid range header", http.StatusBadRequest)
		return
	}

	_, err = f.Seek(rangeStart, 0)
	if err != nil {
		http.Error(w, fmt.Sprintf("error seeking file: %v", err), http.StatusInternalServerError)
		return
	}

	contentLength := rangeEnd - rangeStart + 1

	w.Header().Set("content-type", "audio/mpeg")
	w.Header().Set("content-length", strconv.FormatInt(contentLength, 10))
	w.Header().Set("content-range", fmt.Sprintf("bytes %d-%d/%d", rangeStart, rangeEnd, fileSize))

	// https://developer.mozilla.org/en-US/docs/Web/HTTP/Reference/Headers/Range#:~:text=If%20the%20server%20sends%20back%20ranges%2C%20it%20uses%20the%20206%20Partial%20Content%20status%20code%20for%20the%20response
	w.WriteHeader(http.StatusPartialContent)

	io.CopyN(w, f, contentLength)
}

// parseRangeHeader parses a range header value in the following format `bytes=<range-start>-<range-end>`
func parseRangeHeader(value string, defaultRangeEnd int64) (start, end int64, err error) {
	value = strings.TrimPrefix(value, "bytes=")

	rangeStart := int64(0)
	rangeEnd := defaultRangeEnd

	parts := strings.Split(value, "-")
	if parts[0] != "" {
		rangeStart, err = strconv.ParseInt(parts[0], 10, 64)
	}
	if parts[1] != "" {
		rangeEnd, err = strconv.ParseInt(parts[1], 10, 64)
	}
	if err != nil {
		return 0, 0, err
	}

	return rangeStart, rangeEnd, nil
}
