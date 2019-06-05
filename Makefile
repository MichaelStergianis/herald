.PHONY: run

DBPATH=gitlab.stergianis.ca/michael/warbler/db

run: compile compile-frontend
	./warbler

compile-db:
	go build ${DBPATH}
	go install ${DBPATH}

compile: main.go routes.go routes_test.go compile-db
	go build -v

compile-frontend:
	make -C frontend/ compile

test:
	go test -coverprofile=db/coverage.out ${DBPATH}
	go test -coverprofile=coverage.out

clean:
	rm warbler
	make -C frontend/ clean
