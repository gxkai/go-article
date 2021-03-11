CGO_ENABLED=1 GOOS=linux GOARCH=amd64 CC="/usr/local/bin/x86_64-linux-musl-gcc"  go build -o go-api  main.go
rsync -zP ./go-api root@139.196.102.55:/go