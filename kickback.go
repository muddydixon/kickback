package main

import (
	"bytes"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"reflect"
	"strings"
	"text/template"
	"time"

	"gopkg.in/yaml.v2"

	"github.com/Sirupsen/logrus"
	"github.com/gin-gonic/gin"
)

var Version = "0.1.0"

type Config struct {
	Tasks []Task
	Log   map[string]string
	Port  int32
}

type Task struct {
	Name   string
	Path   string
	Method string
	Procs  []string
}

var config *Config
var accessLogger *logrus.Logger
var systemLogger *logrus.Logger

func addTask(r *gin.Engine, task Task) {
	method := strings.ToUpper(task.Method)
	api := reflect.ValueOf(r).MethodByName(method)
	api.Call([]reflect.Value{reflect.ValueOf(task.Path), reflect.ValueOf(execProcsGen(task))})
}

func getTasks(c *gin.Context) {
	c.JSON(200, gin.H{"tasks": config.Tasks})
}

func execProcsGen(task Task) gin.HandlerFunc {
	return func(c *gin.Context) {
		var err error
		var outs []string

		data := make(map[string]string)
		x, _ := ioutil.ReadAll(c.Request.Body)
		for _, b := range strings.Split(string(x), "&") {
			if b != "" {
				kv := strings.Split(b, "=")
				data[strings.ToUpper(kv[0])] = kv[1]
			}
		}
		for _, q := range strings.Split(c.Request.URL.RawQuery, "&") {
			if q != "" {
				kv := strings.Split(q, "=")
				data[strings.ToUpper(kv[0])] = kv[1]
			}
		}
		c.Set("task", task)
		c.Set("data", data)

		for _, proc := range task.Procs {
			t := template.New("procs")
			tmpl, _ := t.Parse(proc)
			var b bytes.Buffer
			var out bytes.Buffer

			tmpl.Execute(&b, data)
			cmd := exec.Command("sh", "-c", b.String())
			cmd.Stdout = &out
			err = cmd.Run()
			if err != nil {
				systemLogger.WithFields(logrus.Fields{
					"name":   task.Name,
					"method": task.Method,
					"proc":   b.String(),
				}).Info("some error")
			}
			outs = append(outs, out.String())
		}
		c.JSON(200, gin.H{"stdout": strings.Join(outs, "\n")})
	}
}

func readConfig(confFile string) {
	data, err := ioutil.ReadFile(confFile)
	if err != nil {
		data = []byte("# kickback settings\nport: 9021\nlog:\n  dir: log\ntasks: []")
		err := ioutil.WriteFile(confFile, data, os.ModePerm)
		if err != nil {
			panic("cannot open task file: " + confFile)
		}
		fmt.Println("created .kickback.yml, modify this")
	}
	err = yaml.Unmarshal(data, &config)
	if err != nil {
		panic("invalid task yml")
	}
}

func createLogger(logPath string) *logrus.Logger {
	var err error
	logger := logrus.New()
	logFile, err := os.OpenFile(logPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		panic(err)
	}
	logger.Formatter = &logrus.TextFormatter{FullTimestamp: true, DisableColors: true}
	logger.Out = logFile

	return logger
}

func main() {
	port := flag.Int("port", 9201, "start port")
	confFile := flag.String("conf", ".kickback.yml", "config file path")
	version := flag.Bool("version", false, "show version")
	flag.Parse()

	if *version {
		fmt.Println(Version)
		os.Exit(0)
	}

	pwd, err := os.Getwd()
	if err != nil {
		panic(err)
	}

	readConfig(fmt.Sprintf("%s/%s", pwd, *confFile))
	err = os.Mkdir(fmt.Sprintf("%s/%s", pwd, config.Log["dir"]), 0755)
	if err != nil {
		if strings.Index(err.Error(), "file exists") > -1 {
		} else {
			panic(err)
		}
	}

	accessLogger = createLogger(fmt.Sprintf("%s/%s/access.log", pwd, config.Log["dir"]))
	systemLogger = createLogger(fmt.Sprintf("%s/%s/system.log", pwd, config.Log["dir"]))

	gin.SetMode(gin.ReleaseMode)
	r := gin.Default()
	r.Use(KBLog(accessLogger, time.RFC3339, false))
	r.GET("/api/tasks", getTasks)

	for _, task := range config.Tasks {
		addTask(r, task)
	}

	r.Run(fmt.Sprintf(":%d", *port))
}
