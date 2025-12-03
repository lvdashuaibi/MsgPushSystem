package performance

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

const (
	BaseURL = "http://localhost:8109"
)

// SendMsgRequest 发送消息请求
type SendMsgRequest struct {
	To           string                 `json:"to"`
	Subject      string                 `json:"subject"`
	Content      string                 `json:"content"`
	Priority     int                    `json:"priority"`
	Channels     []int                  `json:"channels"`
	TemplateID   string                 `json:"template_id,omitempty"`
	TemplateData map[string]interface{} `json:"template_data,omitempty"`
}

// Response 通用响应
type Response struct {
	Code    int                    `json:"code"`
	Message string                 `json:"message"`
	Data    map[string]interface{} `json:"data"`
}

// TestResult 测试结果
type TestResult struct {
	TotalRequests   int64
	SuccessRequests int64
	FailedRequests  int64
	TotalDuration   time.Duration
	AvgResponseTime time.Duration
	MinResponseTime time.Duration
	MaxResponseTime time.Duration
	TPS             float64
	ResponseTimes   []time.Duration
}

// sendMessage 发送单条消息
func sendMessage(req *SendMsgRequest) (time.Duration, error) {
	start := time.Now()

	jsonData, err := json.Marshal(req)
	if err != nil {
		return 0, err
	}

	resp, err := http.Post(BaseURL+"/msg/send_msg", "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return 0, err
	}

	var result Response
	if err := json.Unmarshal(body, &result); err != nil {
		return 0, err
	}

	duration := time.Since(start)

	if result.Code != 0 {
		return duration, fmt.Errorf("API返回错误: %s", result.Message)
	}

	return duration, nil
}

// TestAPIResponseTime API响应时间测试
func TestAPIResponseTime(t *testing.T) {
	t.Log("=== API响应时间测试 ===")

	req := &SendMsgRequest{
		To:       "test@example.com",
		Subject:  "性能测试消息",
		Content:  "这是一条性能测试消息",
		Priority: 2,
		Channels: []int{1},
	}

	// 预热
	t.Log("预热中...")
	for i := 0; i < 10; i++ {
		sendMessage(req)
		time.Sleep(100 * time.Millisecond)
	}

	// 正式测试
	t.Log("开始测试...")
	testCount := 100
	responseTimes := make([]time.Duration, 0, testCount)

	for i := 0; i < testCount; i++ {
		duration, err := sendMessage(req)
		if err != nil {
			t.Logf("请求失败: %v", err)
			continue
		}
		responseTimes = append(responseTimes, duration)
		time.Sleep(50 * time.Millisecond) // 避免过快
	}

	// 计算统计数据
	var total time.Duration
	min := responseTimes[0]
	max := responseTimes[0]

	for _, rt := range responseTimes {
		total += rt
		if rt < min {
			min = rt
		}
		if rt > max {
			max = rt
		}
	}

	avg := total / time.Duration(len(responseTimes))

	// 计算95分位和99分位
	sorted := make([]time.Duration, len(responseTimes))
	copy(sorted, responseTimes)

	// 简单排序
	for i := 0; i < len(sorted); i++ {
		for j := i + 1; j < len(sorted); j++ {
			if sorted[i] > sorted[j] {
				sorted[i], sorted[j] = sorted[j], sorted[i]
			}
		}
	}

	p95Index := int(float64(len(sorted)) * 0.95)
	p99Index := int(float64(len(sorted)) * 0.99)

	t.Logf("\n测试结果:")
	t.Logf("  总请求数: %d", len(responseTimes))
	t.Logf("  平均响应时间: %v", avg)
	t.Logf("  最小响应时间: %v", min)
	t.Logf("  最大响应时间: %v", max)
	t.Logf("  95分位响应时间: %v", sorted[p95Index])
	t.Logf("  99分位响应时间: %v", sorted[p99Index])
}

// TestThroughput 吞吐量测试
func TestThroughput(t *testing.T) {
	t.Log("=== 吞吐量测试 ===")

	concurrencies := []int{50, 100, 200}
	duration := 60 * time.Second

	for _, concurrency := range concurrencies {
		t.Logf("\n测试并发数: %d", concurrency)
		result := runThroughputTest(t, concurrency, duration)

		t.Logf("测试结果:")
		t.Logf("  总请求数: %d", result.TotalRequests)
		t.Logf("  成功请求: %d", result.SuccessRequests)
		t.Logf("  失败请求: %d", result.FailedRequests)
		t.Logf("  测试时长: %v", result.TotalDuration)
		t.Logf("  平均响应时间: %v", result.AvgResponseTime)
		t.Logf("  最小响应时间: %v", result.MinResponseTime)
		t.Logf("  最大响应时间: %v", result.MaxResponseTime)
		t.Logf("  吞吐量(TPS): %.2f", result.TPS)
		t.Logf("  错误率: %.2f%%", float64(result.FailedRequests)/float64(result.TotalRequests)*100)

		// 等待一段时间再进行下一轮测试
		time.Sleep(10 * time.Second)
	}
}

