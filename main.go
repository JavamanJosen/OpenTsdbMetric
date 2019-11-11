package main

import (
	"OpenTsdbMetric/Report"
	"fmt"
	"time"
)

func main() {

	factor := &Report.ReportFactor{}
	meter1 := "test.test_123"
	meter2 := "test.test_234"
	meter3 := "test.test_345"
	meter4 := "test.test_456"
	meter5 := "test.test_567"
	report := getReport(meter1)
	println(report)
	getReport(meter2)
	getReport(meter3)
	getReport(meter4)
	getReport(meter5)
	time.Sleep(1 * time.Second)

	for i := 0; i < 10000; i++ {
		go func() {
			for j := 0; j < 10000; j++ {
				go func() {
					//report1.Report()
					factor.Report(meter1)
				}()
			}

			for j := 0; j < 20000; j++ {
				go func() {
					//report2.Report()
					factor.Report(meter2)
				}()
			}

			for j := 0; j < 30000; j++ {
				go func() {
					//report3.Report()
					factor.Report(meter3)
				}()
			}

			for j := 0; j < 40000; j++ {
				go func() {
					//report4.Report()
					factor.Report(meter4)
				}()
			}

			for j := 0; j < 50000; j++ {
				go func() {
					//report5.Report()
					factor.Report(meter5)
				}()
			}
		}()

		time.Sleep(1 * time.Second)

		println(fmt.Sprintf("报告 %d 点", i+1))
		//x := rand.Intn(5)   //生成0-99随机整数

		//println(fmt.Sprintf("i = %d, report = %s, sleep = %d", i, resp, x))
		//time.Sleep(time.Duration(x) * time.Millisecond)
		if (i+1)%1000 == 0 {
			time.Sleep(time.Second * 100)
		}
	}

	time.Sleep(100000 * time.Second)

}

func getReport(meter string) *Report.MonitorMessage {
	report := &Report.MonitorMessage{}
	report.OpenTsDbUrl = ""
	//report.OpenTsDbUrl = ""
	hostName, _ := Report.GetHostName()
	report.Host = hostName
	report.Period = 1
	report.Meter = meter

	report.Register()

	return report
}
