bin/filmdetect: pkg/filmdetect/filmdetect.go go.mod main.go cmd/root.go
	go build -o bin/filmdetect .
