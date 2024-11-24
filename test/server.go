package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"
)

func main() {
	icons := [5]string{"🍏", "🍎", "🍉", "🍊", "🍋"}
	alive := true
	port := os.Args[1]
	i, err := strconv.ParseInt(port, 10, 64)
	if err != nil {
		log.Fatal(err)
	}
	i = i % 5
	icon := icons[i]
	toggleState := func(i int64, alive *bool) {
		t := time.NewTicker(time.Duration((i+3)*2) * time.Second)
		for {
			select {
			case <-t.C:
				*alive = !*alive
				if *alive {
					fmt.Printf("[%s] %s alive 😃\n", icon, port)
				} else {
					fmt.Printf("[%s] %s dead 😵\n", icon, port)
				}
			}
		}
	}
	handler := func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(fmt.Sprintf("[%s] hello from port %s\n", icon, port)))
	}
	healthChackHandler := func(w http.ResponseWriter, r *http.Request) {
		if alive {
			w.WriteHeader(http.StatusOK)
		} else {
			w.WriteHeader(http.StatusServiceUnavailable)
			w.Write([]byte(fmt.Sprintf("[%s] %s gone 😇\n", icon, port)))
		}
	}

	fmt.Printf("[%s] Server listening on port %s\n", icon, port)
	http.HandleFunc("/", handler)
	http.HandleFunc("/health", healthChackHandler)
	go toggleState(i, &alive)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}
