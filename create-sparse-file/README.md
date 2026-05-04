Create a small file (12 bytes) with a large hole in the middle (15000 bytes).

```sh
➜  [todo] cp-include-holes git:(main) ✗ go run .                      
➜  [todo] cp-include-holes git:(main) ✗ du -h test.txt
8.0K    test.txt
➜  [todo] cp-include-holes git:(main) ✗ du -h --apparent-size test.txt
15K     test.txt
```