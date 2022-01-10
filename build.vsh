system('CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags "-s -w" -o wdm')

system('upx -9 wdm')

println("done!")
