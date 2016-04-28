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

	"github.com/Sirupsen/logrus"
	"github.com/gin-gonic/contrib/ginrus"
	"github.com/gin-gonic/gin"
	"gopkg.in/yaml.v2"
)

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

var accessLogger *logrus.Logger
var systemLogger *logrus.Logger

func addTask(r *gin.Engine, task Task) {
	method := strings.ToUpper(task.Method)
	api := reflect.ValueOf(r).MethodByName(method)
	api.Call([]reflect.Value{reflect.ValueOf(task.Path), reflect.ValueOf(execProcsGen(task.Procs))})
}

func getTasks(c *gin.Context) {

}

func createTask(c *gin.Context) {

}

func getTask(c *gin.Context) {

}

func modifyTask(c *gin.Context) {

}

func deleteTask(c *gin.Context) {

}

func execProcsGen(procs []string) gin.HandlerFunc {
	t := template.New("procs")
	return func(c *gin.Context) {
		var err error
		var outs []string

		x, _ := ioutil.ReadAll(c.Request.Body)
		bodies := strings.Split(string(x), "&")
		data := make(map[string]string)
		for _, body := range bodies {
			kv := strings.Split(body, "=")
			data[kv[0]] = kv[1]
		}

		for _, proc := range procs {
			tmpl, _ := t.Parse(proc)
			var b bytes.Buffer
			var out bytes.Buffer

			tmpl.Execute(&b, data)
			cmd := exec.Command("sh", "-c", b.String())
			cmd.Stdout = &out
			err = cmd.Run()
			if err != nil {
				systemLogger.WithFields(logrus.Fields{
					"api": "execProcs",
				}).Info("some error")
			}
			outs = append(outs, out.String())
		}
		c.JSON(200, gin.H{"stdout": strings.Join(outs, "\n")})
	}
}

func readConfig(confFile string) Config {
	data, err := ioutil.ReadFile(confFile)
	if err != nil {
		err := ioutil.WriteFile(confFile, []byte("# kickback tasks"), os.ModePerm)
		if err != nil {
			panic("cannot open task file: " + confFile)
		}
	}
	config := Config{}
	err = yaml.Unmarshal(data, &config)
	if err != nil {
		panic("invalid task yml")
	}
	return config
}

func createLogger(logPath string) *logrus.Logger {
	var err error
	logger := logrus.New()
	logFile, err := os.OpenFile(logPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		panic(err)
	}
	logger.Formatter = &logrus.TextFormatter{FullTimestamp: true, DisableColors: false}
	logger.Out = logFile

	return logger
}

func main() {
	port := flag.Int("port", 9201, "start port")
	confFile := flag.String("conf", ".kickback.yml", "config file path")
	flag.Parse()

	pwd, err := os.Getwd()
	if err != nil {
		panic(err)
	}

	config := readConfig(fmt.Sprintf("%s/%s", pwd, *confFile))
	err = os.Mkdir(fmt.Sprintf("%s/%s", pwd, config.Log["dir"]), 0755)
	if err != nil {
		if strings.Index(err.Error(), "file exists") > -1 {
		} else {
			panic(err)
		}
	}

	accessLogger = createLogger(fmt.Sprintf("%s/%s/access.log", pwd, config.Log["dir"]))
	systemLogger = createLogger(fmt.Sprintf("%s/%s/system.log", pwd, config.Log["dir"]))

	r := gin.Default()
	r.Use(ginrus.Ginrus(accessLogger, time.RFC3339, true))
	r.GET("/api/tasks", getTasks)
	r.POST("/api/tasks", createTask)
	r.GET("/api/tasks/:task", getTask)
	r.PUT("/api/tasks/:task", modifyTask)
	r.DELETE("/api/tasks/task", deleteTask)

	for _, task := range config.Tasks {
		addTask(r, task)
	}

	r.Run(fmt.Sprintf(":%d", *port))
}
