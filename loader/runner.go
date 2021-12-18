package loader

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"strings"
	"sync/atomic"
	"time"

	"github.com/ryanking8215/go-wrk/util"
)

const (
	USER_AGENT = "go-wrk"
)

type Runner struct {
	cfg             Config
	statsAggregator chan *RequesterStats
	interrupted     int32
}

func NewRunner(cfg Config, statsAggregater chan *RequesterStats) *Runner {
	return &Runner{
		cfg:             cfg,
		statsAggregator: statsAggregater,
		interrupted:     0,
	}
}

//Requester a go function for repeatedly making requests and aggregating statistics as long as required
//When it is done, it sends the results using the statsAggregator channel
func (runner *Runner) RunSingleSession(script *ScriptContext) {
	stats := &RequesterStats{MinRequestTime: time.Minute}
	start := time.Now()
	cfg := &runner.cfg
	if script != nil {
		cfg = &script.Config
	}

	httpClient, err := client(cfg.DisableCompression, cfg.DisableKeepAlive, cfg.SkipVerify,
		cfg.Timeoutms, cfg.AllowRedirects, cfg.ClientCert, cfg.ClientKey, cfg.CaCert, cfg.Http2)
	if err != nil {
		log.Fatal(err)
	}

	for time.Since(start).Seconds() <= float64(cfg.Duration) && atomic.LoadInt32(&runner.interrupted) == 0 {
		if script != nil && script.request != nil {
			script.request()
		}

		respSize, reqDur := DoRequest(httpClient, cfg.Header, cfg.Method, cfg.Host, cfg.TestUrl, cfg.ReqBody, script)
		if respSize > 0 {
			stats.TotRespSize += int64(respSize)
			stats.TotDuration += reqDur
			stats.MaxRequestTime = util.MaxDuration(reqDur, stats.MaxRequestTime)
			stats.MinRequestTime = util.MinDuration(reqDur, stats.MinRequestTime)
			stats.NumRequests++
		} else {
			stats.NumErrs++
		}

		if script != nil && script.stop != nil && script.stop() {
			runner.Stop()
		}

		if script != nil && script.delay != nil {
			time.Sleep(time.Duration(script.delay()) * time.Millisecond)
		}
	}

	runner.statsAggregator <- stats
}

func (runner *Runner) Stop() {
	atomic.StoreInt32(&runner.interrupted, 1)
}

// RequesterStats used for colelcting aggregate statistics
type RequesterStats struct {
	TotRespSize    int64
	TotDuration    time.Duration
	MinRequestTime time.Duration
	MaxRequestTime time.Duration
	NumRequests    int
	NumErrs        int
}

func escapeUrlStr(in string) string {
	qm := strings.Index(in, "?")
	if qm != -1 {
		qry := in[qm+1:]
		qrys := strings.Split(qry, "&")
		var query string = ""
		var qEscaped string
		var first bool = true
		for _, q := range qrys {
			qSplit := strings.Split(q, "=")
			if len(qSplit) == 2 {
				qEscaped = qSplit[0] + "=" + url.QueryEscape(qSplit[1])
			} else {
				qEscaped = qSplit[0]
			}
			if first {
				first = false
			} else {
				query += "&"
			}
			query += qEscaped

		}
		return in[:qm] + "?" + query
	} else {
		return in
	}
}

//DoRequest single request implementation. Returns the size of the response and its duration
//On error - returns -1 on both
func DoRequest(httpClient *http.Client, header map[string]string, method, host, loadUrl, reqBody string, script *ScriptContext) (respSize int, duration time.Duration) {
	respSize = -1
	duration = -1

	loadUrl = escapeUrlStr(loadUrl)

	var buf io.Reader
	if len(reqBody) > 0 {
		buf = bytes.NewBufferString(reqBody)
	}

	req, err := http.NewRequest(method, loadUrl, buf)
	if err != nil {
		fmt.Println("An error occured doing request", err)
		return
	}

	for hk, hv := range header {
		req.Header.Add(hk, hv)
	}

	req.Header.Add("User-Agent", USER_AGENT)
	if host != "" {
		req.Host = host
	}
	start := time.Now()
	resp, err := httpClient.Do(req)
	if err != nil {
		fmt.Println("redirect?")
		//this is a bit weird. When redirection is prevented, a url.Error is retuned. This creates an issue to distinguish
		//between an invalid URL that was provided and and redirection error.
		rr, ok := err.(*url.Error)
		if !ok {
			fmt.Println("An error occured doing request", err, rr)
			return
		}
		fmt.Println("An error occured doing request", err)
	}
	if resp == nil {
		fmt.Println("empty response")
		return
	}
	defer func() {
		if resp != nil && resp.Body != nil {
			resp.Body.Close()
		}
	}()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Println("An error occured reading body", err)
	}
	if resp.StatusCode == http.StatusOK || resp.StatusCode == http.StatusCreated {
		duration = time.Since(start)
		respSize = len(body) + int(util.EstimateHttpHeadersSize(resp.Header))
	} else if resp.StatusCode == http.StatusMovedPermanently || resp.StatusCode == http.StatusTemporaryRedirect {
		duration = time.Since(start)
		respSize = int(resp.ContentLength) + int(util.EstimateHttpHeadersSize(resp.Header))
	} else {
		fmt.Println("received status code", resp.StatusCode, "from", resp.Header, "content", string(body), req)
	}

	if script != nil && script.response != nil {
		script.response(resp.StatusCode, resp.Header, string(body))
	}

	return
}
