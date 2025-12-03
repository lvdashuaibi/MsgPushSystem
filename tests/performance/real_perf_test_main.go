package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sync"
	"sync/atomic"
	"time"
)

const (
	BaseURL = "http://localhost:8109"
)

type TestResult struct {
	TotalRequests   int64
	SuccessRequests int64
	FailedRequests  int64
	TotalDuration   time.Duration
	MinLatency      time.Duration
	MaxLatency      time.Duration
	AvgLatency      time.Duration
	P95Latency      time.Duration
	P99Latency      time.Duration
	TPS             float64
}

type SendMsgRequest struct {
	To       string `json:"to"`
	Subject  string `json:"subject"`
	Content  string `json:"content"`
	Priority int    `json:"priority"`
}

func main() {
	fmt.Println("🚀 开始真实性能测试...")
	fmt.Println("=" + string(make([]byte, 60)))

	// 测试1: API响应时间测试
	fmt.Println("\n📊 测试1: API响应时间测试 (100并发, 持续30秒)")
	apiResult := testAPIResponse(100, 30*time.Second)
	printResult("API响应时间测试", apiResult)

	time.Sleep(5 * time.Second)

	// 测试2: 不同并发下的吞吐量测试
	fmt.Println("\n📊 测试2: 吞吐量测试")
	concurrencies := []int{50, 100, 200}
	for _, c := range concurrencies {
		fmt.Printf("\n  测试并发数: %d\n", c)
		result := testThroughput(c, 20*time.Second)
		printResult(fmt.Sprintf("并发%d", c), result)
		time.Sleep(3 * time.Second)
	}

	fmt.Println("\n✅ 所有测试完成！")
}

func testAPIResponse(concurrency int, duration time.Duration) *TestResult {
	var (
		totalRequests   int64
		successRequests int64
		failedRequests  int64
		latencies       []time.Duration
		latenciesMutex  sync.Mutex
	)

	startTime := time.Now()
	endTime := startTime.Add(duration)

	var wg sync.WaitGroup

	// 启动并发goroutine
	for i := 0; i < concurrency; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()

			for time.Now().Before(endTime) {
				reqStart := time.Now()

				// 发送请求
				req := SendMsgRequest{
					To:       fmt.Sprintf("test%d@example.com", id),
					Subject:  fmt.Sprintf("性能测试消息 #%d", atomic.AddInt64(&totalRequests, 1)),
					Content:  "这是一条性能测试消息",
					Priority: 2,
				}

				success := sendMessage(req)
				latency := time.Since(reqStart)

				if success {
					atomic.AddInt64(&successRequests, 1)
				} else {
					atomic.AddInt64(&failedRequests, 1)
				}

				latenciesMutex.Lock()
				latencies = append(latencies, latency)
				latenciesMutex.Unlock()

				// 控制请求速率
				time.Sleep(10 * time.Millisecond)
			}
		}(i)
	}

	wg.Wait()
	totalDuration := time.Since(startTime)

	return calculateResult(totalRequests, successRequests, failedRequests, totalDuration, latencies)
}

func testThroughput(concurrency int, duration time.Duration) *TestResult {
	return testAPIResponse(concurrency, duration)
}

func sendMessage(req SendMsgRequest) bool {
	data, _ := json.Marshal(req)

	resp, err := http.Post(
		BaseURL+"/msg/send_msg",
		"application/json",
		bytes.NewBuffer(data),
	)

	if err != nil {
		return false
	}
	defer resp.Body.Close()

	// 读取响应
	_, _ = io.ReadAll(resp.Body)

	return resp.StatusCode == 200
}

func calculateResult(total, success, failed int64, duration time.Duration, latencies []time.Duration) *TestResult {
	if len(latencies) == 0 {
		return &TestResult{}
	}

	// 排序延迟
	sortLatencies(latencies)

	// 计算统计数据
	var totalLatency time.Duration
	for _, l := range latencies {
		totalLatency += l
	}

	result := &TestResult{
		TotalRequests:   total,
		SuccessRequests: success,
		FailedRequests:  failed,
		TotalDuration:   duration,
		MinLatency:      latencies[0],
		MaxLatency:      latencies[len(latencies)-1],
		AvgLatency:      totalLatency / time.Duration(len(latencies)),
		P95Latency:      latencies[int(float64(len(latencies))*0.95)],
		P99Latency:      latencies[int(float64(len(latencies))*0.99)],
		TPS:             float64(success) / duration.Seconds(),
	}

	return result
}

func sortLatencies(latencies []time.Duration) {
	// 简单的冒泡排序
	n := len(latencies)
	for i := 0; i < n-1; i++ {
		for j := 0; j < n-i-1; j++ {
			if latencies[j] > latencies[j+1] {
				latencies[j], latencies[j+1] = latencies[j+1], latencies[j]
			}
		}
	}
}

func printResult(testName string, result *TestResult) {
	fmt.Printf("\n【%s 结果】\n", testName)
	fmt.Printf("  总请求数:     %d\n", result.TotalRequests)
	fmt.Printf("  成功请求:     %d\n", result.SuccessRequests)
	fmt.Printf("  失败请求:     %d\n", result.FailedRequests)
	fmt.Printf("  成功率:       %.2f%%\n", float64(result.SuccessRequests)/float64(result.TotalRequests)*100)
	fmt.Printf("  测试时长:     %.2f秒\n", result.TotalDuration.Seconds())
	fmt.Printf("  吞吐量(TPS):  %.2f\n", result.TPS)
	fmt.Printf("  平均延迟:     %v\n", result.AvgLatency)
	fmt.Printf("  最小延迟:     %v\n", result.MinLatency)
	fmt.Printf("  最大延迟:     %v\n", result.MaxLatency)
	fmt.Printf("  P95延迟:      %v\n", result.P95Latency)
	fmt.Printf("  P99延迟:      %v\n", result.P99Latency)

	// 保存结果到文件
	saveResultToFile(testName, result)
}

func saveResultToFile(testName string, result *TestResult) {
	_ = testName
	_ = result
	// 简化处理，结果已通过stdout输出
}
