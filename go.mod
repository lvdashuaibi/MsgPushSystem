module github.com/lvdashuaibi/MsgPushSystem

go 1.16

require (
	github.com/BurntSushi/toml v1.2.1
	github.com/IBM/sarama v1.42.1
	github.com/alibabacloud-go/darabonba-openapi/v2 v2.0.10
	github.com/alibabacloud-go/dysmsapi-20170525/v4 v4.1.1
	github.com/alibabacloud-go/tea v1.2.2
	github.com/alibabacloud-go/tea-utils/v2 v2.0.7
	github.com/gin-gonic/gin v1.9.1
	github.com/go-redsync/redsync/v4 v4.8.1
	github.com/go-sql-driver/mysql v1.7.0
	github.com/google/uuid v1.5.0
	github.com/redis/go-redis/v9 v9.7.0
	github.com/sirupsen/logrus v1.9.3
	gopkg.in/alexcesaro/quotedprintable.v3 v3.0.0-20150716171945-2caba252f4dc // indirect
	gopkg.in/gomail.v2 v2.0.0-20160411212932-81ebce5c23df
	gorm.io/driver/mysql v1.5.2
	gorm.io/gorm v1.25.5
	gopkg.in/natefinch/lumberjack.v2 v2.2.1
)

replace github.com/go-critic/go-critic v0.0.0-20181204210945-ee9bf5809ead => github.com/go-critic/go-critic v0.4.4-0.20200527115708-e76e5c043d31

replace github.com/golangci/errcheck v0.0.0-20181003203344-ef45e06d44b6 => github.com/golangci/errcheck v0.0.0-20181223084120-ef45e06d44b6

replace github.com/golangci/go-tools v0.0.0-20180109140146-af6baa5dc196 => github.com/golangci/go-tools v0.0.0-20190124090046-35a9f45a5db0

replace github.com/golangci/gofmt v0.0.0-20181105071733-0b8337e80d98 => github.com/golangci/gofmt v0.0.0-20190930125516-244bba706f1a

replace github.com/golangci/gosec v0.0.0-20180901114220-66fb7fc33547 => github.com/golangci/gosec v0.0.0-20180901114220-8afd9cbb6cfb

replace golang.org/x/tools v0.0.0-20190314010720-f0bfdbff1f9c => golang.org/x/tools v0.0.0-20200604042327-9b20fe4cabe8

replace mvdan.cc/unparam v0.0.0-20190124213536-fbb59629db34 => mvdan.cc/unparam v0.0.0-20200501210554-b37ab49443f7
