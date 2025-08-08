package main

import (
	"log"
	"net/http"

	"github.com/DKE-Data/agrirouter-sdk-go/internal/tests/test_server"
	"github.com/tmaxmax/go-sse"
)

func main() {
	server := test_server.NewServer()

	r := http.NewServeMux()

	events := &sse.Server{}

	go func() {
		for event := range server.GetTestEvents() {
			eventType, err := sse.NewType(event.EventType)
			if err != nil {
				log.Printf("Error creating SSE type: %v", err)
				continue
			}
			m := &sse.Message{
				Type: eventType,
			}
			m.AppendData(event.Data)
			events.Publish(m)
			log.Println("Server sent event:", event.EventType, "with data:", event.Data)
		}
	}()

	r.Handle("/_events", events)

	strict := test_server.NewStrictHandler(server, nil)

	h := test_server.HandlerFromMux(strict, r)

	s := &http.Server{
		Handler: h,
		Addr:    "0.0.0.0:8080",
	}
	log.Fatal(s.ListenAndServe())
}
