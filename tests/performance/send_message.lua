-- wrk Lua脚本：发送消息性能测试

-- 请求计数器
request_count = 0

-- 初始化
function setup(thread)
   thread:set("id", request_count)
   request_count = request_count + 1
end

-- 生成请求
function request()
   local headers = {}
   headers["Content-Type"] = "application/json"

   local body = string.format([[{
      "to": "test%d@example.com",
      "subject": "性能测试消息",
      "content": "这是一条性能测试消息，用于测试系统性能",
      "priority": 2,
      "channels": [1]
   }]], math.random(1, 10000))

   return wrk.format("POST", "/msg/send_msg", headers, body)
end

-- 响应处理
function response(status, headers, body)
   if status ~= 200 then
      print("Error: HTTP " .. status)
      print("Body: " .. body)
   end
end

-- 完成统计
function done(summary, latency, requests)
   io.write("------------------------------\n")
   io.write("测试完成统计:\n")
   io.write(string.format("  总请求数: %d\n", summary.requests))
   io.write(string.format("  总时长: %.2fs\n", summary.duration / 1000000))
   io.write(string.format("  平均TPS: %.2f\n", summary.requests / (summary.duration / 1000000)))
   io.write(string.format("  平均延迟: %.2fms\n", latency.mean / 1000))
   io.write(string.format("  最小延迟: %.2fms\n", latency.min / 1000))
   io.write(string.format("  最大延迟: %.2fms\n", latency.max / 1000))
   io.write(string.format("  50分位延迟: %.2fms\n", latency:percentile(50) / 1000))
   io.write(string.format("  90分位延迟: %.2fms\n", latency:percentile(90) / 1000))
   io.write(string.format("  95分位延迟: %.2fms\n", latency:percentile(95) / 1000))
   io.write(string.format("  99分位延迟: %.2fms\n", latency:percentile(99) / 1000))
   io.write(string.format("  错误数: %d\n", summary.errors.connect + summary.errors.read + summary.errors.write + summary.errors.status + summary.errors.timeout))
   io.write("------------------------------\n")
end
