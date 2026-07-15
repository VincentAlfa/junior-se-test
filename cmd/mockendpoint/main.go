package main

import (
	"flag"
	"log"
	"net/http"
)

var (
	mode       = flag.String("mode", "fail", "behavior: fail | succeed | slow | flaky")
	port       = flag.String("port", "9000", "port to listen on")
	flakyCount = 0
)

func handleRecieve(w http.ResponseWriter, r *http.Request) {
	switch *mode {
	case "succeed":
		log.Println("mockendpoint: returning 200")
		w.WriteHeader(http.StatusOK)
	case "flaky":
		flakyCount++
		if flakyCount <= 2 {
			log.Printf("mockendpoint: flaky mode, attempt %d, returning 500\n", flakyCount)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		log.Printf("mockendpoint: flaky mode, attempt %d, returning 200\n", flakyCount)
		w.WriteHeader(http.StatusOK)
	default:
		log.Printf("mockendpoint: returning 500")
		w.WriteHeader(http.StatusInternalServerError)
	}
}

func main() {
	flag.Parse()
	http.HandleFunc("/receive", handleRecieve)

	log.Printf("mockendpoint: running on :%s in %s mode\n", *port, *mode)
	err := http.ListenAndServe(":"+*port, nil)
	if err != nil {
		log.Fatal("Listen and serve", err)
	}
}
