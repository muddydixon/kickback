all:
	go fmt index.go
	goimports -w index.go
	GOOS=linux GOARCH=amd64 go build -o kickback index.go
