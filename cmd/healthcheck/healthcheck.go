/*
Healthcheck is a simple program that sends an HTTP request to the local host (self) to a configured port number.
It's used in environment where you need a simple probe for health checks (e.g., an empty container in docker).
The probe URL is derived from host and port flags.

Usage:

	healthcheck [flags]

The flags are:

	-host <hostname>
		Change the host where the request is sent.

	-port <1-65535>
		Change the port where the request is sent.

Return values (exit codes):

	0
		The request was successful (HTTP 200 or HTTP 204)

	> 0
		The request was not successful (connection error or unexpected HTTP status code)
*/
package main

import (
	"flag"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"os"
)

func main() {
	var host = flag.String("host", "localhost", "HTTP host for healthcheck")
	var port = flag.Int("port", 3000, "HTTP port for healthcheck")

	flag.Parse()

	target := url.URL{
		Scheme: "http",
		Host:   net.JoinHostPort(*host, fmt.Sprintf("%d", *port)),
		Path:   "/liveness",
	}

	res, err := http.Get(target.String())
	if err != nil {
		_, _ = fmt.Fprintln(os.Stderr, err.Error())
		os.Exit(1)
	} else if res.StatusCode != http.StatusOK && res.StatusCode != http.StatusNoContent {
		_ = res.Body.Close()
		_, _ = fmt.Fprintln(os.Stderr, "Healthcheck request not OK: ", res.Status)
		os.Exit(1)
	}
	_ = res.Body.Close()
	os.Exit(0)
}
