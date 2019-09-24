package agent

import (
	"encoding/json"
	"io"
	"os"
	"path"
	"path/filepath"
	"time"

	"github.com/hpcloud/tail"
	"github.com/schoeu/gopsinfo"
	"github.com/schoeu/nma/util"
	"github.com/urfave/cli"
)

func StartAction(c *cli.Context) error {
	configFile := util.GetAbsPath(util.GetCwd(), c.Args().First())
	conf, err := util.GetConfig(configFile)
	logFile := conf.LogDir

	if logFile == "" {
		//logFileDir:= util.GetHomeDir()
		logFileDir := util.GetTempDir()
		logFile = path.Join(logFileDir, util.LogDir)

		exist, err := util.PathExist(logFile)
		util.ErrHandler(err)
		if !exist {
			err = os.Mkdir(logFile, os.ModePerm)
			util.ErrHandler(err)
		}
	}

	logChan := make(chan int)

	paths, err := filepath.Glob(path.Join(logFile, util.FilePattern))
	for _, v := range paths {
		go pushLog(v, conf)
	}

	util.ErrHandler(err)
	// 阻塞主goroutines
	<-logChan
	return err
}

func pushLog(logFile string, conf util.Config) {
	t, err := tail.TailFile(logFile, tail.Config{
		Location: &tail.SeekInfo{
			Whence: io.SeekEnd,
		},
		Follow: true,
	})
	st := time.Now()

	logServer := util.LogServer
	if conf.LogServer != "" {
		logServer = conf.LogServer
	}

	for line := range t.Lines {
		var psInfo gopsinfo.PsInfo
		if !conf.NoSysInfo {
			et := time.Now()
			during := et.Sub(st)
			timeSub := int(during)
			if timeSub < 1 {
				during = time.Microsecond * 1000
			}
			psInfo = gopsinfo.GetPsInfo(during)
			st = et
		}
		var nodeInfo interface{}
		err = json.Unmarshal([]byte(line.Text), &nodeInfo)
		combineRs := util.CombineData(nodeInfo, psInfo, conf.NoSysInfo)
		if logServer != "" {
			PushData(combineRs, logServer)
		}
	}
	util.ErrHandler(err)
}
