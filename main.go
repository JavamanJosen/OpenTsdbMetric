package main

import (
	"OpenTsdbMetric/Report"
	"time"
)

func main()  {
	report := &Report.MonitorMessage{}
	locMsg := &Report.LocalMessage{}
	//openTsDb地址
	locMsg.OpenTsDbUrl = ""
	hostName, _ := Report.GetHostName()
	locMsg.Host = hostName
	locMsg.Period = 10
	report.Meter = "test.test_123"
	report.Value = locMsg

	report.Register()

	time.Sleep(1*time.Second)

	for i:=0; i<10; i++{
		for j:=0; j<1000000; j++{
			report.Report()
		}

		time.Sleep(1*time.Second)
		//x := rand.Intn(5)   //生成0-99随机整数

		//println(fmt.Sprintf("i = %d, report = %s, sleep = %d", i, resp, x))
		//time.Sleep(time.Duration(x) * time.Millisecond)
	}

	time.Sleep(10000 * time.Second)

}
