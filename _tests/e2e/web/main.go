package main

import (
	"io/ioutil"
	"log"
	"net/http"
	"net/http/httputil"
	"os"
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
			resp, err := http.Get(url[0])
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
