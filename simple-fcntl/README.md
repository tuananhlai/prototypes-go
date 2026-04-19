We can use the fcntl() F_SETFL command to modify some of the open file status flags.
The flags that can be modified are O_APPEND, O_NONBLOCK, O_NOATIME, O_ASYNC, and
O_DIRECT. Attempts to modify other flags are ignored. (Some other UNIX imple-
mentations allow fcntl() to modify other flags, such as O_SYNC.)

To modify the open file status flags, we use fcntl() to retrieve a copy of the existing
flags, then modify the bits we wish to change, and finally make a further call to fcntl()
to update the flags. 