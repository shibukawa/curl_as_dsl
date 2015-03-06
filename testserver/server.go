package main

import (
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
)

func handler(w http.ResponseWriter, r *http.Request) {
	log.Println(r.URL.String(), r.Method)
	log.Println("Method:", r.Proto)
	log.Println("Header", r.Header)
	fmt.Println("--body--")
	defer r.Body.Close()
	byte, _ := ioutil.ReadAll(r.Body)
	log.Println(string(byte))

	fmt.Fprintf(w, "hello\n")
}

func jsHandler(w http.ResponseWriter, r *http.Request) {
	log.Println(r.URL.String(), r.Method)
	log.Println("Method:", r.Proto)

	filename := "test.html"

	if _, err := os.Stat(filename); os.IsNotExist(err) {
		fmt.Printf("no such file or directory: %s", filename)
		w.WriteHeader(404)
		fmt.Fprint(w, "404 Not Found (1)\n")
		return
	}

	in, err := os.Open(filename)
	if err != nil {
		fmt.Printf("file read errory: %v", err)
		w.WriteHeader(404)
		fmt.Fprint(w, "404 Not Found (2)\n")
		return
	}
	defer in.Close()
	w.Header().Add("Access-Control-Allow-Origin", "*")
	io.Copy(w, in)
}

func authHandler(w http.ResponseWriter, r *http.Request) {
	log.Println(r.URL.String(), r.Method)
	log.Println("Method:", r.Proto)
	log.Println("Header", r.Header)
	fmt.Println("--body--")
	defer r.Body.Close()
	byte, _ := ioutil.ReadAll(r.Body)
	log.Println(string(byte))

	if r.Header.Get("Authorization") == "" {
		w.Header().Add("WWW-Authenticate", `Digest realm="testrealm@host.com", nonce="dcd98b7102dd2f0e8b11d0f600bfb0c093", opaque="5ccc069c403ebaf9f0171e9517f40e41"`)
		w.WriteHeader(401)
	} else {
		fmt.Fprintf(w, "hello\n")
	}
}

func main() {
	http.HandleFunc("/auth", authHandler)
	http.HandleFunc("/", handler)
	http.HandleFunc("/js", jsHandler)
	log.Println("start listening :18888")
	http.ListenAndServe(":18888", nil)
}
