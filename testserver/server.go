package main

import (
	"log"
	"fmt"
	"net/http"
	"io/ioutil"
)

func handler(w http.ResponseWriter, r *http.Request) {
	log.Println(r.URL.String(), r.Method)
	log.Println("Proto:", r.Proto)
	log.Println("Header", r.Header)
	fmt.Println("--body--")
	defer r.Body.Close()
	byte, _ := ioutil.ReadAll(r.Body)
	log.Println(string(byte))


	fmt.Fprintf(w, "hello")

}

func main() {
	http.HandleFunc("/", handler)
	log.Println("start listening :18888")
	http.ListenAndServe(":18888", nil)
}
