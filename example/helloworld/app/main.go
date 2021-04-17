package main

import (
	"flag"
	"log"
	"net/http"
)

func main() {
	flagAddress := flag.String("address", ":80", "address to listen to ")
	flagGreeting := flag.String("greeting", "HELLO", "sets the greeting message")
	flag.Parse()

	greeting := []byte(*flagGreeting)

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write(greeting)
	})

	log.Fatal(http.ListenAndServe(*flagAddress, handler))
}
