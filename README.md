# Overview

Analyzes a URL for response times and HTTP errors.

Given a url this program will issue concurrent http requests and result both summary data on response times across the sample set.  It will also display a histogram of the response times.

If 5XX errors are detected, it will note that in the output and calculate the perctage of errors relative to the number of samples taken.

# Motivation

Occassionaly while developing systems for fun or profit, I will encounter a service URL that I suspect may not be completely reliable.  I wrote this to GET a url a number of times and produce
stats on response times and errors.

# Caution

This program is designed to send a number of http request concurrently.   Setting sample size too high could cause problems for the people running the service you are analyzing so proceed carefully.
Setting it too high could also potentially cause resource starvation on the computer you are running it from and thus affect the results.   I have not tested to see what the upper bound might be.  Use at your own risk.

# Usage

```go-urlchecker [-verbose] [-samples=3] http://example.com```

  Verbose defaults to off.  If you turn it on you get one line per sample printed as the program runs.  Otherwise, by default you see one x printed for each sample as it completes.
  
  Samples defaults to 3.   I find 10 - 100 samples is what I use most often.

## Example:

```go-urlchecker -verbose -samples=5 http://example.com```


## Sample output


```
$ ./go-urlchecker -verbose -samples=3 http://google.com
URL response analyzer.
Taking 3 concurrent samples of http://google.com
        Result: 200     Elapsed Time: 200.207114ms for http://google.com
        Result: 200     Elapsed Time: 202.283383ms for http://google.com
        Result: 200     Elapsed Time: 205.302299ms for http://google.com

Max reponse time: 205.302299ms
Min reponse time: 200.207114ms
Median reponse time: 202.283383ms
99P reponse time: 205.302299ms

 200.207ms - 200.716ms -
 200.716ms - 201.226ms -
 201.226ms - 201.735ms -

No errors detected
```

