package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"
)

type Result struct {
	responseTime time.Duration
	response     *http.Response
}

func get_url(url string, c chan Result) {

	start := time.Now()
	timeout := time.Duration(13 * time.Second)

	client := http.Client{
		Timeout: timeout,
	}

	resp, err := client.Get(url)
	if err != nil {
		fmt.Println(err)
		return
	}
	elapsed := time.Since(start)

	if resp.StatusCode < 500 {
		c <- Result{responseTime: elapsed, response: resp}
	}

	return
}

func doit(url string, concurrentRequests int) http.Response {

	c := make(chan Result)

	log.Printf("Making %d concurrent requests for %s\n", concurrentRequests, url)
	for i := 0; i < concurrentRequests; i++ {
		go get_url(url, c)
	}

	select {
	case r := <-c:
		log.Printf("Request fullilled in: %s with HTTP status code of %d\n", r.responseTime, r.response.StatusCode)
		return *r.response

	case <-time.After(3 * time.Second):
		log.Println("Timeout: None of the concurrent request returned fast enough. Returning 504 Gateway timeout")
		//		body := "Hello world"

		return http.Response{
			Status:     "504 Gateway Timeout",
			StatusCode: 504,
			Proto:      "HTTP/1.1",
			ProtoMajor: 1,
			ProtoMinor: 1,
			// Body:          ioutil.NopCloser(bytes.NewBufferString(body)),
			// ContentLength: int64(len(body)),
			//Request: req,
			Header: make(http.Header, 0),
		}
	}

}

func main() {

	version := "1.0.0"
	originPtr := flag.String("origin", "http://golliher.net/test", "URL to use at the origin. e.g. http://example.com")
	samplesPtr := flag.Int("samples", 3, "Number of times to run the check against the target url")
	versionPtr := flag.Bool("version", false, "Prints the version")

	flag.Parse()

	if *versionPtr == true {
		fmt.Printf("Version: %s\n", version)
		os.Exit(0)
	}

	origin := *originPtr

	log.Println("Starting up proxy: Defense against unreliable backend services")
	log.Println("Version:", version)
	log.Println("Proxying to origin:", origin)
	log.Printf("Will make %d calls to origin and take the fastest non-500 response\n", *samplesPtr)

	resp := doit(origin, *samplesPtr)
	log.Println(resp)

	// Leaving off.. I now have a function that request a response object
	// TODO: turn this into a proxy server

}
