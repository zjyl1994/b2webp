build: go-build compress
go-build:
	go build -ldflags="-s -w" -o b2webp .
compress:
	upx ./b2webp