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

	//每次上报的计算机信息
	Value *LocalMessage `json:"value"`

	//mark
	Mark int64 `json:"mark"`
}

/**
打点机器的信息
*/
type LocalMessage struct {
	Tags map[string]interface{} `json:"tags"`

	Host string `json:"host"`

	//多久发送一次
	Period int64 `json:"period"`

	//记录上一次report的时间
	LastSendTs int64 `json:"last_send_ts"`

	//opentsdb 地址
	OpenTsDbUrl string `json:"open_ts_db_url"`
}

var (
	log = miner.Log()

	//用来盛放
	meterMap = map[string]*MonitorMessage{}
	//registerMuxMap sync.Map

	chReport = make(chan *MonitorMessage, 10000000)

	isStartPro = false
)

const (
	//默认20s提交一次
	PeriodDefault = 20
)

/**
注册打点机器信息
*/
func (msg *MonitorMessage) Register() string {

	meter := meterMap[msg.Meter]

	//如果发现没有注册，执行注册
	if meter == nil {
		meterStruct := &MonitorMessage{}

		locMsg := &LocalMessage{}
		locMsg.LastSendTs = time.Now().Unix()

		tags := make(map[string]interface{})
		tags["host"] = msg.Value.Host
		locMsg.Tags = tags

		if msg.Meter == "" || msg.Value.Host == "" {
			return "local message param is empty!"
		}

		period := msg.Value.Period
		if period < 1 {
			period = PeriodDefault
		}
		locMsg.Period = period
		locMsg.OpenTsDbUrl = msg.Value.OpenTsDbUrl

		meterStruct.Value = locMsg

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
		}
	}

}

/**
定时上报程序
*/
func timieReport() {

	for {
		go func() {
			for key, value := range meterMap {

				//满足上报条件，执行上报
				if (time.Now().Unix() - value.Value.LastSendTs) >= value.Value.Period {

					currTime := time.Now().Unix()

					//重置上报时间
					value.Value.LastSendTs = currTime

					reportStruct := make(map[string]interface{})
					reportStruct["metric"] = key
					reportStruct["timestamp"] = currTime
					reportStruct["value"] = value.Mark

					//把点数重置为0
					value.Mark = 0
					reportStruct["tags"] = value.Value.Tags

					body, err := json.Marshal(reportStruct)
					if err != nil {
						log.Error(err)
						return
					}

					log.Infof("metric = %s, ts = %d, body = %s", key, currTime, string(body))

					//上报信息
					cmd := fmt.Sprintf("/usr/bin/curl -i -X POST -d '%s' %s", string(body), value.Value.OpenTsDbUrl)
					out, err := exec.Command("bash", "-c", cmd).Output()
					if err != nil {
						log.Error(err)
					}
					log.Infof("report data %s = %s", key,string(out))
				}

			}
		}()
		log.Infof("sleep %d s", 1)
		time.Sleep(1 * time.Second)
	}
}

/**
获取计算机名字
*/
func GetHostName() (string, error) {
	return os.Hostname()
}
