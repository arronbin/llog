package main

import (
	"encoding/json"
	"fmt"
	"github.com/hpcloud/tail"
	"github.com/schoeu/gopsinfo"
	"github.com/schoeu/pslog_agent/util"
	"github.com/urfave/cli"
	"io"
	"io/ioutil"
	"os"
	"reflect"
)

type Config struct {
	AppId    string
	Secret   string
	LogDir   string
	Interval int
}

func main() {
	app := cli.NewApp()

	app.Version = "1.0.0"
	app.Name = "pslog_agent"
	app.Usage = "Agent for ps log"

	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name:  "config, c",
			Value: "",
			Usage: "configuration file path.",
		},
	}

	defer func() {
		if err := recover(); err != nil {
			fmt.Println(err)
		}
	}()

	app.Action = startAction
	app.Commands = []cli.Command{
		{
			Name:   "start",
			Usage:  "start app on agent.",
			Action: startAction,
		},
		{
			Name:   "stop",
			Usage:  "stop app on agent.",
			Action: stopAction,
		},
		{
			Name:   "status",
			Usage:  "show app status.",
			Action: statusAction,
		},
		{
			Name:    "remove",
			Aliases: []string{"rm"},
			Usage:   "remove app.",
			Action:  removeAction,
		},
	}

	err := app.Run(os.Args)
	util.ErrHandler(err)
}

func getConfig(p string) (Config, error) {
	p = util.GetAbsPath(util.GetHomeDir(), p)

	c := Config{}
	data, err := ioutil.ReadFile(p)
	err = json.Unmarshal(data, &c)

	return c, err
}

func removeAction(c *cli.Context) error {
	fmt.Println("removeAction")
	return nil
}

func statusAction(c *cli.Context) error {
	fmt.Println("statusAction")
	return nil
}

func startAction(c *cli.Context) error {
	fmt.Println("startAction")
	configFile := util.GetAbsPath("", c.Args().First())
	conf, err := getConfig(configFile)
	t, err := tail.TailFile(conf.LogDir, tail.Config{
		Location: &tail.SeekInfo{
			Whence: io.SeekEnd,
		},
		Follow: true,
	})
	for line := range t.Lines {
		psInfo := gopsinfo.GetPsInfo(100)
		var nodeInfo interface{}
		err = json.Unmarshal([]byte(line.Text), &nodeInfo)
		transJson(nodeInfo, psInfo)
		//agent.PushData(&psInfo,  conf.AppId, conf.Secret, )
	}
	return err
}
func transJson(inputVal interface{}, info gopsinfo.PsInfo) {
	fieldVal, ok := inputVal.(map[string]interface{})
	if !ok {
		panic("json unmarshal error.")
	}

	getType := reflect.TypeOf(info)
	getValue := reflect.ValueOf(info)
	rs := map[string]interface{}{}
	// 获取方法字段
	// 1. 先获取interface的reflect.Type，然后通过NumField进行遍历
	// 2. 再通过reflect.Type的Field获取其Field
	// 3. 最后通过Field的Interface()得到对应的value
	for i := 0; i < getType.NumField(); i++ {
		field := getType.Field(i)
		value := getValue.Field(i).Interface()
		//fmt.Printf("%s: %v = %v\n", field.Name, field.Type, value)
		if field.Name != "" {
			rs[field.Name] = value
		}
	}

	for i, v := range fieldVal {
		rs[i] = v
	}

	fmt.Println(rs)
}

func stopAction(c *cli.Context) error {
	fmt.Println("stopAction")
	return nil
}

//func defaultAction(c *cli.Context) error {
//	configFile := util.GetAbsPath("", c.Args().First())
//	ext := path.Ext(configFile)
//	if ext == ".json" {
//		conf, err := getConfig(configFile)
//		util.ErrHandler(err)
//		psInfoTimer(conf)
//
//	} else {
//		fmt.Println("Invited json file.")
//	}
//
//	return nil
//}
//
//func psInfoTimer(conf Config) {
//	d := time.Duration(time.Millisecond * time.Duration(conf.Interval))
//	t := time.NewTicker(d)
//	defer t.Stop()
//
//	for {
//		<-t.C
//		psInfo := gopsinfo.GetPsInfo(conf.Interval)
//		agent.PushData(&psInfo,  conf.AppId, conf.Secret)
//	}
//}
