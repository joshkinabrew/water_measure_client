package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"gopkg.in/alecthomas/kingpin.v2"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"regexp"
	"strings"
	"time"

	"github.com/tarm/serial"
)

var (
	serverURL      = kingpin.Arg("server", "URL of server to POST results to").Required().String()
	fileDir        = "/home/pi/backup_readings/"
	backupFileName = "backup_readings.json"
	seq            = Readings{}
)

type Readings []Reading

type Reading struct {
	Value string `json:"value"`
	Time  int64  `json:"time"`
	Host  string `json:"host"`
}

func main() {
	kingpin.Parse()

	r := Reading{}
	r.setHostname()
	r.setReadingTime()
	r.setValueFromSerial()

	status := r.sendJSONToServer()
	log.Printf("Status Code: %v", status)

	if strings.HasPrefix(status, "200") {
		log.Print("Successfully POSTed reading to server")
		hasBackupReadings, backupReadings := r.hasBackupReadingsInJSONFile()
		if hasBackupReadings {
			log.Print(fmt.Sprint("There are ", len(backupReadings), " readings in backup JSON file"))
			sendBackupReadingsToServer(backupReadings)
		}
	} else {
		log.Print("Couldn't POST to server, saving in backup file to try again later")
		r.writeReadingToFile(fileDir)
	}
}

func sendBackupReadingsToServer(backupReadings Readings) {
	for index, reading := range backupReadings {
		log.Print(fmt.Sprint("Sending Reading ", index+1, " of ", len(backupReadings)))
		status := reading.sendJSONToServer()

		if strings.HasPrefix(status, "200") {
			log.Print("Successfully sent reading to server")
			reading.removeReadingFromBackupFile()
		}

	}

}

func check(err error) {
	if err != nil {
		log.Fatal(err.Error())
	}
}

func writeJSONToFile(json []byte) error {
	path := fmt.Sprint(fileDir, backupFileName)
	werr := ioutil.WriteFile(path, json, os.ModePerm)

	return werr
}

func readSerialValue() (value string) {
	c := &serial.Config{Name: "/dev/ttyAMA0", Baud: 9600}
	s, err := serial.OpenPort(c)
	check(err)

	buf := make([]byte, 128)
	n, err := s.Read(buf)
	check(err)

	return strings.Trim(string(buf[:n]), "\r")
}

func (readings Readings) pos(reading Reading) int {
	for index, v := range readings {
		if v.Time == reading.Time {
			return index
		}
	}
	return -1
}

func (r *Reading) removeReadingFromBackupFile() {
	path := fmt.Sprint(fileDir, backupFileName)
	readingsFromJSONFile := Readings{}

	file, e := ioutil.ReadFile(path)
	check(e)

	dec := json.NewDecoder(strings.NewReader(string(file)))
	decErr := dec.Decode(&readingsFromJSONFile)
	check(decErr)

	indexOfReading := readingsFromJSONFile.pos(*r)

	// Remove the current reading from the slice of readings in the file
	readingsFromJSONFile = append(readingsFromJSONFile[:indexOfReading], readingsFromJSONFile[indexOfReading+1:]...)

	j, err := json.Marshal(readingsFromJSONFile)
	check(err)

	log.Print("Removing reading from backup file...")
	werr := writeJSONToFile(j)
	check(werr)
}

func (r *Reading) hasBackupReadingsInJSONFile() (bool, Readings) {
	readingsFromJSONFile := Readings{}
	path := fmt.Sprint(fileDir, backupFileName)
	file, e := ioutil.ReadFile(path)
	check(e)
	dec := json.NewDecoder(strings.NewReader(string(file)))
	decErr := dec.Decode(&readingsFromJSONFile)
	check(decErr)

	if len(readingsFromJSONFile) > 0 {
		return true, readingsFromJSONFile
	} else {
		return false, readingsFromJSONFile
	}
}

func (r *Reading) writeReadingToFile(dir string) {
	if _, err := os.Stat(dir); err != nil {
		if os.IsNotExist(err) {
			os.Mkdir(dir, 0755)
		} else {
			log.Println(err)
		}
	}

	path := fmt.Sprint(dir, backupFileName)

	file, e := ioutil.ReadFile(path)
	check(e)

	dec := json.NewDecoder(strings.NewReader(string(file)))
	decErr := dec.Decode(&seq)
	check(decErr)

	seq = append(seq, *r)

	j, err := json.Marshal(seq)
	check(err)

	werr := writeJSONToFile(j)
	check(werr)
}

func (r *Reading) setReadingTime() {
	r.Time = time.Now().Unix()
}

func (r *Reading) setHostname() {
	h, err := os.Hostname()
	check(err)
	r.Host = h
}

func (r *Reading) setValueFromSerial() {
	var success bool
	var retries int

	// Retry up to 5 times getting a correct reading from the serial pin
	for success != true && retries < 4 {
		retries++
		val := readSerialValue()
		// A correct reading should always look like this: 'R0569'
		// i.e. an "R", and 4 ints after the "R"
		match, _ := regexp.MatchString("^R[0-9]{4}", val)
		if match && len(val) == 5 {
			success = true
			r.Value = val
		} else {
			log.Print("Couldn't get a correct reading from sensor, retrying...")
		}
	}

	if success != true {
		panic("Was unable to get a correct reading from sensor")
	}
}

func (r *Reading) sendJSONToServer() (status string) {
	b, err := json.Marshal(r)

	req, err := http.NewRequest("POST", *serverURL, bytes.NewBuffer(b))
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	check(err)
	defer resp.Body.Close()

	return resp.Status
}
