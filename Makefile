all:
	go fmt kickback.go kb_log.go
	goimports -w kickback.go kb_log.go
	go build -o kickback kickback.go kb_log.go
