package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/BitofferHub/msgcenter/src/config"
	"github.com/BitofferHub/msgcenter/src/ctrl/consumer"
	"github.com/BitofferHub/msgcenter/src/data"
	"github.com/BitofferHub/msgcenter/src/initialize"
	"github.com/BitofferHub/pkg/middlewares/log"
	"github.com/gin-gonic/gin"
)

func main() {
	// 初始化配置
	config.Init()
	_, err := data.NewData(config.Conf)
	if err != nil {
		log.Errorf("initialize NewData err %s", err.Error())
		return
	}
	cs := consumer.NewMsgConsume()
	cs.Consume()
	var tmc consumer.TimerMsgConsume
	tmc.Consume()
	consumer.InitMsgProc()

	// 启动定时消息调度器
	smc := consumer.NewScheduledMessageConsumer()
	smc.Start()

	// 设置信号处理，确保在程序退出前释放分布式锁
	setupSignalHandler(cs, &tmc)

	// 创建一个web服务
	router := gin.Default()

	// 设置中间件处理中文编码
	router.Use(func(c *gin.Context) {
		c.Header("Content-Type", "application/json; charset=utf-8")
		c.Next()
	})

	// 这里跳进去就能看到有哪些接口
	initialize.RegisterRouter(router)
	fmt.Println("before router run")
	// 启动web server，这一步之后这个主协程启动会阻塞在这里，请求可以通过gin的子协程进来
	err = router.Run(fmt.Sprintf(":%d", config.Conf.Common.Port))
	fmt.Println(err)
}

// setupSignalHandler 设置信号处理，确保在程序退出前释放锁
func setupSignalHandler(cs *consumer.MsgConsume, tmc *consumer.TimerMsgConsume) {
	c := make(chan os.Signal, 1)
	// 监听 SIGINT, SIGTERM, SIGQUIT 信号
	signal.Notify(c, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)

	go func() {
		sig := <-c
		log.Infof("接收到系统信号: %v，准备优雅退出", sig)

		// 释放所有分布式锁
		log.Info("释放所有分布式锁...")
		cs.UnlockAll()

		log.Info("锁释放完成，程序退出")
		os.Exit(0)
	}()
}
