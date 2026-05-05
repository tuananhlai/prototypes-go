Write a program like cp that, when used to copy a regular file that contains holes
(sequences of null bytes), also creates corresponding holes in the target file.

```sh
# build test sparse files.
go run ./sparsefiles

# copy a sample test file
go run . sparse_data_hole.bin out

# compare copied file's content
cmp sparse_data_hole.bin out
```