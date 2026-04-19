Start receiver

```sh
go run .
```

Send a broadcast packet

```sh
go run ./sender
```

-> even though a packet isn't targetted specifically at the receiver, it still gets the broadcast message.

<https://en.wikipedia.org/wiki/Broadcasting_(networking)>