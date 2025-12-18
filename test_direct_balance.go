package main

import (
	"fmt"
	"net/http"
)

// 直接测试Gate.io API，不使用日志库，确保能看到完整输出
func main() {
	// 测试API密钥
	apiKey := "643f71c728188c157207b5c9f79d1b1a"
	secretKey := "43a97ba7bd31ddfac27bc43bfc3c01a2812d8972f8d3de8abe2be4431407ff47"

	// 直接调用Gate.io API获取余额
	resp, err := http.Get("https://api.gateio.ws/api/v4/futures/usdt/accounts")
	if err != nil {
		fmt.Printf("HTTP请求失败: %v\n", err)
		return
	}
	defer resp.Body.Close()

	fmt.Printf("Status: %s\n", resp.Status)
	fmt.Printf("Headers: %v\n", resp.Header)
}
