# CGO_ENABLED Analysis

This experiment demonstrates the effect of the `CGO_ENABLED` environment variable on a simple Go program that uses the `net` package.

## Build Commands

We build two binaries from `main.go`, one with CGO enabled and one with CGO disabled:

```bash
CGO_ENABLED=1 go build -o bin_cgo_1 main.go
CGO_ENABLED=0 go build -o bin_cgo_0 main.go
```

## 1. `ldd` (List Dynamic Dependencies)

The `ldd` command shows which shared libraries are required by each executable.

**Command:**
```bash
ldd bin_cgo_1
ldd bin_cgo_0
```

**Results:**
- `bin_cgo_1`:
  ```
	linux-vdso.so.1 (0x0000ffffbf00c000)
	libc.so.6 => /lib/aarch64-linux-gnu/libc.so.6 (0x0000ffffbee20000)
	/lib/ld-linux-aarch64.so.1 (0x0000ffffbefd0000)
  ```
  *Conclusion: Dynamically links to the C standard library (`libc.so.6`).*

- `bin_cgo_0`:
  ```
	not a dynamic executable
  ```
  *Conclusion: No dynamic dependencies. It's a statically linked binary.*

## 2. `file` (Determine File Type)

The `file` command provides details about the binary structure.

**Command:**
```bash
file bin_cgo_1
file bin_cgo_0
```

**Results:**
- `bin_cgo_1`: `ELF 64-bit LSB executable, ARM aarch64, version 1 (SYSV), dynamically linked, interpreter /lib/ld-linux-aarch64.so.1, ...`
- `bin_cgo_0`: `ELF 64-bit LSB executable, ARM aarch64, version 1 (SYSV), statically linked, ...`

*Conclusion: The `file` utility confirms that `bin_cgo_1` is dynamically linked while `bin_cgo_0` is statically linked.*

## 3. `strace` (Trace System Calls and Signals)

We trace the execution to observe which files (like `.so` files or `/etc/resolv.conf`) are opened during runtime.

**Command:**
```bash
strace ./bin_cgo_1 2>&1 | grep -E '\.so|resolv\.conf'
strace ./bin_cgo_0 2>&1 | grep -E '\.so|resolv\.conf'
```

**Results:**
- `bin_cgo_1`:
  ```
  faccessat(AT_FDCWD, "/etc/ld.so.preload", R_OK) = -1 ENOENT (No such file or directory)
  openat(AT_FDCWD, "/etc/ld.so.cache", O_RDONLY|O_CLOEXEC) = 3
  openat(AT_FDCWD, "/lib/aarch64-linux-gnu/libc.so.6", O_RDONLY|O_CLOEXEC) = 3
  openat(AT_FDCWD, "/etc/resolv.conf", O_RDONLY|O_CLOEXEC) = 4
  ```
  *Conclusion: At runtime, the OS loads `libc.so.6` from the file system. It then reads `/etc/resolv.conf` to perform DNS resolution.*

- `bin_cgo_0`:
  ```
  openat(AT_FDCWD, "/etc/resolv.conf", O_RDONLY|O_CLOEXEC) = 4
  ```
  *Conclusion: Does not attempt to load any shared object (`.so`) files since it is statically linked. It uses Go's pure-Go DNS resolver, which still needs to read `/etc/resolv.conf`.*

## Summary

Setting `CGO_ENABLED=1` causes the Go toolchain to dynamically link to the host system's C library (e.g., for `net` operations). Setting `CGO_ENABLED=0` forces Go to use its pure-Go implementations, resulting in a fully statically linked, portable binary with no external C library dependencies.
