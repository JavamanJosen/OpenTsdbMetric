package main

import (
	"OpenTsdbMetric/Report"
	"fmt"
	"time"
)

func main() {
	report := &Report.MonitorMessage{}
	report.OpenTsDbUrl = ""
	hostName, _ := Report.GetHostName()
	report.Host = hostName
	report.Period = 2
	report.Meter = "test.test_123"

	report.Register()

	time.Sleep(1 * time.Second)

	for i := 0; i < 10000; i++ {
		go func() {
			for j := 0; j < 500000; j++ {
				report.Report()
			}
		}()

		time.Sleep(1 * time.Second)

		println(fmt.Sprintf("报告 %d 点", i+1))
		//x := rand.Intn(5)   //生成0-99随机整数

		//println(fmt.Sprintf("i = %d, report = %s, sleep = %d", i, resp, x))
		//time.Sleep(time.Duration(x) * time.Millisecond)
	}

	time.Sleep(10000 * time.Second)

}
