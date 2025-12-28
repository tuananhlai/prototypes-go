package main

import (
	"bufio"
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/md5"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"net"
	"os"
	"strings"
	"time"
)

func main() {
	host := getenv("PGHOST", "127.0.0.1")
	port := getenv("PGPORT", "5432")
	user := getenv("PGUSER", "postgres")
	db := getenv("PGDATABASE", user)
	pass := getenv("PGPASSWORD", "postgres") // optional if trust auth

	addr := net.JoinHostPort(host, port)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	d := net.Dialer{}
	conn, err := d.DialContext(ctx, "tcp", addr)
	if err != nil {
		panic(err)
	}
	defer conn.Close()

	br := bufio.NewReader(conn)
	bw := bufio.NewWriter(conn)

	if err := sendStartup(bw, user, db); err != nil {
		panic(err)
	}
	if err := bw.Flush(); err != nil {
		panic(err)
	}

	if err := authAndWaitReady(br, bw, user, pass); err != nil {
		panic(err)
	}

	// Simple query protocol: Query message ('Q')
	if err := sendQuery(bw, "SELECT 1"); err != nil {
		panic(err)
	}
	if err := bw.Flush(); err != nil {
		panic(err)
	}

	val, err := readFirstFieldFromFirstRow(br)
	if err != nil {
		panic(err)
	}
	fmt.Println(val)

	// Terminate ('X')
	_ = sendTerminate(bw)
	_ = bw.Flush()
}

func getenv(k, def string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return def
}

// --- Wire protocol helpers (Postgres protocol v3) ---

func sendStartup(w *bufio.Writer, user, db string) error {
	// StartupMessage:
	// int32 length, int32 protocol(196608), then key\0val\0 ... \0
	var body bytes.Buffer
	_ = binary.Write(&body, binary.BigEndian, int32(196608)) // 3.0

	writeCString(&body, "user")
	writeCString(&body, user)
	writeCString(&body, "database")
	writeCString(&body, db)
	writeCString(&body, "client_encoding")
	writeCString(&body, "UTF8")
	body.WriteByte(0) // terminator

	totalLen := int32(body.Len() + 4) // includes length itself
	if err := binary.Write(w, binary.BigEndian, totalLen); err != nil {
		return err
	}
	_, err := w.Write(body.Bytes())
	return err
}

func sendPasswordMessage(w *bufio.Writer, password string) error {
	// PasswordMessage: 'p' + int32 len + password\0
	var body bytes.Buffer
	writeCString(&body, password)
	return sendTyped(w, 'p', body.Bytes())
}

func sendQuery(w *bufio.Writer, sql string) error {
	// Query: 'Q' + int32 len + sql\0
	var body bytes.Buffer
	writeCString(&body, sql)
	return sendTyped(w, 'Q', body.Bytes())
}

func sendTerminate(w *bufio.Writer) error {
	// Terminate: 'X' + int32 len (4)
	return sendTyped(w, 'X', nil)
}

func sendTyped(w *bufio.Writer, typ byte, payload []byte) error {
	if err := w.WriteByte(typ); err != nil {
		return err
	}
	// length includes itself, excludes type byte
	length := int32(4 + len(payload))
	if err := binary.Write(w, binary.BigEndian, length); err != nil {
		return err
	}
	_, err := w.Write(payload)
	return err
}

func writeCString(b *bytes.Buffer, s string) {
	b.WriteString(s)
	b.WriteByte(0)
}

type msg struct {
	typ     byte
	payload []byte
}

func readMsg(r *bufio.Reader) (msg, error) {
	typ, err := r.ReadByte()
	if err != nil {
		return msg{}, err
	}
	var n int32
	if err := binary.Read(r, binary.BigEndian, &n); err != nil {
		return msg{}, err
	}
	if n < 4 {
		return msg{}, errors.New("invalid message length")
	}
	payload := make([]byte, int(n-4))
	if _, err := io.ReadFull(r, payload); err != nil {
		return msg{}, err
	}
	return msg{typ: typ, payload: payload}, nil
}

// --- Startup/auth/ready loop ---

