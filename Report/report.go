package Report

import (
	"encoding/json"
	"fmt"
	"github.com/hunterhug/marmot/miner"
	"os"
	"os/exec"
	"time"
)

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
	LastMarkSendTs    int64 `json:"last_send_ts"`
	LastMarkM1SendTs  int64 `json:"last_mark_m_1_send_ts"`
	LastMarkM5SendTs  int64 `json:"last_mark_m_5_send_ts"`
	LastMarkM15SendTs int64 `json:"last_mark_m_15_send_ts"`

	//mark
	Mark    int64 `json:"mark"`
	MarkM1  int64 `json:"mark_m_1"`
	MarkM5  int64 `json:"mark_m_5"`
	MarkM15 int64 `json:"mark_m_15"`

	//Lock *sync.Mutex
}

var (
	log = miner.Log()

	//用来盛放
	meterMap = map[string]*MonitorMessage{}
	//registerMuxMap sync.Map

	chReport = make(chan *MonitorMessage, 10000000)

	isStartPro = false

	isReport = true
)

const (
	//默认20s提交一次
	PeriodDefault = 20
)

/**
注册打点机器信息
*/
func (msg *MonitorMessage) Register() string {

	// qps/1m/5m/10m/30m
	meter := meterMap[msg.Meter]

	//如果发现没有注册，执行注册
	if meter == nil {
		meterStruct := &MonitorMessage{}

		currTime := time.Now().Unix()
		meterStruct.LastMarkSendTs = currTime
		meterStruct.LastMarkM1SendTs = currTime
		meterStruct.LastMarkM5SendTs = currTime
		meterStruct.LastMarkM15SendTs = currTime

		tags := make(map[string]interface{})
		tags["host"] = msg.Host
		meterStruct.Tags = tags

		if msg.Meter == "" || msg.Host == "" {
			return "local message param is empty!"
		}

		period := msg.Period
		if period < 1 {
			period = PeriodDefault
		}
		meterStruct.Period = period
		meterStruct.OpenTsDbUrl = msg.OpenTsDbUrl

		//Lock = new(sync.Mutex)

		meterMap[msg.Meter] = meterStruct
	}

	//程序第一次启动，启动消费和定时上报程序
	if isStartPro == false {

		//启动消费程序
		go consumerReport()

		//启动定时上报程序
		go timieReport()

		isStartPro = true
	}

	return "Register monitor success!"
}

/**
上报信息入口
*/
func (monitor *MonitorMessage) Report() string {

	if monitor.Meter == "" {
		return "monitor message meter is empty!"
	}

	chReport <- monitor

	return "success";
}

/**
从管道里面消费请求
*/
func consumerReport() {

	for {
		monitor := <-chReport

		if meterMap[monitor.Meter] == nil {
			log.Errorf("meter = %s, 未注册！！！")
			continue
		} else {
			monMessage := meterMap[monitor.Meter]
			monMessage.Mark = monMessage.Mark + 1
			monMessage.MarkM1 = monMessage.MarkM1 + 1
			monMessage.MarkM5 = monMessage.MarkM5 + 1
			monMessage.MarkM15 = monMessage.MarkM15 + 1

			//检查是否满足上报，由timieReport控制是否要上报
			if isReport {
				log.Infof("meterMap size is %d", len(meterMap))
				for key, value := range meterMap {
					log.Infof("key = %s", key)
					currTime := time.Now().Unix()
					//满足上报条件，执行上报
					if (currTime - value.LastMarkSendTs) >= value.Period {

						qps := value.Mark
						markM1 := value.MarkM1
						markM5 := value.MarkM5
						markM15 := value.MarkM15

						//把打点的值置为0
						if (currTime - value.LastMarkSendTs) >= value.Period {
							value.Mark = 0
							value.LastMarkSendTs = currTime
						}
						if (currTime - value.LastMarkM1SendTs) >= 60 {
							value.MarkM1 = 0
							value.LastMarkM1SendTs = currTime
						}
						if (currTime - value.LastMarkM5SendTs) >= 300 {
							value.MarkM5 = 0
							value.LastMarkM5SendTs = currTime
						}
						if (currTime - value.LastMarkM15SendTs) >= 900 {
							value.MarkM15 = 0
							value.LastMarkM15SendTs = currTime
						}

						//重置上报时间
						value.LastMarkSendTs = currTime

						for _, meterKey := range []string{".qps", ".m1", ".m5", ".m15"} {

							//go func(meterKey string) {
							reportStruct := make(map[string]interface{})
							reportStruct["metric"] = key + meterKey
							reportStruct["timestamp"] = currTime
							if meterKey == ".qps" {
								reportStruct["value"] = qps
							} else if meterKey == ".m1" { //m1的值是每过一分钟清空一次，下面一次类推
								reportStruct["value"] = markM1
							} else if meterKey == ".m5" {
								reportStruct["value"] = markM5
							} else if meterKey == ".m15" {
								reportStruct["value"] = markM15
							}

							reportStruct["tags"] = value.Tags

							body, err := json.Marshal(reportStruct)
							if err != nil {
								log.Error(err)
								return
							}
							log.Infof("metric = %s, ts = %d, body = %s", key, currTime, string(body))
							//上报信息
							go func(body []byte) {
								cmd := fmt.Sprintf("/usr/bin/curl -i -X POST -d '%s' %s", string(body), value.OpenTsDbUrl)
								exec.Command("bash", "-c", cmd).Output()
							}(body)

							//}(meterKey)
						}

					}

				}
				isReport = false
			}
		}
	}

}

/**
定时上报程序
*/
func timieReport() {

	for {
		isReport = true
		time.Sleep(1 * time.Second)
	}
}

/**
获取计算机名字
*/
func GetHostName() (string, error) {
	return os.Hostname()
}
