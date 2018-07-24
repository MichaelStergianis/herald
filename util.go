package main

import "log"

// check ...
// Checks errors and upon error exits
func check(e error) {
	if e != nil {
		log.Fatalf("%v", e)
	}
}
