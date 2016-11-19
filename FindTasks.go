// This is a program to find the tasks of a Marathon deployment and all of
// its addresses. It works by simply asking Marathon.
//
// Usage: FindTasks [ -marathon <Marathon-URL> ] [ -ids ] [ -minimum <nr> ] 
//                  <appId>
//
// Output: 
//   <host>:<ports>:<id>:<slaveid>
// or
//   <host>:<ports>
// one line per task found.


package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"time"
)

var marathonURL string 
var showIds bool
var minimum int
var retries int
var oneLine bool
var option string
var prefix string

func main() {
	// Parse command line:
	flag.StringVar(&marathonURL, "marathon", "http://marathon.mesos:8080",
	               "Marathon URL")
	flag.BoolVar(&showIds, "ids", false, "show task and agent id")
	flag.IntVar(&minimum, "minimum", 0,
	            "repeat until this amount of tasks there")
	flag.IntVar(&retries, "retries", 10, "number of retries")
	flag.BoolVar(&oneLine, "oneline", false, "result on one line")
	flag.StringVar(&option, "option", "", "option for one line")
	flag.StringVar(&prefix, "prefix", "", "prefix for one line")
	flag.Parse()
	if option != "" || prefix != "" {
		oneLine = true
	}
	args := flag.Args()
	if len(args) != 1 {
		fmt.Fprintln(os.Stderr, "Need an appId as command line argument")
		return
	}
	appId := args[0]

	for count := 1; count <= retries; count++ {
		resp, err := http.Get(marathonURL + "/v2/apps/" + appId)
		if err != nil || resp.Body == nil || resp == nil{
			fmt.Fprintln(os.Stderr, "Error querying Marathon:", err, resp,
								   "retry", count, "out of", retries)
			if count >= retries {
				return
			}
		} else {
			defer resp.Body.Close()
			respBody, _ := ioutil.ReadAll(resp.Body)
			var result map[string]map[string]interface{}
			json.Unmarshal(respBody, &result)
			taskArray := result["app"]["tasks"].([]interface{})
			if len(taskArray) >= minimum {
				for i := 0; i < len(taskArray); i++ {
					task := taskArray[i].(map[string]interface{})
					host := task["host"].(string)
					ports := task["ports"].([]interface{})
					id := task["id"].(string)
					slaveId := task["slaveId"].(string)
					if !oneLine {
						fmt.Printf("%s:", host)
						for j := 0; j < len(ports); j++ {
							if j > 0 {
								fmt.Printf(",")
							}
							if j > 0 {
								fmt.Printf(",%d", int(ports[j].(float64)))
							} else {
								fmt.Printf("%d", int(ports[j].(float64)))
							}
						}
						if showIds {
							fmt.Printf(":%s:%s", id, slaveId)
						}
						fmt.Println()
					} else {  // oneLine == true
						if option != "" {
							fmt.Printf(" %s", option)
						}
						fmt.Printf(" %s%s:", prefix, host)
						for j := 0; j < len(ports); j++ {
							if j > 0 {
								fmt.Printf(",%d", int(ports[j].(float64)))
							} else {
								fmt.Printf("%d", int(ports[j].(float64)))
							}
						}
					}
				}
				return
			}
			fmt.Fprintln(os.Stderr, "Found only", len(taskArray), "instead of",
									 minimum, "tasks, waiting...")
		}
		time.Sleep(1000000000)
	}
}

