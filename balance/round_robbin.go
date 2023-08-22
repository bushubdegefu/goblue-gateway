package balance

import (
	"fmt"
	"net"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	"semaygateway.com/bluewrapper"
	"semaygateway.com/gatelogger"
	"semaygateway.com/gateparse"
	"semaygateway.com/ratelimit"
)

var mu sync.Mutex
var idx int = 0
var error_counter int = 0 // used to control how many times to retry if backends are down

// round robbin load balancing and ratelimiting logics implemented here
// returns the base http url path for the next response based on round robbin balancing
func RobbinBalance(w http.ResponseWriter, r *http.Request) {
	// Finding which service to Load balance from request URL path
	service_name := strings.Split(r.URL.Path, "/")

	//ratelimit Checks
	ratelimtter := ratelimit.NewSlidingWindowRateLimiter(service_name[1])

	rate_allow := ratelimtter.Allow()
	if !rate_allow {
		r.Body.Close()
		return
	}

	// Getting target List based on service from above
	targets, _ := gateparse.GetTargetLists(service_name[1])
	maxLen := len(targets)

	// Round Robin, mutex to prevent race condition updates
	mu.Lock()
	targetURL, err := url.Parse(targets[idx%maxLen])
	if err != nil {
		gatelogger.GateLoggerInfo(err.Error())
	}
	// if target host is not alive increment
	if !IsAlive(targetURL) {
		idx++
		targetURL, _ = url.Parse(targets[idx%maxLen])
	}
	idx++
	if idx == (4 * maxLen) {
		idx = 0
	}

	mu.Unlock()
	// ratelimit.Slider.AddCounter()

	// custom route adjusted reverse proxy, basically returns httputil.ReverseProxy
	// with rewrite on URL with provided serveice name
	reverseProxy := bluewrapper.NewMultipleHostReverseProxy(targetURL, service_name[1])
	reverseProxy.Transport = &http.Transport{
		ResponseHeaderTimeout: 2 * time.Second,
	}
	reverseProxy.ErrorHandler = func(w http.ResponseWriter, r *http.Request, e error) {
		mu.Lock()
		// NOTE: This is Retry Limit if the current backend is dead, mutex to prevent race condition updates
		error_counter++
		if error_counter > (3 * maxLen) {
			error_counter = 0
		}
		mu.Unlock()

		// reset the adjusted path before recursive call
		r.URL.Path = "/" + service_name[1] + r.URL.Path
		gatelogger.GateLoggerInfo(fmt.Sprintf("%v is dead.", targetURL))
		if error_counter < 5 {

			RobbinBalance(w, r)
		}
	}

	gatelogger.GateLoggerInfo("Current Backend: => " + targetURL.Host)

	reverseProxy.ServeHTTP(w, r)

}

// pingBackend checks if the backend is alive.
func IsAlive(url *url.URL) bool {
	conn, err := net.DialTimeout("tcp", url.Host, time.Minute*1)

	if err != nil {
		alive_err := fmt.Sprintf("Unreachable to %v, error:%v ", url.Host, err.Error())
		gatelogger.GateLoggerInfo(alive_err)

		return false
	}
	defer conn.Close()
	return true
}
