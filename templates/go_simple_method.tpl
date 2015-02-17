/*
import (
    "net/http"
    "io/ioutil"
    "log"
)
*/

client := &http.Client{}
request, err := http.NewRequest("{{ .Method }}", {{ .Url }}, nil)
resp, err := client.Do(request)
if err != nil {
    log.Fatal(err)
}
defer resp.Body.Close()
body, err := ioutil.ReadAll(resp.Body)
if err != nil {
    log.Fatal(err)
}
log.Print(string(body))