func authAndWaitReady(r *bufio.Reader, w *bufio.Writer, user, pass string) error {
	var scram *scramState
	for {
		m, err := readMsg(r)
		if err != nil {
			return err
		}
		switch m.typ {
		case 'R': // Authentication
			if err := handleAuth(m.payload, w, user, pass, &scram); err != nil {
				return err
			}
			if err := w.Flush(); err != nil {
				return err
			}
		case 'S': // ParameterStatus
			// ignore
		case 'K': // BackendKeyData
			// ignore
		case 'Z': // ReadyForQuery
			return nil
		case 'E': // ErrorResponse
			return fmt.Errorf("server error: %s", parseError(m.payload))
		default:
			// ignore other messages for this minimal client
		}
	}
}

func handleAuth(payload []byte, w *bufio.Writer, user, pass string, scram **scramState) error {
	if len(payload) < 4 {
		return errors.New("bad auth message")
	}
	authType := int32(binary.BigEndian.Uint32(payload[:4]))
	switch authType {
	case 0: // AuthenticationOk
		return nil
	case 3: // AuthenticationCleartextPassword
		if pass == "" {
			return errors.New("server requested cleartext password, but PGPASSWORD is empty")
		}
		return sendPasswordMessage(w, pass)
	case 5: // AuthenticationMD5Password (salt 4 bytes follows)
		if len(payload) < 8 {
			return errors.New("bad md5 auth message")
		}
		if pass == "" {
			return errors.New("server requested md5 password, but PGPASSWORD is empty")
		}
		salt := payload[4:8]
		return sendPasswordMessage(w, md5Password(pass, user, salt))
	case 10: // AuthenticationSASL
		if pass == "" {
			return errors.New("server requested sasl auth, but PGPASSWORD is empty")
		}
		mech, err := pickSCRAMMechanism(payload[4:])
		if err != nil {
			return err
		}
		state, err := initSCRAM(mech, user, pass)
		if err != nil {
			return err
		}
		*scram = state
		return sendSASLInitialResponse(w, mech, state.clientFirstMessage)
	case 11: // AuthenticationSASLContinue
		if *scram == nil {
			return errors.New("sasl continue without state")
		}
		clientFinal, err := (*scram).handleServerFirst(string(payload[4:]))
		if err != nil {
			return err
		}
		return sendSASLResponse(w, []byte(clientFinal))
	case 12: // AuthenticationSASLFinal
		if *scram == nil {
			return errors.New("sasl final without state")
		}
		if err := (*scram).handleServerFinal(string(payload[4:])); err != nil {
			return err
		}
		return nil
	default:
		return fmt.Errorf("unsupported auth method: %d", authType)
	}
}

func md5Password(password, user string, salt []byte) string {
	// md5(md5(password + user) + salt), hex, prefixed with "md5"
	h1 := md5.Sum([]byte(password + user))
	h2 := md5.Sum(append([]byte(fmt.Sprintf("%x", h1)), salt...))
	return "md5" + fmt.Sprintf("%x", h2)
}

type scramState struct {
	user               string
	pass               string
	mech               string
	clientNonce        string
	clientFirstBare    string
	clientFirstMessage string
	serverFirst        string
	authMessage        string
	saltedPassword     []byte
}

func pickSCRAMMechanism(payload []byte) (string, error) {
	// Server sends a NULL-terminated list of SASL mechanisms; we pick SCRAM-SHA-256.
	// payload is a list of null-terminated mechanism names ending with an extra 0 byte
	var mechs []string
	start := 0
	for start < len(payload) {
		end := bytes.IndexByte(payload[start:], 0)
		if end < 0 {
			break
		}
		if end == 0 {
			break
		}
		mechs = append(mechs, string(payload[start:start+end]))
		start += end + 1
	}
	for _, m := range mechs {
		if m == "SCRAM-SHA-256" {
			return m, nil
		}
	}
	return "", fmt.Errorf("server does not offer SCRAM-SHA-256 (got %v)", mechs)
}

func initSCRAM(mech, user, pass string) (*scramState, error) {
	// Create the initial client-first-message and stash fields needed for later steps.
	nonce, err := randomNonce()
	if err != nil {
		return nil, err
	}
	clientFirstBare := fmt.Sprintf("n=%s,r=%s", user, nonce)
	clientFirst := "n,," + clientFirstBare
	return &scramState{
		user:               user,
		pass:               pass,
		mech:               mech,
		clientNonce:        nonce,
		clientFirstBare:    clientFirstBare,
		clientFirstMessage: clientFirst,
	}, nil
}

