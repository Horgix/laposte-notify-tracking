package main

import (
  "os"
  "time"
  "io/ioutil"
  "fmt"
  "net/http"
  "encoding/json"
  "log"
  "strings"
  "crypto/x509"
  "crypto/tls"

  "github.com/nlopes/slack"
)

func initClient() *http.Client {
  var err error
  caFile := "/etc/ssl/certs/ca-certificates.crt"

  var ca []byte
  if _, errCA := os.Stat(caFile); errCA == nil {
    ca, err = ioutil.ReadFile(caFile)
    if err != nil {
      log.Fatal("Failed to read CA. %s", err)
      return nil
    }
  } else {
    fmt.Println("NO CA CERT")
  }
  //print(string(ca))
  pool := x509.NewCertPool()
  ok := pool.AppendCertsFromPEM(ca)
  if ! ok {
    fmt.Println("WARNING")
  }
  client := &http.Client{
    Transport: &http.Transport{
      TLSClientConfig: &tls.Config{RootCAs: pool},
    },
  }
  return client
}

var client = initClient()
var counter = 0

func getCurrentStatus() (string, string) {
  trackingId := os.Getenv("LAPOSTE_TRACKING_ID")
  laposteApiKey := os.Getenv("LAPOSTE_API_KEY")

  req, err := http.NewRequest("GET", "https://api.laposte.fr/suivi/v1/" + trackingId, nil)
  if err != nil {
    log.Fatal("NewRequest: ", err)
    return "", "Fail"
  }

  req.Header.Add("X-Okapi-Key", laposteApiKey)
  resp, err := client.Do(req)
  if err != nil {
    log.Fatal("Do: ", err)
    return "", "Fail"
  }
  defer resp.Body.Close()

  if resp.StatusCode != 200 { // OK
    log.Fatal("No 200 StatusCode: ", resp.StatusCode)
    return "", "Fail"
  }
  bodyBytes, _ := ioutil.ReadAll(resp.Body)

  var raw map[string]interface{}
  json.Unmarshal(bodyBytes, &raw)

  out, _ := json.Marshal(raw["message"])
  message := strings.Trim(string(out), "\"")

  return message, ""
}

func handleStatus(message string) {
  counter = counter + 1
  slackToken := os.Getenv("SLACK_TOKEN")

  if message == "En cours de traitement" {
    fmt.Println("Still treating")
    if counter % 60 == 0 {
      api := slack.New(slackToken)
      params :=  slack.PostMessageParameters{Username: "laposte", AsUser: true}
      channelID, timestamp, err := api.PostMessage("@horgix", "Still \"" + message + "\"", params)

      if err != nil {
        fmt.Printf("%s\n", err)
        return
      }
      fmt.Printf("Message successfully sent to channel %s at %s\n", channelID, timestamp)
    }
  } else {
    fmt.Println("Yeaaaah it changed!")
    fmt.Println(message)
    api := slack.New(slackToken)
    params :=  slack.PostMessageParameters{Username: "laposte", AsUser: true}
    channelID, timestamp, err := api.PostMessage("@horgix", message, params)

    if err != nil {
      fmt.Printf("%s\n", err)
      return
    }
    fmt.Printf("Message successfully sent to channel %s at %s\n", channelID, timestamp)
  }
}

func doEvery(d time.Duration, f func(time.Time)) {
  for x := range time.Tick(d) {
    f(x)
  }
}

func getAndNotify(t time.Time) {
  message, e := getCurrentStatus()
  if e == "Fail" {
    log.Fatal("FAIL")
    return
  }
  handleStatus(message)
}

func main() {

  doEvery(30*time.Second, getAndNotify)
}
