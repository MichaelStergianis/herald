package main

// Close ...
// close the web socket
const Close int = 400

// Text ...
// text message
const Text int = 1

// Message ...
// message passing struct
type Message struct {
	Type    int    `edn:"type"`
	Message string `edn:"message"`
}
