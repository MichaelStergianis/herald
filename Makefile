.PHONY: run

DBPATH=gitlab.stergianis.ca/michael/herald/db

run: compile
	./herald

compile-db:
	go build ${DBPATH}
	go install gitlab.stergianis.ca/michael/herald/db

compile: main.go routes.go routes_test.go compile-db
	go build -v

test:
	go test -v ./...
