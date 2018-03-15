package main

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"strings"

	"github.com/robertkrimen/otto"
)

var (
	// ErrScriptNotFound 脚本未找到
	ErrScriptNotFound = errors.New("script not found")
	// ErrReadScript 读取脚本文件错误
	ErrReadScript = errors.New("read script error")
)

var (

	// 是否调试模式
	debug = false

	// 默认端口
	port = "7788"

	// 显示帮助
	help = false

	// 脚本文件
	scriptFile string
)

const (
	// 广告
	banner string = `
   __ ____     ______        __    ____                    
  / // / /    /_  __/__ ___ / /_  / __/__ _____  _____ ____
 / _  / /__    / / / -_|_-</ __/ _\ \/ -_) __/ |/ / -_) __/
/_//_/____/   /_/  \__/___/\__/ /___/\__/_/  |___/\__/_/   	
	
	`
)

// 入口
func main() {
	parseCommand()
	handleSignal()
	startServer()
}

// 当前目录
func getCurrentDirectory() string {
	dir, err := filepath.Abs(filepath.Dir(os.Args[0]))
	if err != nil {
		log.Fatal(err)
	}
	return strings.Replace(dir, "\\", "/", -1)
}

// 处理退出信号
func handleSignal() {
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	go func() {
		for _ = range c {
			fmt.Println("Bye Bye")
			os.Exit(0)
		}
	}()
}

// 解析参数
func parseCommand() {
	flag.BoolVar(&help, "h", false, "this help")
	flag.StringVar(&port, "p", port, "port")
	flag.StringVar(&scriptFile, "s", getCurrentDirectory()+"/script.js", "script.js path")
	flag.Parse()
	if help {
		flag.Usage()
		os.Exit(0)
	}
}

// 加载脚本文件
func loadScript() (string, error) {
	_, err := os.Stat(scriptFile)
	if err != nil {
		return "", ErrScriptNotFound
	}
	data, err := ioutil.ReadFile(scriptFile)
	if err != nil {
		return "", ErrReadScript
	}
	return string(data), err
}

// 启动http服务
func startServer() {
	fmt.Println(banner)
	fmt.Printf("Server Listen at %s\n", port)
	fmt.Printf("Script file is %s\n", scriptFile)
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		result, err := handleJS(r, w)
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}

		bytes, err := json.Marshal(result)
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}
		w.Header().Set("content-type", "application/json")
		w.Write(bytes)
	})
	log.Fatal(http.ListenAndServe(":"+port, nil))
}

// 调用js
func handleJS(r *http.Request, w http.ResponseWriter) (result interface{}, err error) {
	result = ""

	// 读取脚本
	data, err := loadScript()
	if err != nil {
		return
	}

	jsvm := otto.New()

	// 设置常用方法
	jsvm.Set("getHost", func(call otto.FunctionCall) otto.Value {
		result, _ := jsvm.ToValue(r.Host)
		return result
	})
	jsvm.Set("getMethod", func(call otto.FunctionCall) otto.Value {
		result, _ := jsvm.ToValue(r.Method)
		return result
	})
	jsvm.Set("getUri", func(call otto.FunctionCall) otto.Value {
		result, _ := jsvm.ToValue(r.URL.RequestURI())
		return result
	})
	jsvm.Set("getQuery", func(call otto.FunctionCall) otto.Value {
		result, _ := jsvm.ToValue(r.URL.Query())
		return result
	})
	jsvm.Set("getForm", func(call otto.FunctionCall) otto.Value {
		err = r.ParseForm()
		result, _ := jsvm.ToValue(r.Form)
		return result
	})
	jsvm.Set("getBody", func(call otto.FunctionCall) otto.Value {
		body, _ := ioutil.ReadAll(r.Body)
		result, _ := jsvm.ToValue(string(body))
		return result
	})

	if _, err = jsvm.Run(data); err != nil {
		return
	}

	value, err := jsvm.Get("result")
	if err != nil {
		return
	}

	result, err = value.Export()

	return
}
