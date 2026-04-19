```
➜  simple-dup git:(main) ✗ go run .
fd 4: 64
newfd 5: 64
```

duplicated file descriptor shares the same offset value and open file status flags with the original fd.