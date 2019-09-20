package main

import (
	"bufio"
	"bytes"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"sync"
)

func readFromURL(url string) (string, error) {
	response, err := http.Get(url)
	if err != nil {
		return "", err
	}
	defer response.Body.Close()

	buf := new(bytes.Buffer)
	_, err = buf.ReadFrom(response.Body)
	if err != nil {
		return "", err
	}

	return buf.String(), nil
}

func findGo(url string, goCounter chan int, urlsLimiter chan bool, wg *sync.WaitGroup) {
	defer wg.Done()

	contents, err := readFromURL(url)
	if err != nil {
		log.Fatalf("Error when executing GET request to %s: ", url, err)
	}
	numberOfGo := strings.Count(contents, "Go")
	fmt.Printf("Count for %s: %d\n", url, numberOfGo)
	goCounter <- numberOfGo
	<-urlsLimiter
}

func main() {
	totalGo := 0
	wg := sync.WaitGroup{}
	counter := make(chan int, 5)
	limiter := make(chan bool, 5)

	exiter := make(chan bool)
	go func(exit chan bool) {
		exitSignal := false
		for {
			select {
			case goNumber := <-counter:
				totalGo += goNumber
			case <-exit:
				exitSignal = true
				break
			}
			if exitSignal {
				break
			}
		}
		<-exit
	}(exiter)

	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		wg.Add(1)
		limiter <- true
		go findGo(scanner.Text(), counter, limiter, &wg)
	}
	if scanner.Err() != nil {
		log.Fatal("Error when reading from stdin:", scanner.Err())
	}

	wg.Wait()
	exiter <- true
	exiter <- true

	fmt.Println("Total: ", totalGo)
}
