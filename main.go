/*
本程序由Bing Ai生成
*/
package main

import (
	"bytes"
	"context"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"strings"
	"time"

	"gopkg.in/yaml.v2"
)

type Config struct {
	Sites    []string `yaml:"sites"`
	Days     int      `yaml:"days"`
	Timeout  int      `yaml:"timeout"`
	External string   `yaml:"external"`
	Method   string   `yaml:"method"`
	Args     string   `yaml:"args"`
}

func getConfig() Config {
	data, err := os.ReadFile("config.yaml")
	if err != nil {
		panic(err)
	}
	config := Config{}
	err = yaml.Unmarshal(data, &config)
	if err != nil {
		panic(err)
	}
	return config
}

func main() {
	config := getConfig()
	for _, site := range config.Sites {
		checkCertificate(site, config.Days, config.Timeout, config.External, config.Method, config.Args)
	}
}

func checkCertificate(site string, days int, timeout int, external string, method string, argsTemplate string) {
	client := http.Client{
		Timeout: time.Duration(timeout) * time.Second,
	}
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(timeout)*time.Second)
	defer cancel()
	req, err := http.NewRequestWithContext(ctx, "GET", site, nil)
	if err != nil {
		message := fmt.Sprintf("无法创建请求 %s: %v\n", site, err)
		fmt.Print(message)
		runExternalProgram(external, message, method, argsTemplate)
		return
	}
	resp, err := client.Do(req)
	if err != nil {
		message := fmt.Sprintf("无法连接到 %s: %v", site, err)
		fmt.Println(message)
		runExternalProgram(external, message, method, argsTemplate)
		return
	}
	defer resp.Body.Close()

	if resp.TLS == nil {
		fmt.Printf("%s 没有使用 SSL/TLS\n", site)
		return
	}

	// 获取时区
	location, err := time.LoadLocation("Asia/Shanghai")
	if err != nil {
		panic(err)
	}

	expTime := ""
	for _, cert := range resp.TLS.PeerCertificates {
		if cert.IsCA {
			continue
		}

		expTime = cert.NotAfter.In(location).Format("2006-01-02")
		if time.Now().After(cert.NotAfter) {
			message := fmt.Sprintf("%s 的 SSL 证书已经过期. 过期时间: %s", site, expTime)
			fmt.Println(message)
			runExternalProgram(external, message, method, argsTemplate)
			return
		}

		if time.Now().AddDate(0, 0, days).After(cert.NotAfter) {
			message := fmt.Sprintf("%s 的 SSL 证书将在 %d 天内过期. 过期时间: %s", site, days, expTime)
			fmt.Println(message)
			runExternalProgram(external, message, method, argsTemplate)
			return
		}
	}

	fmt.Printf("%s 的 SSL 证书未过期,过期时间: %s\n", site, expTime)
}

func runExternalProgram(external string, message string, method string, argsTemplate string) {
	var cmd *exec.Cmd
	if method == "args" {
		args := strings.Split(argsTemplate, " ")
		for i, arg := range args {
			args[i] = strings.Replace(arg, "{message}", message, -1)
		}
		cmd = exec.Command(external, args...)
		fmt.Printf("运行命令: %s %v\n", external, args)
	} else {
		cmd = exec.Command(external)
		cmd.Stdin = bytes.NewBufferString(message)
	}
	var out bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &stderr
	err := cmd.Run()
	if err != nil {
		fmt.Printf("运行外部程序 %s 出错: %v\n", external, err)
		fmt.Printf("错误信息: %s\n", stderr.String())
	} else {
		fmt.Printf("运行外部程序 %s 成功\n", external)
		fmt.Printf("运行信息: %s\n", out.String())
	}
}
