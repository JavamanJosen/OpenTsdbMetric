package main

import (
	"OpenTsdbMetric/Report"
	"time"
)

func main() {

	factor := &Report.ReportFactor{}
	meter1 := "test.test_123"
	meter2 := "test.test_234"
	getReport()

	for aa := 0; aa< 20; aa++{
		for j := 0; j < 2000; j++ {
			go func() {
				factor.Report(meter1, 1)
				factor.Report(meter2, 1)
			}()
			time.Sleep(10 * time.Millisecond)
		}

		time.Sleep(20 * time.Second)
	}

	time.Sleep(100000 * time.Second)

}

func getReport() *Report.MonitorMessage {
	report := &Report.MonitorMessage{}
	report.OpenTsDbUrl = ""
	//report.OpenTsDbUrl = ""
	hostName, _ := Report.GetHostName()
	report.Host = hostName

	report.Register()

	return report
}
