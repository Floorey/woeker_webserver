package main

import (
	"fmt"
	"log"
	"net/http"
	"sync"
)

func worker(id int, jobs <-chan string, results chan<- string) {
	for url := range jobs {
		results <- url
		fmt.Printf("Worker %d started processing %s\n", id, url)

		results <- fmt.Sprintf("Worker %d finished processing %s\n", id, url)
	}
}
func logger(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Printf("%s %s %s\n", r.RemoteAddr, r.Method, r.URL)
		next.ServeHTTP(w, r)
	})
}

func main() {
	jobs := make(chan string)
	results := make(chan string)

	for w := 1; w <= 3; w++ {
		go worker(w, jobs, results)
	}
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		for {
			select {
			case url, ok := <-results:
				if !ok {
					return
				}
				fmt.Printf(url)
			}
		}
	}()

	http.Handle("/", logger(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		jobs <- r.URL.String()
		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, "Anfrage für URL %s wurde an die Worker gesendet", r.URL.String())
	})))
	go func() {
		if err := http.ListenAndServe(":8080", nil); err != nil {
			fmt.Println(err)
		}
	}()
	fmt.Printf("Server läuft auf Port 8080...")
	wg.Wait()

	close(results)
}