func randomNonce() (string, error) {
	// SCRAM requires a client nonce; use random bytes and base64 for ASCII transport.
	var b [18]byte
	if _, err := rand.Read(b[:]); err != nil {
		return "", err
	}
	return base64.StdEncoding.EncodeToString(b[:]), nil
}

func sendSASLInitialResponse(w *bufio.Writer, mech string, initial string) error {
	// SASLInitialResponse uses type 'p' with mechanism name + initial response length.
	var body bytes.Buffer
	writeCString(&body, mech)
	if err := binary.Write(&body, binary.BigEndian, int32(len(initial))); err != nil {
		return err
	}
	body.WriteString(initial)
	return sendTyped(w, 'p', body.Bytes())
}

func sendSASLResponse(w *bufio.Writer, payload []byte) error {
	// SASLResponse reuses the password message type with the raw SCRAM payload.
	return sendTyped(w, 'p', payload)
}

func (s *scramState) handleServerFirst(serverFirst string) (string, error) {
	// Parse server-first-message, derive keys, and build client-final-message with proof.
	s.serverFirst = serverFirst
	attrs := parseSCRAMAttrs(serverFirst)
	serverNonce, ok := attrs["r"]
	if !ok || !strings.HasPrefix(serverNonce, s.clientNonce) {
		return "", errors.New("invalid server nonce")
	}
	saltB64, ok := attrs["s"]
	if !ok {
		return "", errors.New("missing salt in server-first-message")
	}
	iterStr, ok := attrs["i"]
	if !ok {
		return "", errors.New("missing iteration count in server-first-message")
	}
	salt, err := base64.StdEncoding.DecodeString(saltB64)
	if err != nil {
		return "", err
	}
	iter, err := parsePositiveInt(iterStr)
	if err != nil {
		return "", err
	}

	s.saltedPassword = pbkdf2SHA256([]byte(s.pass), salt, iter, 32)

	clientFinalNoProof := "c=biws,r=" + serverNonce
	s.authMessage = s.clientFirstBare + "," + serverFirst + "," + clientFinalNoProof

	clientKey := hmacSHA256(s.saltedPassword, []byte("Client Key"))
	storedKey := sha256.Sum256(clientKey)
	clientSignature := hmacSHA256(storedKey[:], []byte(s.authMessage))
	clientProof := xorBytes(clientKey, clientSignature)

	proofB64 := base64.StdEncoding.EncodeToString(clientProof)
	return clientFinalNoProof + ",p=" + proofB64, nil
}

func (s *scramState) handleServerFinal(serverFinal string) error {
	// Verify server signature (v=) to complete SCRAM authentication.
	attrs := parseSCRAMAttrs(serverFinal)
	serverSigB64, ok := attrs["v"]
	if !ok {
		return errors.New("missing server signature in server-final-message")
	}
	serverSig, err := base64.StdEncoding.DecodeString(serverSigB64)
	if err != nil {
		return err
	}
	serverKey := hmacSHA256(s.saltedPassword, []byte("Server Key"))
	expected := hmacSHA256(serverKey, []byte(s.authMessage))
	if !hmac.Equal(serverSig, expected) {
		return errors.New("server signature mismatch")
	}
	return nil
}

func parseSCRAMAttrs(s string) map[string]string {
	// SCRAM messages are comma-separated k=v pairs (single-char keys).
	out := make(map[string]string)
	parts := strings.Split(s, ",")
	for _, p := range parts {
		if len(p) < 3 || p[1] != '=' {
			continue
		}
		out[p[:1]] = p[2:]
	}
	return out
}

func parsePositiveInt(s string) (int, error) {
	// Strict ASCII integer parse to avoid strconv for this small payload.
	var n int
	for i := 0; i < len(s); i++ {
		ch := s[i]
		if ch < '0' || ch > '9' {
			return 0, errors.New("invalid iteration count")
		}
		n = n*10 + int(ch-'0')
	}
	if n <= 0 {
		return 0, errors.New("invalid iteration count")
	}
	return n, nil
}

