.PHONY: run

DBPATH=gitlab.stergianis.ca/michael/warbler/db

run: compile
	./warbler

compile-db:
	go build ${DBPATH}
	go install ${DBPATH}

compile: main.go routes.go routes_test.go compile-db
	go build -v
	make -C frontend/ compile

test:
	go test -coverprofile=db/coverage.out gitlab.stergianis.ca/michael/warbler/db
	go test -coverprofile=coverage.out

clean:
	rm warbler
	make -C frontend/ clean
