package main

import (
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"os/signal"
	"runtime"
	"strings"
	"time"

	"github.com/ryanking8215/go-wrk/loader"
	"github.com/ryanking8215/go-wrk/util"
)

const APP_VERSION = "0.10"

//default that can be overridden from the command line
var versionFlag bool = false
var helpFlag bool = false
var headerFlags util.HeaderList
var statsAggregator chan *loader.RequesterStats
var playbackFile string
var scriptFile string
var config = loader.NewConfig()
var script *loader.ScriptContext

func init() {
	flag.BoolVar(&versionFlag, "v", false, "Print version details")
	flag.BoolVar(&config.AllowRedirects, "redir", false, "Allow Redirects")
	flag.BoolVar(&helpFlag, "help", false, "Print help")
	flag.BoolVar(&config.DisableCompression, "no-c", false, "Disable Compression - Prevents sending the \"Accept-Encoding: gzip\" header")
	flag.BoolVar(&config.DisableKeepAlive, "no-ka", false, "Disable KeepAlive - prevents re-use of TCP connections between different HTTP requests")
	flag.BoolVar(&config.SkipVerify, "no-vr", false, "Skip verifying SSL certificate of the server")
	flag.IntVar(&config.Goroutines, "c", 10, "Number of goroutines to use (concurrent connections)")
	flag.IntVar(&config.Duration, "d", 10, "Duration of test in seconds")
	flag.IntVar(&config.Timeoutms, "T", 1000, "Socket/request timeout in ms")
	flag.StringVar(&config.Method, "M", "GET", "HTTP method")
	flag.StringVar(&config.Host, "host", "", "Host Header")
	flag.Var(&headerFlags, "H", "Header to add to each request (you can define multiple -H flags)")
	flag.StringVar(&playbackFile, "f", "<empty>", "Playback file name")
	flag.StringVar(&config.ReqBody, "body", "", "request body string or @filename")
	flag.StringVar(&config.ClientCert, "cert", "", "CA certificate file to verify peer against (SSL/TLS)")
	flag.StringVar(&config.ClientKey, "key", "", "Private key file name (SSL/TLS")
	flag.StringVar(&config.CaCert, "ca", "", "CA file to verify peer against (SSL/TLS)")
	flag.BoolVar(&config.Http2, "http", true, "Use HTTP/2")
	flag.StringVar(&scriptFile, "s", "", "Load javascript script file")
}

//printDefaults a nicer format for the defaults
func printDefaults() {
	fmt.Println("Usage: go-wrk <options> <url>")
	fmt.Println("Options:")
	flag.VisitAll(func(flag *flag.Flag) {
		fmt.Println("\t-"+flag.Name, "\t", flag.Usage, "(Default "+flag.DefValue+")")
	})
}

func cfg() *loader.Config {
	if script != nil {
		return &script.Config
	}
	return &config
}

func main() {
	flag.Parse() // Scan the arguments list
	if scriptFile != "" {
		var err error
		script, err = loader.LoadScript(config, scriptFile)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
	}

	//raising the limits. Some performance gains were achieved with the + goroutines (not a lot).
	runtime.GOMAXPROCS(runtime.NumCPU() + cfg().Goroutines)

	for _, hdr := range headerFlags {
		hp := strings.SplitN(hdr, ":", 2)
		cfg().Header[hp[0]] = hp[1]
	}

	if playbackFile != "<empty>" {
		file, err := os.Open(playbackFile) // For read access.
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		defer file.Close()
		url, err := io.ReadAll(file)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		cfg().TestUrl = string(url)
	} else {
		if len(cfg().TestUrl) == 0 { // if not defined in script, take it from commad.
			cfg().TestUrl = flag.Arg(0)
		}
	}

	if versionFlag {
		fmt.Println("Version:", APP_VERSION)
		return
	} else if helpFlag || len(cfg().TestUrl) == 0 {
		printDefaults()
		return
	}

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt)

	fmt.Printf("Running %vs test @ %v\n  %v goroutine(s) running concurrently\n", cfg().Duration, cfg().TestUrl, cfg().Goroutines)

	if len(cfg().ReqBody) > 0 && cfg().ReqBody[0] == '@' {
		bodyFilename := cfg().ReqBody[1:]
		data, err := ioutil.ReadFile(bodyFilename)
		if err != nil {
			fmt.Println(fmt.Errorf("could not read file %q: %v", bodyFilename, err))
			os.Exit(1)
		}
		cfg().ReqBody = string(data)
	}

	statsAggregator = make(chan *loader.RequesterStats, cfg().Goroutines)
	runner := loader.NewRunner(config, statsAggregator)
	for i := 0; i < cfg().Goroutines; i++ {
		var s *loader.ScriptContext
		if scriptFile != "" {
			s, _ = loader.LoadScript(*cfg(), scriptFile)
		}
		go runner.RunSingleSession(s)
	}

	responders := 0
	aggStats := loader.RequesterStats{MinRequestTime: time.Minute}

	for responders < cfg().Goroutines {
		select {
		case <-sigChan:
			runner.Stop()
			fmt.Printf("stopping...\n")
		case stats := <-statsAggregator:
			aggStats.NumErrs += stats.NumErrs
			aggStats.NumRequests += stats.NumRequests
			aggStats.TotRespSize += stats.TotRespSize
			aggStats.TotDuration += stats.TotDuration
			aggStats.MaxRequestTime = util.MaxDuration(aggStats.MaxRequestTime, stats.MaxRequestTime)
			aggStats.MinRequestTime = util.MinDuration(aggStats.MinRequestTime, stats.MinRequestTime)
			responders++
		}
	}

	if aggStats.NumRequests == 0 {
		fmt.Printf("Error: No statistics collected / no requests found\n")
		return
	}

	avgThreadDur := aggStats.TotDuration / time.Duration(responders) //need to average the aggregated duration

	reqRate := float64(aggStats.NumRequests) / avgThreadDur.Seconds()
	avgReqTime := aggStats.TotDuration / time.Duration(aggStats.NumRequests)
	bytesRate := float64(aggStats.TotRespSize) / avgThreadDur.Seconds()
	fmt.Printf("%v requests in %v, %v read\n", aggStats.NumRequests, avgThreadDur, util.ByteSize{Size: float64(aggStats.TotRespSize)})
	fmt.Printf("Requests/sec:\t\t%.2f\nTransfer/sec:\t\t%v\nAvg Req Time:\t\t%v\n", reqRate, util.ByteSize{Size: bytesRate}, avgReqTime)
	fmt.Printf("Fastest Request:\t%v\n", aggStats.MinRequestTime)
	fmt.Printf("Slowest Request:\t%v\n", aggStats.MaxRequestTime)
	fmt.Printf("Number of Errors:\t%v\n", aggStats.NumErrs)
}
