Create a small file (12 bytes) with a large hole in the middle (15000 bytes).

```sh
➜  create-sparse-file git:(main) ✗ go run .                      
➜  create-sparse-file git:(main) ✗ du -h test.txt
8.0K    test.txt
➜  create-sparse-file git:(main) ✗ du -h --apparent-size test.txt
15K     test.txt
```

If we had written 15000 null bytes instead.

```sh
➜  create-sparse-file git:(main) ✗ go run .                      
➜  create-sparse-file git:(main) ✗ du -h test.txt                
16K     test.txt
➜  create-sparse-file git:(main) ✗ du -h --apparent-size test.txt
15K     test.txt
```

When a file with holes are tracked by Git (via `git add`), the holes are replaced with actual null bytes. **In order to see the effect of file holes, we need to rerun `main.go` to create a new file that's not tracked by Git.**