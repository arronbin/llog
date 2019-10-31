package agent

import (
	"bytes"
	"errors"
	"fmt"

	"os"
	"path/filepath"
	"strconv"
	"time"

	jsoniter "github.com/json-iterator/go"
	"github.com/schoeu/llog/util"
)

type logStruct map[string]string

var apiServer, name string
var maxLinesDefault = 10
var json = jsoniter.ConfigCompatibleWithStandardLibrary

const errorType = "error"
const normalType = "normal"
const systemType = "system"

func fileGlob(sc *util.SingleConfig) {
	allLogs := sc.LogDir
	if len(allLogs) == 0 {
		logFileDir := util.GetTempDir()
		allLogs = append(allLogs, filepath.Join(logFileDir, util.LogDir, util.FilePattern))
	}

	// allLogs: - /var/logs/**/*.log
	for _, v := range allLogs {
		v = pathPreProcess(v)
		// paths: ["/var/logs/1.log","/var/logs/2.log"]
		p, err := filepath.Glob(v)
		util.ErrHandler(err)
		if len(p) > 0 {
			initState(p, sc)
		}
		// update file state.
	}
}

func pathPreProcess(p string) string {
	exist, err := util.PathExist(p)
	util.ErrHandler(err)
	if !exist {
		err = os.Mkdir(filepath.Dir(p), os.ModePerm)
	}
	if !filepath.IsAbs(p) {
		p = util.GetAbsPath("", p)
	}
	return p
}

var buf = bytes.Buffer{}
var count = 0

func lineFilter(k string) func(*[]byte) {
	fi, err := getLogInfoIns(k)
	util.ErrHandler(err)
	if fi != nil {
		sc := fi.sc

		include, exclude, multiline := sc.Include, sc.Exclude, sc.Multiline.Pattern
		confMaxByte, maxLines := sc.MaxBytes, sc.Multiline.MaxLines

		if maxLines == 0 {
			maxLines = maxLinesDefault
		}

		return func(l *[]byte) {
			line := *l
			// multiple mode
			if multiline != "" {
				// multiple head line
				if util.IsInclude(line, []string{multiline}) {
					if buf.Len() > 0 {
						ok, rs := filter(include, exclude, buf.Bytes(), confMaxByte)
						if ok {
							return
						}
						doPush(rs, errorType)
						count = 0
						buf = bytes.Buffer{}
					}
				}
				count++
				// 匹配多行其他内容
				if count < maxLines {
					//logContent.Write(line)
					buf.Write(line)
				}
			} else {
				ok, rs := filter(include, exclude, line, confMaxByte)
				if ok {
					return
				}
				doPush(rs, normalType)
			}
		}
	}
	return nil
}

func filter(include, exclude []string, line []byte, max int) (bool, *[]byte) {
	if len(include) > 0 && !util.IsInclude(line, include) {
		return true, nil
	}
	if len(exclude) > 0 && util.IsInclude(line, exclude) {
		return true, nil
	}
	if max != 0 && len(line) > max {
		line = line[:max]
	}
	return false, &line
}

func doPush(text *[]byte, types string) {
	// 日志签名
	var rs = logStruct{
		"@message":    string(*text),
		"@version":    util.Version,
		"@logId":      util.UUID(),
		"@timestamps": strconv.FormatInt(time.Now().UnixNano()/1e6, 10),
		"@types":      types,
		"@name":       name,
	}

	if apiServer != "" {
		go apiPush(&rs)
	}
	if indexServer != nil {
		go esPush(&rs)
	}
}

func getLogInfoIns(p string) (*logInfo, error) {
	logContent, ok := sm.Get(p)
	if !ok {
		return nil, errors.New(fmt.Sprintf("file: %s is not exist in sync map", p))
	}
	li := logContent.(logInfo)
	return &li, nil
}
