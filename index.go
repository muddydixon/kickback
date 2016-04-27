package main

import (
	"bytes"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path"
	"reflect"
	"strings"
	"text/template"

	"github.com/Sirupsen/logrus"
	// "github.com/go-errors/errors"
	"github.com/gin-gonic/gin"
	"gopkg.in/yaml.v2"
)


type Tasks struct {
	tasks []Task
}

type Task struct {
	Name   string
	Path   string
	Method string
	Procs  []string
}

var taskFile *string
var logger *logrus.Logger

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
				logrus.WithFields(logrus.Fields{
					"method": "method",
				}).Info("some error")
			}
			outs = append(outs, out.String())
		}
		c.JSON(200, gin.H{"stdout": strings.Join(outs, "\n")})
	}
}

func readTaskFile() []Task {
	p := path.Join(".", *taskFile)
	data, err := ioutil.ReadFile(p)
	if err != nil {
		err := ioutil.WriteFile(p, []byte("# kickback tasks"), os.ModePerm)
		if err != nil {
			panic("cannot open task file: " + p)
		}
	}
	tasks := []Task{}
	err = yaml.Unmarshal(data, &tasks)
	if err != nil {
		panic("invalid task yml")
	}
	return tasks
}

func main() {
	logger = logrus.New()

	taskFile = flag.String("taskfile", ".kickback.yml", "taskfile path")
	flag.Parse()
	fmt.Println("taskFile: " + fmt.Sprint(*taskFile))

	tasks := readTaskFile()

	r := gin.Default()
	r.GET("/api/tasks", getTasks)
	r.POST("/api/tasks", createTask)
	r.GET("/api/tasks/:task", getTask)
	r.PUT("/api/tasks/:task", modifyTask)
	r.DELETE("/api/tasks/task", deleteTask)

	for _, task := range tasks {
		addTask(r, task)
	}

	r.Run(fmt.Sprintf(":%d", 8080))
}