func hmacSHA256(key, data []byte) []byte {
	// Small helper to keep SCRAM computations readable.
	h := hmac.New(sha256.New, key)
	h.Write(data)
	return h.Sum(nil)
}

func pbkdf2SHA256(password, salt []byte, iter, keyLen int) []byte {
	// Minimal PBKDF2 implementation (RFC 2898) for SCRAM salted password.
	hLen := 32
	numBlocks := (keyLen + hLen - 1) / hLen
	var out []byte
	for i := 1; i <= numBlocks; i++ {
		block := pbkdf2Block(password, salt, iter, i)
		out = append(out, block...)
	}
	return out[:keyLen]
}

func pbkdf2Block(password, salt []byte, iter, blockIndex int) []byte {
	// Compute a single PBKDF2 block (U1 xor U2 xor ...).
	u := hmacSHA256(password, append(salt, byte(blockIndex>>24), byte(blockIndex>>16), byte(blockIndex>>8), byte(blockIndex)))
	out := make([]byte, len(u))
	copy(out, u)
	for i := 1; i < iter; i++ {
		u = hmacSHA256(password, u)
		for j := 0; j < len(out); j++ {
			out[j] ^= u[j]
		}
	}
	return out
}

func xorBytes(a, b []byte) []byte {
	// XOR helper for client proof computation.
	n := len(a)
	if len(b) < n {
		n = len(b)
	}
	out := make([]byte, n)
	for i := 0; i < n; i++ {
		out[i] = a[i] ^ b[i]
	}
	return out
}

// --- Query result parsing (very minimal) ---

func readFirstFieldFromFirstRow(r *bufio.Reader) (string, error) {
	for {
		m, err := readMsg(r)
		if err != nil {
			return "", err
		}
		switch m.typ {
		case 'T': // RowDescription
			// ignore (we'll just read the first column)
		case 'D': // DataRow
			return parseFirstField(m.payload)
		case 'C': // CommandComplete
			// ignore
		case 'Z': // ReadyForQuery
			return "", errors.New("no rows returned")
		case 'E': // ErrorResponse
			return "", fmt.Errorf("query error: %s", parseError(m.payload))
		default:
			// ignore
		}
	}
}

func parseFirstField(payload []byte) (string, error) {
	// DataRow: int16 numCols, then for each col: int32 len (-1 NULL), then bytes
	if len(payload) < 2 {
		return "", errors.New("bad DataRow")
	}
	ncols := int(binary.BigEndian.Uint16(payload[:2]))
	if ncols < 1 {
		return "", errors.New("no columns")
	}
	i := 2
	if len(payload) < i+4 {
		return "", errors.New("bad DataRow (len)")
	}
	l := int(int32(binary.BigEndian.Uint32(payload[i : i+4])))
	i += 4
	if l == -1 {
		return "NULL", nil
	}
	if l < 0 || len(payload) < i+l {
		return "", errors.New("bad DataRow (field size)")
	}
	return string(payload[i : i+l]), nil
}

// ErrorResponse is a sequence of fields: (byte code)(cstring msg)... ending with 0
func parseError(payload []byte) string {
	var (
		out []string
		i   int = 0
	)
	for i < len(payload) && payload[i] != 0 {
		code := payload[i]
		i++
		j := bytes.IndexByte(payload[i:], 0)
		if j < 0 {
			break
		}
		val := string(payload[i : i+j])
		i += j + 1
		// Common useful fields: 'S' severity, 'M' message, 'C' SQLSTATE
		switch code {
		case 'S', 'M', 'C':
			out = append(out, fmt.Sprintf("%c=%s", code, val))
		}
	}
	if len(out) == 0 {
		return "unknown error"
	}
	return stringsJoin(out, ", ")
}

func stringsJoin(a []string, sep string) string {
	if len(a) == 0 {
		return ""
	}
	if len(a) == 1 {
		return a[0]
	}
	var b bytes.Buffer
	b.WriteString(a[0])
	for i := 1; i < len(a); i++ {
		b.WriteString(sep)
		b.WriteString(a[i])
	}
	return b.String()
}
