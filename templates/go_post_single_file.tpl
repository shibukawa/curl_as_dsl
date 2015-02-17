package main

import (
    "os"
    "log"
    "io/ioutil"
    "net/http"
)

func main() {
    file, err := os.Open("{{ .FilePath }}")
    if err != nil {
        log.Fatal(err)
    }
    resp, err := http.Post({{ .Url }}, "{{ .ContentType }}", file)
    if err != nil {
        log.Fatal(err)
    }
    defer resp.Body.Close()
    body, err := ioutil.ReadAll(resp.Body)
    if err != nil {
        log.Fatal(err)
    }
    log.Print(string(body))
}
