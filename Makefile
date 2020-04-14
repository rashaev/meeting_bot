build:
	# MacOS 64-bit
	GOOS=darwin GOARCH=amd64 go build -o bin/meetingbot-macos main.go
	# Linux 64-bit
	GOOS=linux GOARCH=amd64 go build -o bin/meetingbot-linux main.go

clean:
	@rm -rf bin/

all: build