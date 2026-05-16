

build out the program binary, change the owner to root and set the saved id bit.
```sh

go build -o main . && sudo chown root:root main && sudo chmod u+s main
```
