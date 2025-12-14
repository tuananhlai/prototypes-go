# Login flow using a single access token

A login flow which uses a single access token with 7-day expiration stored inside an http-only cookie to authenticate users.

## Commands

Run the demo
```sh
go run .
```

Generate a private key
```sh
openssl ecparam -name prime256v1 -genkey -noout -out key.pem
```
