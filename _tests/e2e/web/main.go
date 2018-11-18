package main

import (
	"context"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"net/http/httputil"
	"os"
	"strings"
	"time"
)

func main() {
	mux := http.NewServeMux()

	mux.HandleFunc("/ping", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("pong"))
	})
	mux.HandleFunc("/echo", func(w http.ResponseWriter, r *http.Request) {
		if msg, ok := r.URL.Query()["message"]; ok {
			w.Write([]byte(msg[0]))
			return
		}
		w.WriteHeader(http.StatusBadRequest)
	})
	mux.HandleFunc("/delegate", func(w http.ResponseWriter, r *http.Request) {
		if url, ok := r.URL.Query()["url"]; ok {
			resp, err := httpClient().Get(url[0])
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				w.Write([]byte(err.Error()))
				return
			}
			data, err := ioutil.ReadAll(resp.Body)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				w.Write([]byte(err.Error()))
				return
			}
			w.Write(data)
			return
		}
		w.WriteHeader(http.StatusBadRequest)
	})

	h := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		dump, err := httputil.DumpRequest(r, false)
		if err == nil {
			log.Println(string(dump))
		}
		mux.ServeHTTP(w, r)
	})

	if err := http.ListenAndServe(":"+os.Getenv("PORT"), h); err != nil {
		log.Fatal(err.Error())
	}
}

func httpClient() *http.Client {
	dialerFunc := func(ctx context.Context, network, address string) (net.Conn, error) {
		d := net.Dialer{}
		port := "53"
		if v, ok := os.LookupEnv("DNS_PORT"); ok {
			port = v
		}
		return d.DialContext(ctx, "udp", strings.Replace(address, ":53", fmt.Sprintf(":%s", port), 1))
	}

	resolver := &net.Resolver{PreferGo: true, Dial: dialerFunc}
	dialer := net.Dialer{Resolver: resolver}

	transport := &http.Transport{
		Proxy:                 http.ProxyFromEnvironment,
		Dial:                  dialer.Dial,
		DialContext:           dialer.DialContext,
		MaxIdleConns:          100,
		IdleConnTimeout:       90 * time.Second,
		TLSHandshakeTimeout:   10 * time.Second,
		ExpectContinueTimeout: 1 * time.Second,
	}

	return &http.Client{Transport: transport}
}