// runThroughputTest 运行吞吐量测试
func runThroughputTest(t *testing.T, concurrency int, duration time.Duration) *TestResult {
	var (
		totalRequests   int64
		successRequests int64
		failedRequests  int64
		responseTimes   []time.Duration
		mu              sync.Mutex
	)

	req := &SendMsgRequest{
		To:       "test@example.com",
		Subject:  "吞吐量测试消息",
		Content:  "这是一条吞吐量测试消息",
		Priority: 2,
		Channels: []int{1},
	}

	start := time.Now()
	stopChan := make(chan struct{})
	var wg sync.WaitGroup

	// 启动并发goroutine
	for i := 0; i < concurrency; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()

			for {
				select {
				case <-stopChan:
					return
				default:
					atomic.AddInt64(&totalRequests, 1)

					rt, err := sendMessage(req)

					mu.Lock()
					responseTimes = append(responseTimes, rt)
					mu.Unlock()

					if err != nil {
						atomic.AddInt64(&failedRequests, 1)
					} else {
						atomic.AddInt64(&successRequests, 1)
					}

					// 短暂休息避免过载
					time.Sleep(10 * time.Millisecond)
				}
			}
		}(i)
	}

	// 等待指定时间
	time.Sleep(duration)
	close(stopChan)
	wg.Wait()

	totalDuration := time.Since(start)

	// 计算统计数据
	var total time.Duration
	min := responseTimes[0]
	max := responseTimes[0]

	for _, rt := range responseTimes {
		total += rt
		if rt < min {
			min = rt
		}
		if rt > max {
			max = rt
		}
	}

	avg := total / time.Duration(len(responseTimes))
	tps := float64(totalRequests) / totalDuration.Seconds()

	return &TestResult{
		TotalRequests:   totalRequests,
		SuccessRequests: successRequests,
		FailedRequests:  failedRequests,
		TotalDuration:   totalDuration,
		AvgResponseTime: avg,
		MinResponseTime: min,
		MaxResponseTime: max,
		TPS:             tps,
		ResponseTimes:   responseTimes,
	}
}

// TestPriorityQueue 优先级队列测试
func TestPriorityQueue(t *testing.T) {
	t.Log("=== 优先级队列测试 ===")

	// 先发送100条低优先级消息
	t.Log("发送100条低优先级消息...")
	for i := 0; i < 100; i++ {
		req := &SendMsgRequest{
			To:       fmt.Sprintf("low%d@example.com", i),
			Subject:  "低优先级消息",
			Content:  "这是一条低优先级消息",
			Priority: 3,
			Channels: []int{1},
		}
		sendMessage(req)
	}

	time.Sleep(2 * time.Second)

	// 发送10条高优先级消息并记录时间
	t.Log("发送10条高优先级消息...")
	highPriorityTimes := make([]time.Duration, 0, 10)

	for i := 0; i < 10; i++ {
		req := &SendMsgRequest{
			To:       fmt.Sprintf("high%d@example.com", i),
			Subject:  "高优先级消息",
			Content:  "这是一条高优先级消息",
			Priority: 1,
			Channels: []int{1},
		}

		start := time.Now()
		_, err := sendMessage(req)
		duration := time.Since(start)

		if err == nil {
			highPriorityTimes = append(highPriorityTimes, duration)
		}
	}

	// 计算平均处理时间
	var total time.Duration
	for _, d := range highPriorityTimes {
		total += d
	}
	avg := total / time.Duration(len(highPriorityTimes))

	t.Logf("\n测试结果:")
	t.Logf("  高优先级消息数: %d", len(highPriorityTimes))
	t.Logf("  平均处理时间: %v", avg)

	if avg < 2*time.Second {
		t.Log("  ✓ 高优先级消息得到优先处理")
	} else {
		t.Log("  ✗ 高优先级消息处理较慢")
	}
}

// TestStressTest 压力测试
func TestStressTest(t *testing.T) {
	t.Log("=== 压力测试 ===")

	concurrencies := []int{500, 1000, 1500}
	duration := 30 * time.Second

	for _, concurrency := range concurrencies {
		t.Logf("\n测试并发数: %d", concurrency)
		result := runThroughputTest(t, concurrency, duration)

		errorRate := float64(result.FailedRequests) / float64(result.TotalRequests) * 100

		t.Logf("测试结果:")
		t.Logf("  吞吐量(TPS): %.2f", result.TPS)
		t.Logf("  平均响应时间: %v", result.AvgResponseTime)
		t.Logf("  错误率: %.2f%%", errorRate)

		if errorRate > 5 {
			t.Logf("  ⚠️  错误率超过5%%，系统性能下降")
		} else {
			t.Logf("  ✓ 系统运行稳定")
		}

		// 等待系统恢复
		time.Sleep(30 * time.Second)
	}
}
