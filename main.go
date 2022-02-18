package main

import (
	"bufio"
	"bytes"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/exec"
	"strings"
)

var logpath string
var collectHost string
var collectPort string

func init() {

	////////////////FILE confJWT/////
	paramsfile, err := os.Open("conf")
	if err != nil {
		log.Println("init", "failed opening file params: "+err.Error())
	}
	defer paramsfile.Close()

	sc := bufio.NewScanner(paramsfile)

	for sc.Scan() {
		str := sc.Text() // GET the line string
		if str == "" || str[0:1] == "#" {
			continue
		}

		nam := str[0:strings.Index(str, "=")]
		val := str[strings.Index(str, "=")+1:]
		log.Println("main", "nam", nam, "val", val)

		switch nam {
		case "logpath":
			logpath = val
		case "collectHost":
			collectHost = val
		case "collectPort":
			collectPort = val
		}
	}

	log.Println("main", "logpath=", logpath, "collectHost=", collectHost, "collectPort=", collectPort)

	if err := sc.Err(); err != nil {
		log.Println("init", "scan file error: "+err.Error())
	}

	//////////////////
}

func main() {
	fmt.Println("main", "started")
	defer fmt.Println("main", "finished")
	fmt.Println("main", "finished")

	cmd := exec.Command("tail", "-f", logpath)

	reader, err := cmd.StdoutPipe()
	if err != nil {
		return
	}

	scanner := bufio.NewScanner(reader)
	go func() {
		for scanner.Scan() {
			line := scanner.Text()
			fmt.Println("line =", line)
			sendLogRec(line)
		}
	}()

	err = cmd.Start()
	if err != nil {
		fmt.Println("start err =", err.Error())
		return
	}

	err = cmd.Wait()
	if err != nil {
		fmt.Println("wait err=", err.Error())
		return
	}
}

func sendLogRec(logrec string) {
	url := "http://" + collectHost + ":" + collectPort + "/inslog"
	fmt.Println("URL:>", url)

	var jsonStr = []byte(logrec)
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonStr))
	req.Header.Set("X-Custom-Header", "myvalue")
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()

	fmt.Println("response Status:", resp.Status)
	fmt.Println("response Headers:", resp.Header)
	body, _ := ioutil.ReadAll(resp.Body)
	fmt.Println("response Body:", string(body))
}
