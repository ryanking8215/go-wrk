go-wrk - an HTTP benchmarking tool with JavaScript support.
============================================================

Now you can write JavaScript file to setup your complicated benchmarking tasks when using go-wrk!

go-wrk is a modern HTTP benchmarking tool capable of generating significant load when run on a single multi-core CPU. It builds on go language go routines and scheduler for behind the scenes async IO and concurrency.

It was created mostly to examine go language (http://golang.org) performance and verbosity compared to C (the language wrk was written in. See - <https://github.com/wg/wrk>).  
It turns out that it is just as good in terms of throughput! And with a lot less code.  

The majority of go-wrk is the product of one afternoon, and its quality is comparable to wrk.

By the way, the JavaScript support also took me about one afternoon. -:)


Building
--------

    go get github.com/ryanking8215/go-wrk

This will download and compile go-wrk. The binary will be placed under your $GOPATH/bin directory  
   
Command line parameters (./go-wrk -help)  
	
       Usage: go-wrk <options> <url>
       Options:
        -H       Header to add to each request (you can define multiple -H flags) (Default )
        -M       HTTP method (Default GET)
        -T       Socket/request timeout in ms (Default 1000)
        -body    request body string or @filename (Default )
        -c       Number of goroutines to use (concurrent connections) (Default 10)
        -ca      CA file to verify peer against (SSL/TLS) (Default )
        -cert    CA certificate file to verify peer against (SSL/TLS) (Default )
        -d       Duration of test in seconds (Default 10)
        -f       Playback file name (Default <empty>)
        -help    Print help (Default false)
        -host    Host Header (Default )
        -http    Use HTTP/2 (Default true)
        -key     Private key file name (SSL/TLS (Default )
        -no-c    Disable Compression - Prevents sending the "Accept-Encoding: gzip" header (Default false)
        -no-ka   Disable KeepAlive - prevents re-use of TCP connections between different HTTP requests (Default false)
        -no-vr   Skip verifying SSL certificate of the server (Default false)
        -redir   Allow Redirects (Default false)
        -v       Print version details (Default false)
        -s       Load javascript script file. 

Basic Usage
-----------

    ./go-wrk -c 80 -d 5  http://192.168.1.118:8080/json

This runs a benchmark for 5 seconds, using 80 go routines (connections)

Output:

    Running 10s test @ http://192.168.1.118:8080/json
      80 goroutine(s) running concurrently
       142470 requests in 4.949028953s, 19.57MB read
         Requests/sec:		28787.47
         Transfer/sec:		3.95MB
         Avg Req Time:		0.0347ms
         Fastest Request:	0.0340ms
         Slowest Request:	0.0421ms
         Number of Errors:	0


Javascript Supports
--------------------
Javascript is powered by [Goja](https://github.com/dop251/goja).

```shell
  ./go-wrk -c 10 -d 5 -s <my-script-file> http://192.168.1.118:8080/json
  ./go-wrk -s <my-script-file> # all defines in script file
```

Feel free to use javascript to setup your benchmarking tasks.
Please refer to the example files in 'scripts' directory.

Benchmarking Tips
-----------------

  The machine running go-wrk must have a sufficient number of ephemeral ports
  available and closed sockets should be recycled quickly. To handle the
  initial connection burst the server's listen(2) backlog should be greater
  than the number of concurrent connections being tested.

Acknowledgements
----------------

  [The original author](https://github.com/tsliwowicz) fully credit the wrk project (https://github.com/wg/wrk) for the inspiration and even parts of this text.  
  [The original author](https://github.com/tsliwowicz) also used similar command line arguments format and output format.

  I added JavaScript support in go-wrk like lua in wrk.
