# simple-webrtc

Minimal WebRTC data-channel demo with a tiny Go signaling server.

## Run

```bash
cd "[todo] simple-webrtc"
GOCACHE=$(pwd)/.gocache go run .
```

Open `http://localhost:8080` in 2 browser tabs:

- Tab 1: choose peer `A`, room `demo`, click `Start`
- Tab 2: choose peer `B`, room `demo`, click `Start`

When `data channel open` appears, send messages between tabs.
