package main

import (
	"testing"
	. "github.com/smartystreets/goconvey/convey"
	"strings"
)

func TestSmtpCheck(t *testing.T) {

	inputEmails := `ka-zaido@yandex.ru
ka-zaido2222@yandex.ru
pavel@kredito.de
pavel2222@kredito.de
####@$$$$$

hello@hello44444444.io`

	var GroupedInput map[string][]string = make(map[string][]string)
	var TotalEmails int = 0

	Convey("parse input", t, func(){

			var Lines []string = strings.Split(string(inputEmails), "\n")

			GroupedInput, TotalEmails = parseGroupedInput(Lines)

			So(len(GroupedInput["yandex.ru"]), ShouldEqual, 2)
			So(TotalEmails, ShouldEqual, 6)
		})

	Convey("check emails", t, func() {
			// prepare the jobs and the results channels
			jobs := make(chan CheckJob, len(GroupedInput))
			results := make(chan CheckResult, TotalEmails)

			var checkResults []CheckResult = make([]CheckResult, 0)

			// launch workers
			for i := 0; i < 2; i++ {
				go processDomainGroup(jobs, results)
			}

			// push jobs
			for domainPart, localParts := range GroupedInput {
				jobs <- CheckJob{DomainPart: domainPart, LocalParts: localParts}
			}

			// get the results
			for i := 0; i < TotalEmails; i++ {
				result := <-results

				checkResults = append(checkResults, result)
			}

			// checking the results with the expected results
			for _, checkResult := range checkResults {
				switch checkResult.Email {
				case "ka-zaido@yandex.ru", "pavel@kredito.de":
					So(checkResult.Verified, ShouldEqual, true)
				default:
					So(checkResult.Verified, ShouldEqual, false)
				}
			}
		})
}
