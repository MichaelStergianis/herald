.PHONY: run

DBPATH=gitlab.stergianis.ca/michael/herald/db

run: compile
	./herald

compile-db:
	go build ${DBPATH}
	go install ${DBPATH}

compile: main.go routes.go routes_test.go compile-db
	go build -v

test:
	go test -coverprofile=db/coverage.out gitlab.stergianis.ca/michael/herald/db
	go test -coverprofile=coverage.out
