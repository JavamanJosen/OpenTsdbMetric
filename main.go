package main

import (
	"OpenTsdbMetric/Report"
	"fmt"
	"time"
)

func main() {

	meter1 := "test.test_123"
	meter2 := "test.test_234"
	meter3 := "test.test_345"
	meter4 := "test.test_456"
	meter5 := "test.test_567"
	report1 := getReport(meter1)
	report2 := getReport(meter2)
	report3 := getReport(meter3)
	report4 := getReport(meter4)
	report5 := getReport(meter5)
	time.Sleep(1 * time.Second)

	for i := 0; i < 10000; i++ {
		go func() {
			for j := 0; j < 10000; j++ {
				report1.Report()
			}

			for j := 0; j < 20000; j++ {
				report2.Report()
			}


			for j := 0; j < 30000; j++ {
				report3.Report()
			}

			for j := 0; j < 40000; j++ {
				report4.Report()
			}

			for j := 0; j < 50000; j++ {
				report5.Report()
			}
		}()

		time.Sleep(1 * time.Second)

		println(fmt.Sprintf("报告 %d 点", i+1))
		//x := rand.Intn(5)   //生成0-99随机整数

		//println(fmt.Sprintf("i = %d, report = %s, sleep = %d", i, resp, x))
		//time.Sleep(time.Duration(x) * time.Millisecond)
	}

	time.Sleep(100000 * time.Second)

}

func getReport(meter string) *Report.MonitorMessage {
	report := &Report.MonitorMessage{}
	report.OpenTsDbUrl = ""
	hostName, _ := Report.GetHostName()
	report.Host = hostName
	report.Period = 10
	report.Meter = meter

	report.Register()

	return report
}
