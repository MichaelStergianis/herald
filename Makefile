.PHONY: run

run: compile
	./herald

compile: main.go
	go build -v
