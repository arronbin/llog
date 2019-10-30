package agent

import (
	"encoding/json"
	"time"

	"github.com/schoeu/gopsinfo"
	"github.com/schoeu/llog/util"
)

const aliveTimeDefault = 300

func closeFileHandle(sc *util.SingleConfig) {
	defer util.Recover()

	aliveTime := sc.CloseInactive
	if aliveTime < 1 {
		aliveTime = aliveTimeDefault
	}
	ticker := time.NewTicker(time.Duration(aliveTime) * time.Second)
	for {
		<-ticker.C
		for _, v := range sm.Keys() {
			li := getLogInfoIns(v)
			if li.sc == sc && time.Since(time.Unix(li.status[1], 0)) > time.Second*time.Duration(aliveTime) {
				delInfo(v)
			}
		}
	}
}

func sysInfo() {
	conf := util.GetConfig()
	info := conf.SysInfo

	if info {
		during := conf.SysInfoDuring
		var psInfo gopsinfo.PsInfo
		var d time.Duration
		if during < 1 {
			d = 1
		} else if during == 0 {
			d = 10
		}
		ticker := time.NewTicker(d * time.Second)
		go func() {
			for {
				<-ticker.C

				psInfo = gopsinfo.GetPsInfo(d)
				sysData, err := json.Marshal(psInfo)
				util.ErrHandler(err)
				doPush(&sysData, systemType)
			}
		}()
	}
}
