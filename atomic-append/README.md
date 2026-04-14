```sh
➜  atomic-append git:(main) ✗ ./atomic-append f1 1000000 & ./atomic-append f1 1000000
➜  atomic-append git:(main) ✗ ./atomic-append f2 1000000 x & ./atomic-append f2 1000000 x
➜  atomic-append git:(main) ✗ ls -l
total 5084
-rwxr-xr-x 1 vscode vscode 2042791 Apr 14 13:28 atomic-append
-rw------- 1 vscode vscode 2000000 Apr 14 13:29 f1
-rw------- 1 vscode vscode 1020413 Apr 14 13:30 f2
-rw-r--r-- 1 vscode vscode      99 Apr 14 13:14 go.mod
-rw-r--r-- 1 vscode vscode     153 Apr 14 13:14 go.sum
-rw-r--r-- 1 vscode vscode    1155 Apr 14 13:28 main.go
```

`f2` is way smaller than `f1` in size due to race condition. When writing `f2`, the program does 2 separate `write` and `seek` operations, so two concurrent programs may overwrite each other's data.