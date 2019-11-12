package Report

import (
	"encoding/json"
	"github.com/hunterhug/marmot/miner"
	"os"
	"sync"
	"time"
)

/**
由ReportFactor工厂结构体统一report
*/
type ReportFactor struct {
}

/**
grafana 计算指定时间内打点个数
*/
type MonitorMessage struct {
	Meter string `json:"meter"`

	Tags map[string]interface{} `json:"tags"`

	Host string `json:"host"`

	//多久上报一次
	Period int64 `json:"period"`

	//opentsdb 地址
	OpenTsDbUrl string `json:"open_ts_db_url"`

	//记录上一次report的时间
	LastMarkSendTs int64 `json:"last_send_ts"`

	TimeWindow map[int64]int64 `json:"time_window"`
}

var (
	log = miner.Log()

	//用来盛放
	meterMap = map[string]*MonitorMessage{}

	//初始化 基本信息
	autoStruct = &MonitorMessage{}

	chReport = make(chan *MonitorMessage, 10000)

	isStartPro = false

	isReport = true

	lock  sync.Locker

)

const (
	//默认20s提交一次
	PeriodDefault = 20
	cleanIndex    = 1200
)

/**
注册打点机器信息
*/
func (msg *MonitorMessage) Register() string {

	//如果meter 是空，说明是首次注册，只需要把基本信息放入缓存即可
	if msg.Meter == "" {
		//初始化 锁
		lock = new(sync.Mutex)

		tags := make(map[string]interface{})
		host, _ := GetHostName()
		tags["host"] = host
		autoStruct.Tags = tags
		autoStruct.Host = msg.Host
		autoStruct.OpenTsDbUrl = msg.OpenTsDbUrl

		//程序第一次启动，启动消费和定时上报程序
		if isStartPro == false {

			//启动消费程序
			go consumerReport()

			//启动定时上报程序
			go timieReport()

			isStartPro = true
		}

	} else {//非首次注册只需要把 非基本信息祖册到meterMap即可
		meterStruct := &MonitorMessage{}

		meterStruct.Tags = autoStruct.Tags
		meterStruct.OpenTsDbUrl = autoStruct.OpenTsDbUrl

		period := msg.Period
		if period <= 0 {
			period = PeriodDefault
		}
		meterStruct.Period = period

		meterStruct.TimeWindow = make(map[int64]int64)
		currTime := time.Now().Unix()
		meterStruct.LastMarkSendTs = currTime

		meterStruct.Meter = msg.Meter

		meterMap[msg.Meter] = meterStruct
	}

	return "Register monitor success!"
}

/**
根据监控信息上报信息入口
*/
func (monitor *MonitorMessage) Report() string {

	if monitor.Meter == "" {
		return "monitor message meter is empty!"
	}

	chReport <- monitor

	return "success";
}

func (rf *ReportFactor) Report(meter string, period int64) string {

	if meterStruct := meterMap[meter]; meterStruct == nil {

		newMonitor := &MonitorMessage{}
		newMonitor.Meter = meter
		newMonitor.Period = period

		reMsg := newMonitor.Register()
		log.Infof("reMsg = %s", reMsg)
	}

	monitor := meterMap[meter]

	return monitor.Report()
}

/**
从管道里面消费请求
*/
func consumerReport() {

	for {

		monitor := <-chReport

		go func() {
			lock.Lock()
			defer lock.Unlock()

			monMessage := meterMap[monitor.Meter]

			//当前时间窗口+1
			monMessage.add()
		}()

	}

}

/**
上报打点信息
*/
func (mm *MonitorMessage) ReportMonter(key, meterKey string, currTime int64) {

	reportStruct := make(map[string]interface{})
	reportStruct["metric"] = key + meterKey
	reportStruct["timestamp"] = currTime

	if meterKey == ".qps" {
		reportStruct["value"] = mm.TimeWindow[currTime]
	} else if meterKey == ".m1" { //m1的值是每过一分钟清空一次，下面一次类推
		reportStruct["value"] = mm.getMnCount(currTime, 60)
	} else if meterKey == ".m5" {
		reportStruct["value"] = mm.getMnCount(currTime, 300)
	} else if meterKey == ".m15" {
		reportStruct["value"] = mm.getMnCount(currTime, 900)
	}

	reportStruct["tags"] = mm.Tags

	body, err := json.Marshal(reportStruct)
	if err != nil {
		log.Error(err)
		return
	}
	log.Infof("metric = %s, ts = %d, body = %s", key, currTime, string(body))
	//上报信息
	//go func(body []byte) {
	//	cmd := fmt.Sprintf("/usr/bin/curl -i -X POST -d '%s' %s", string(body), mm.OpenTsDbUrl)
	//	exec.Command("bash", "-c", cmd).Output()
	//}(body)

}

/**
定时上报程序
*/
func timieReport() {

	for {
		//isReport = true
		time.Sleep(1 * time.Second)

		go func() {
			lock.Lock()
			defer lock.Unlock()

			for key, value := range meterMap {

				currTime := time.Now().Unix()

				//qps每秒上报一次
				value.ReportMonter(key, ".qps", currTime-1)

				//满足上报条件，执行上报
				if (currTime - value.LastMarkSendTs) >= value.Period {

					//重置上报时间
					value.LastMarkSendTs = currTime

					for _, meterKey := range []string{".m1", ".m5", ".m15"} {

						value.ReportMonter(key, meterKey, currTime)

					}

				}

				//清理时间窗口数据
				value.deleteTimeWindows(currTime-cleanIndex-value.Period, value.Period)

			}
		}()
	}
}

/**
获取计算机名字
*/
func GetHostName() (string, error) {
	return os.Hostname()
}

/**
滑动时间窗口计数加一
*/
func (mm *MonitorMessage) add() {
	nowT := time.Now().Unix()
	nowTCount := mm.TimeWindow[nowT]
	if nowTCount == 0 {
		mm.TimeWindow[nowT] = 1
	} else {
		mm.TimeWindow[nowT] = nowTCount + 1
	}
	//log.Infof("tag = %s, count = %d", mm.Meter, mm.TimeWindow[nowT])
}

/**
获取Mx的点数
*/
func (mm *MonitorMessage) getMnCount(nowTime, n int64) int64 {
	var reCount int64
	MinTime := nowTime - n
	for i := MinTime; i <= nowTime; i++ {
		reCount += mm.TimeWindow[i]
	}

	return reCount
}

/**
删除时间窗口中的无用元素
*/
func (mm *MonitorMessage) deleteTimeWindows(startTime, limit int64) {

	for i := startTime; i < startTime+limit; i++ {
		delete(mm.TimeWindow, i)
		//log.Infof("delete %d 的数据,还剩 %d 个元素", i, len(mm.TimeWindow))
	}

}
