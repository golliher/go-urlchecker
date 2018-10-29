package main

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"time"

	"github.com/namsral/flag"
)

type Result struct {
	responseTime time.Duration
	response     *http.Response
}

func get_url(targeturl string, c chan Result) {

	start := time.Now()
	timeout := time.Duration(13 * time.Second)

	client := http.Client{
		Timeout: timeout,
	}

	u, err := url.Parse(targeturl)
	if err != nil {
		log.Fatal(err)
	}

	req, _ := http.NewRequest("GET", targeturl, nil)
	req.URL.Host = u.Host

	req.URL.Scheme = u.Scheme
	req.Header.Set("X-Forwarded-Host", req.Header.Get("Host"))
	req.Host = u.Host

	resp, err := client.Do(req)
	if err != nil {
		log.Println(err)
		return
	}
	elapsed := time.Since(start)

	if resp.StatusCode < 500 {
		c <- Result{responseTime: elapsed, response: resp}
	}

	return
}

func fanout_get_url(url string, concurrentRequests int) http.Response {

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
		body := "504 Gateway Timeout (the proxy did not receiver any responses under the allowed time budget)"
		return http.Response{
			Status:        "504 Gateway Timeout",
			StatusCode:    504,
			Proto:         "HTTP/1.1",
			ProtoMajor:    1,
			ProtoMinor:    1,
			Body:          ioutil.NopCloser(bytes.NewBufferString(body)),
			ContentLength: int64(len(body)),
			Header:        make(http.Header, 0),
		}
	}
}

func handleRequestsWithProxy(res http.ResponseWriter, req *http.Request) {
	url := fmt.Sprintf("%v", req.URL)

	resp := fanout_get_url(origin+url, fanout_width)
	defer resp.Body.Close()

	io.Copy(res, resp.Body)

}

var origin string
var fanout_width int

func main() {

	version := "0.0.1"
	originPtr := flag.String("origin", "http://golliher.net/test", "URL to use at the origin. e.g. http://example.com")
	portPtr := flag.String("port", "8080", "Port the proxy should listen on")
	samplesPtr := flag.Int("samples", 3, "Number of times to run the check against the target url")
	versionPtr := flag.Bool("version", false, "Prints the version")

	flag.Parse()

	if *versionPtr == true {
		fmt.Printf("Version: %s\n", version)
		os.Exit(0)
	}

	origin = *originPtr
	fanout_width = *samplesPtr

	log.Println("Starting up proxy: Defense against unreliable backend services")
	log.Println("Version:", version)
	log.Println("Proxying to origin:", origin)
	log.Printf("Will make %d calls to origin and take the fastest non-500 response\n", fanout_width)

	// start server
	http.HandleFunc("/", handleRequestsWithProxy)
	if err := http.ListenAndServe(":"+*portPtr, nil); err != nil {
		panic(err)
	}

}
