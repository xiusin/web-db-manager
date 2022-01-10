system('CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags "-s -w" -o web-db-manager')

system('upx -9 web-db-manager')

println("done!")
