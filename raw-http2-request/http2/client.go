package http2

import (
	"bufio"
	"bytes"
	"crypto/tls"
	"fmt"
	"io"
	"net/url"

	"golang.org/x/net/http2/hpack"
)

type Client struct {
	streams      map[uint32]bool
	lastStreamID uint32
}

func (c *Client) initConn(host string, port string) (*tls.Conn, *bufio.ReadWriter, error) {
	conn, err := tls.Dial("tcp", fmt.Sprintf("%s:%s", host, port), &tls.Config{
		ServerName: host,
		NextProtos: []string{"h2"},
	})
	if err != nil {
		return nil, nil, err
	}

	bufrw := bufio.NewReadWriter(bufio.NewReader(conn), bufio.NewWriter(conn))

	return conn, bufrw, nil
}

// Get sends a GET request to the given URL.
func (c *Client) Get(rawURL string) (*Response, error) {
	targetURL, err := url.Parse(rawURL)
	if err != nil {
		return nil, err
	}

	conn, bufrw, err := c.initConn(targetURL.Host, "443")
	if err != nil {
		return nil, err
	}
	defer conn.CloseWrite()

	err = writeHTTP2ConnPreface(bufrw)
	if err != nil {
		return nil, err
	}

	err = writeHeaderFrame(bufrw, 1, "GET", "https", targetURL.Host, targetURL.Path)
	if err != nil {
		return nil, err
	}

	err = bufrw.Flush()
	if err != nil {
		return nil, err
	}

	for {
		frame, err := parseFrame(bufrw)
		if err != nil {
			return nil, err
		}

		fmt.Println(frame)

		if frame.streamID == 1 && frame.hasFlag(flagEndStream) {
			break
		}
	}

	return &Response{}, nil
}

type Response struct {
}

func writeHeaderFrame(w io.Writer, streamID uint32, method, scheme, host, path string) error {
	var hb bytes.Buffer
	enc := hpack.NewEncoder(&hb)
	headerFields := []hpack.HeaderField{
		{Name: ":method", Value: method},
		{Name: ":scheme", Value: scheme},
		{Name: ":authority", Value: host},
		{Name: ":path", Value: path},
	}

	for _, hf := range headerFields {
		if err := enc.WriteField(hf); err != nil {
			return fmt.Errorf("error encoding header: %v", err)
		}
	}

	if err := writeFrame(w, frameTypeHeaders, flagEndStream|flagEndHeaders, streamID, hb.Bytes()); err != nil {
		return fmt.Errorf("error writing frame: %v", err)
	}

	return nil
}

func writeHTTP2ConnPreface(w io.Writer) error {
	// Make sure that the server supports HTTP/2
	if _, err := w.Write([]byte("PRI * HTTP/2.0\r\n\r\nSM\r\n\r\n")); err != nil {
		return fmt.Errorf("error writing HTTP/2 preamble: %v", err)
	}

	if err := writeFrame(w, frameTypeSettings, flagEmpty, connectionControlStreamID, nil); err != nil {
		return fmt.Errorf("error writing frame: %v", err)
	}

	return nil
}
