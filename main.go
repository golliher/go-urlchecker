package main

import (
	"flag"
	"fmt"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/jamiealquiza/tachymeter"
)

type Result struct {
	statusCode   int
	responseTime time.Duration
}

func get_url(url string, wg *sync.WaitGroup, t *tachymeter.Tachymeter, c chan Result) {

	start := time.Now()
	resp, err := http.Get(url)
	if err != nil {
		fmt.Println("Error with GET")
	}
	elapsed := time.Since(start)

	wg.Done()
	t.AddTime(elapsed)
	if resp.StatusCode != 200 {
	}

	c <- Result{statusCode: resp.StatusCode, responseTime: elapsed}

	return
}

func main() {

	version := "1.0.0"

	samplesPtr := flag.Int("samples", 3, "Number of times to run the check against the target url")
	verbosePtr := flag.Bool("verbose", false, "Enable verbose output.")
	versionPtr := flag.Bool("version", false, "Prints the version")

	flag.Parse()

	if len(os.Args[1:]) < 1 {
		fmt.Println("You must specify a target URL on the command line")
		fmt.Printf("Usage:\n\tgo-urlchecker [-verbose] [-samples=3] http://example.com\n")
		fmt.Printf("\nExample:\ngo-urlchecker -verbose -samples=5 http://example.com\n")
		os.Exit(1)
	}

	if *versionPtr == true {
		fmt.Printf("Version: %s\n", version)
		os.Exit(0)
	}

	url := flag.Arg(0)

	c := make(chan Result)
	t := tachymeter.New(&tachymeter.Config{Size: *samplesPtr})
	var wg sync.WaitGroup

	fmt.Println("URL response analyzer.")
	fmt.Printf("Taking %d concurrent samples of %s\n", *samplesPtr, url)
	for i := 0; i < *samplesPtr; i++ {
		wg.Add(1)
		go get_url(url, &wg, t, c)

	}

	ec := 0
	for i := 0; i < *samplesPtr; i++ {
		r := <-c
		if *verbosePtr == true {
			fmt.Printf("\tResult: %d\tElapsed Time: %s for %s\n", r.statusCode, r.responseTime, url)
		} else {
			fmt.Printf("x")
		}

		if r.statusCode >= 500 {
			ec++
		}
	}

	wg.Wait()

	metrics := t.Calc()
	fmt.Printf("\nMax reponse time: %s\n", metrics.Time.Max)
	fmt.Printf("Min reponse time: %s\n", metrics.Time.Min)
	fmt.Printf("Median reponse time: %s\n", metrics.Time.P50)
	fmt.Printf("99P reponse time: %s\n\n", metrics.Time.P99)
	fmt.Println(metrics.Histogram.String(25))

	ep := float64(float64(ec)/float64(*samplesPtr)) * 100

	if ep > 0 {
		fmt.Printf("ERRORS DETECTED:  There were %d errors over the sample size of %d (%.0f percent of request produced a 5xx code)\n", ec, *samplesPtr, ep)
		fmt.Printf(" This indicates a problem that needs attention\n")
	} else {
		fmt.Println("No errors detected")
	}

}
