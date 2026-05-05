package main

import "golang.org/x/sys/unix"

func main() {
	writeDataDataSparseFile("./sparse_data_data.bin")
	writeHoleHoleSparseFile("./sparse_hole_hole.bin")
	writeDataHoleSparseFile("./sparse_data_hole.bin")
	writeHoleDataSparseFile("./sparse_hole_data.bin")
}

func writeDataDataSparseFile(path string) {
	fd, err := unix.Open(path, unix.O_RDWR|unix.O_CREAT|unix.O_TRUNC, 0o644)
	if err != nil {
		panic(err)
	}
	defer unix.Close(fd)

	_, err = unix.Write(fd, []byte("hello,"))
	if err != nil {
		panic(err)
	}

	_, err = unix.Seek(fd, 15000, unix.SEEK_CUR)
	if err != nil {
		panic(err)
	}

	_, err = unix.Write(fd, []byte("world!"))
	if err != nil {
		panic(err)
	}
}

func writeHoleHoleSparseFile(path string) {
	fd, err := unix.Open(path, unix.O_RDWR|unix.O_CREAT|unix.O_TRUNC, 0o644)
	if err != nil {
		panic(err)
	}
	defer unix.Close(fd)

	_, err = unix.Seek(fd, 15000, unix.SEEK_CUR)
	if err != nil {
		panic(err)
	}

	_, err = unix.Write(fd, []byte("hello,"))
	if err != nil {
		panic(err)
	}

	_, err = unix.Seek(fd, 15000, unix.SEEK_CUR)
	if err != nil {
		panic(err)
	}

	pos, err := unix.Seek(fd, 0, unix.SEEK_CUR)
	if err != nil {
		panic(err)
	}
	if err = unix.Ftruncate(fd, pos); err != nil {
		panic(err)
	}
}

func writeDataHoleSparseFile(path string) {
	fd, err := unix.Open(path, unix.O_RDWR|unix.O_CREAT|unix.O_TRUNC, 0o644)
	if err != nil {
		panic(err)
	}
	defer unix.Close(fd)

	_, err = unix.Write(fd, []byte("hello,"))
	if err != nil {
		panic(err)
	}

	_, err = unix.Seek(fd, 15000, unix.SEEK_CUR)
	if err != nil {
		panic(err)
	}

	pos, err := unix.Seek(fd, 0, unix.SEEK_CUR)
	if err != nil {
		panic(err)
	}
	if err = unix.Ftruncate(fd, pos); err != nil {
		panic(err)
	}
}

func writeHoleDataSparseFile(path string) {
	fd, err := unix.Open(path, unix.O_RDWR|unix.O_CREAT|unix.O_TRUNC, 0o644)
	if err != nil {
		panic(err)
	}
	defer unix.Close(fd)

	_, err = unix.Seek(fd, 15000, unix.SEEK_CUR)
	if err != nil {
		panic(err)
	}

	_, err = unix.Write(fd, []byte("world!"))
	if err != nil {
		panic(err)
	}
}
