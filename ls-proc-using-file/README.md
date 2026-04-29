```sh
➜  testprogram git:(main) ✗ ./testprogram       
temp file: /tmp/demo-2679356517.txt
pid: 35400

➜  prototypes-go git:(main) ✗ ls -l /proc/35400/fd
total 0
lrwx------ 1 vscode vscode 64 Apr 29 13:36 0 -> /dev/pts/0
lrwx------ 1 vscode vscode 64 Apr 29 13:36 1 -> /dev/pts/0
lrwx------ 1 vscode vscode 64 Apr 29 13:36 2 -> /dev/pts/0
lr-x------ 1 vscode vscode 64 Apr 29 13:36 3 -> /sys/fs/cgroup/cpu.max
lrwx------ 1 vscode vscode 64 Apr 29 13:36 4 -> /tmp/demo-2679356517.txt
lrwx------ 1 vscode vscode 64 Apr 29 13:36 5 -> 'anon_inode:[eventpoll]'
lrwx------ 1 vscode vscode 64 Apr 29 13:36 6 -> 'anon_inode:[eventfd]'
```