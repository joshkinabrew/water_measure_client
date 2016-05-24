package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"gopkg.in/alecthomas/kingpin.v2"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/tarm/serial"
)

var (
	url = kingpin.Arg("server", "URL of server to POST results to").Required().String()
)

type Reading struct {
	Value string `json:"value"`
	Time  int64  `json:"time"`
	Host  string `json:"host"`
}

func main() {
	kingpin.Parse()

	c := &serial.Config{Name: "/dev/ttyAMA0", Baud: 9600}
	s, err := serial.OpenPort(c)
	if err != nil {
		log.Fatal(err)
	}

	buf := make([]byte, 128)
	n, err := s.Read(buf)
	if err != nil {
		log.Fatal(err)
	}
	str := strings.Trim(string(buf[:n]), "\r")

	h, err := os.Hostname()
	if err != nil {
		log.Fatal(err)
	}
	r := Reading{Value: str, Time: time.Now().Unix(), Host: h}

	b, err := json.Marshal(r)

	req, err := http.NewRequest("POST", *url, bytes.NewBuffer(b))
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()

	fmt.Println("response Status:", resp.Status)
	fmt.Println("response Headers:", resp.Header)
	// body, _ := ioutil.ReadAll(resp.Body)

	//

}
