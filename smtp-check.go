package main

import (
	"encoding/json"
	"fmt"
	"flag"
	"io/ioutil"
	"net"
	"net/smtp"
	"os"
	"strings"
	"time"
)

type CheckJob struct {
	DomainPart string
	LocalParts []string
}

type CheckResult struct {
	LocalPart     string `json: "localPart"`
	DomainPart    string `json: "domainPart"`
	Email         string `json: "email"`
	Verified      bool   `json: "verified"`
	ExecutionTime string `json: "executionTime"`
	MxFound       bool   `json: "mxFound"`
}

type Response struct {
	Error         bool
	Message       string
	CheckResults  []CheckResult
	ExecutionTime string `json: "executionTime"`
}

func MxLookup(host string) bool {
	mxServers, err := net.LookupMX(host)

	if err != nil {
		return false
	}

	for _, mxServer := range mxServers {
		MxMap[host] = append(MxMap[host], *mxServer)
	}

	return true
}

// process the domain group, domain group contains all the mails for the certain domain
func processDomainGroup(jobs chan CheckJob, results chan CheckResult) {

	for checkJob := range jobs {

		domainPart, localParts := checkJob.DomainPart, checkJob.LocalParts

		if _, ok := MxMap[domainPart]; !ok {
			MxLookup(domainPart)
		}

		if len(MxMap[domainPart]) > 0 {
			for _, localPart := range localParts {
				start := time.Now()

				checkResult := CheckResult{LocalPart: localPart, DomainPart: domainPart, Email: localPart + "@" + domainPart, MxFound: true}

				smtpHost := MxMap[domainPart][0].Host
				smtpPort := "25"

				// to support timeout
				timeout, _ := time.ParseDuration("10s")
				Conn, err := net.DialTimeout("tcp", smtpHost+":"+smtpPort, timeout)

				Client, err := smtp.NewClient(Conn, smtpHost)

				if err != nil {
					//fmt.Println("Connection error: ", err)

					checkResult.ExecutionTime = time.Since(start).String()
					checkResult.Verified = false
					results <- checkResult

					continue
				}

				email := localPart + "@" + domainPart

				err = Client.Hello("hi")

				if err != nil {
					//fmt.Println("Hello error:", err)

					checkResult.ExecutionTime = time.Since(start).String()
					checkResult.Verified = false
					results <- checkResult

					continue
				}
				err = Client.Mail(FromMail)

				if err != nil {
					//fmt.Println("From mail error:", err)

					checkResult.ExecutionTime = time.Since(start).String()
					checkResult.Verified = false
					results <- checkResult

					continue
				}
				err = Client.Rcpt(email)

				if err != nil {
					//fmt.Println(err)
					//fmt.Println(email+" is not verified")

					checkResult.Verified = false
				} else {
					//fmt.Println(email+" is verified")

					checkResult.Verified = true
				}

				Client.Quit()
				Client.Close()

				checkResult.ExecutionTime = time.Since(start).String()
				results <- checkResult
			}
		} else {
			for _, localPart := range localParts {

				start := time.Now()

				checkResult := CheckResult{LocalPart: localPart, DomainPart: domainPart, Email: localPart + "@" + domainPart, Verified: false, MxFound: false}

				checkResult.ExecutionTime = time.Since(start).String()

				results <- checkResult
			}
		}
	}
}

// parse lines array to the GroupedInput format
func parseGroupedInput(Lines []string) (map[string][]string, int) {
	var TotalEmails int = 0
	var GroupedInput map[string][]string = make(map[string][]string)

	for _, value := range Lines {
		var LocalPart string
		var DomainPart string
		var Parts []string

		Parts = strings.Split(value, "@")

		if len(Parts) == 2 {
			LocalPart, DomainPart = Parts[0], Parts[1]

			GroupedInput[DomainPart] = append(GroupedInput[DomainPart], LocalPart)

			TotalEmails++
		} else {
			// skip this line
		}
	}

	return GroupedInput, TotalEmails
}

var Filename string
var MaxGoRoutines int
var FromMail string

var MxMap map[string][]net.MX

// set the flags from the command line, or, from the default values
func init() {
	FilenameFlag := flag.String("filename", "", "name of the file, which contains emails")
	MaxGoRoutinesFlag := flag.Int("max-go-routines", 3, "max Go routines")
	FromMailFlag := flag.String("from-email", "test@test.com", "from email")

	flag.Parse()

	MxMap = make(map[string][]net.MX)

	Filename = *FilenameFlag
	MaxGoRoutines = *MaxGoRoutinesFlag
	FromMail = *FromMailFlag
}

func main() {
	var err error
	var GroupedInput map[string][]string = make(map[string][]string)
	var TotalEmails int = 0

	start := time.Now()

	// prepare the Response
	response := Response{CheckResults: make([]CheckResult, 0)}

	// call the function when everything is done
	defer func() {
		response.ExecutionTime = time.Since(start).String()

		responseText, _ := json.MarshalIndent(response, "", "    ")

		// print response
		fmt.Println(string(responseText))
	}()

	if(len(Filename) == 0) {
		response.Error = true
		response.Message = "Filename should be provided"

		return
	}

	// check if the file exists
	if _, err = os.Stat(Filename); os.IsNotExist(err) {
		response.Error = true
		response.Message = fmt.Sprintf("File %s doesn't exist", Filename)

		return
	}

	Content, err := ioutil.ReadFile(Filename)

	if err != nil {
		response.Error = true
		response.Message = "Error during the file read"

		return
	}

	// split by domain groups
	var Lines []string = strings.Split(string(Content), "\n")

	// parse file contents to the group input format
	GroupedInput, TotalEmails = parseGroupedInput(Lines)

	// prepare the jobs and the results channels
	jobs := make(chan CheckJob, len(GroupedInput))
	results := make(chan CheckResult, TotalEmails)

	// set the correct max go routines value
	if len(GroupedInput) < MaxGoRoutines {
		MaxGoRoutines = len(GroupedInput)
	}

	// launch workers
	for i := 0; i < MaxGoRoutines; i++ {
		go processDomainGroup(jobs, results)
	}

	// push jobs
	for domainPart, localParts := range GroupedInput {
		jobs <- CheckJob{DomainPart: domainPart, LocalParts: localParts}
	}

	// get the results
	for i := 0; i < TotalEmails; i++ {
		result := <-results
		response.CheckResults = append(response.CheckResults, result)
	}

	// finally we're sure that the run is correct
	response.Error = false

}
