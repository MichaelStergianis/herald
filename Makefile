.PHONY: run

run: compile
	./herald

compile: main.go
	go install gitlab.stergianis.ca/herald/db
	go build -v

test:
	go test -v ./...
