package main

import (
	"database/sql"
	"fmt"
	_ "github.com/denisenkom/go-mssqldb"
	"github.com/kardianos/service"
	"github.com/robfig/cron/v3"
	"log"
	"os"
	"time"
	"workwxboot/pkg/workwx"
)
/****
交叉编译为exe
CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build -ldflags "-s -w"
upx workwxboot.exe

注册为windows服务
workwxboot.exe install
workwxboot.exe uninstall
workwxboot.exe start
workwxboot.exe stop
*/


// 数据库连接参数变量
var (
	server ="127.0.0.1"
	port = "1437"
	user = "sa"
	password ="abc123."
	database ="Galasys"
	db *sql.DB
)
// 计划任务
var crontab *cron.Cron

func init() {
	// 初始化连接数据库
	connectToTheDatabase()
}

// region 注册服务相关
var logger service.Logger
type program struct{}
func (p *program) Start(s service.Service) error {
	// Start should not block. Do the actual work async.
	// 启动不应阻塞。异步执行实际工作。
	go p.run()
	return nil
}
func (p *program) run() {
	// Do work here
	// 做计划任务
	scheduledTasks()
}
func (p *program) Stop(s service.Service) error {
	// Stop should not block. Return with a few seconds.
	// 停止不应阻塞。几秒钟后返回。
	// 关闭计划任务
	crontab.Stop()
	// 关闭数据库链接
	db.Close()
	return nil
}
// endregion

func main() {
	// 拦截panic 直接不输出
	defer func(){
		if x := recover(); x != nil {
			//处理panic, 让程序从panicking状态恢复的机会
			os.Exit(0)
		}
	}()

	// region 注册服务相关
	svcConfig := &service.Config{
		Name:        "qingqingRanchDaily",
		DisplayName: "qingqingRanchDaily",
		Description: "qingqingRanchDaily",
	}
	prg := &program{}
	s, err := service.New(prg, svcConfig)
	if err != nil {
		log.Fatal(err)
	}
	logger, err = s.Logger(nil)
	if err != nil {
		log.Fatal(err)
	}
	if len(os.Args) > 1 {
		switch os.Args[1] {
		case "install":
			err :=s.Install()
			if err!=nil {
				log.Fatalf("Install service error:%s\n", err.Error())
			}
			return
		case "uninstall":
			err:=s.Uninstall()
			if err!=nil {
				log.Fatalf("Uninstall service error:%s\n", err.Error())
			}
			return
		case "start":
			err:=s.Start()
			if err !=nil{
				log.Fatalf("start service error:%s\n", err.Error())
			}
			return
		case "stop":
			err:=s.Stop()
			if err != nil {
				log.Fatalf("stop service error:%s\n", err.Error())
			}
			return
		case "status":
			fmt.Println(s.Status())
			return
		case "restart":
			err:=s.Restart()
			if err != nil {
				log.Fatalf("Restart service error:%s\n", err.Error())
			}
			return
		}
	}
	err = s.Run()
	if err != nil {
		logger.Error(err)
	}
	// endregion

}

// scheduledTasks 计划任务
func scheduledTasks()  {
	//新建一个定时任务兑现
	//根据cron表达式进行时间调度，cron可以精确到秒，大部分表达式格式也是从秒开始
	//crontab = cron.New() // 默认从分开始进行时间调度
	crontab := cron.New(cron.WithSeconds())// 精确到秒
	// 定义定时器调用的任务函数
	task := func() {
		// 获取数据并发送
		//getDataFromTheDatabase()
		fmt.Println("hello world", time.Now())
	}
	//定时任务 每天早上8:30执行
	spec:="0 30 8 1/1 * ? *" //cron表达式
	//spec:="*/5 * * * * ?" //cron表达式 每5秒执行一次
	// 添加定是任务
	crontab.AddFunc(spec,task)
	// 启动定时器
	crontab.Start()
	// 定时任务是另起协程执行的,这里使用 select 简答阻塞.实际开发中需要
	// 根据实际情况进行控制
	select {} //阻塞主线程停止
}


func connectToTheDatabase() {
	// 数据库链接
	connString := fmt.Sprintf("server=%s;port=%s;user id=%s;password=%s;database=%s;encrypt=disable",
		server,port, user, password, database)
	conn,err:= sql.Open("sqlserver",connString)
	if err !=nil {
		log.Println("数据库连接失败")
	}
	db=conn
}

func getDataFromTheDatabase()  {

	// 先解析
	stmt, err := db.Prepare(`SELECT  项目名称,项目编码 FROM Twb_info;`)
	if err != nil {
		log.Printf("\nPrepare failed:%T %+v\n", err, err)
	}
	//QueryRow TMD是读取单条记录
	rows, err := stmt.Query()
	if err != nil {
		fmt.Println("Error reading rows: " + err.Error())
	}
	defer stmt.Close()

	count:=0
	for rows.Next(){
		var 项目名称 string
		var 项目编码 string
		err := rows.Scan(&项目名称, &项目编码)
		if err != nil {
			log.Fatal("Scan failed:", err.Error())
		}
		fmt.Println(项目名称,项目编码)
		count++
	}







	// 获取数据

	// 发送数据
	//qingqingRanchIncomeDaily(2,3,5)
}

/***
qingqingRanchIncomeDaily 青青牧场收入日报
闸机人数 numberOfGates
门票售卖 ticketSales
商品售卖 merchandiseSale
 */
func qingqingRanchIncomeDaily(numberOfGates,ticketSales,merchandiseSale int64 )  {
	// 获取上一天日期
	yesTime := time.Now().AddDate(0, 0, -1).Format("2006年01月02日")
	//fmt.Println(yesTime)

	webhook := "https://qyapi.weixin.qq.com/cgi-bin/webhook/send?key=cbeac157-32d9-4ec0-ab82-2e99d376ed96"
	wxbot := workwx.NewRobot(webhook)
	//合计金额
	total := ticketSales + merchandiseSale
	dailyReport := fmt.Sprintf("# %v 收入详情：\n", yesTime)
	dailyReport += fmt.Sprintf("## 进闸机人数：<font color=\"info\">%v</font> \n", numberOfGates)
	dailyReport += fmt.Sprintf("## 门票售卖情况：<font color=\"info\">%v</font> \n", ticketSales)
	dailyReport += fmt.Sprintf("## 商品售卖情况：<font color=\"info\">%v</font> \n", merchandiseSale)
	dailyReport += fmt.Sprintf("# 合计：**<font color=\"warning\">%v</font>** \n", total)

	markdown := workwx.WxBotMessage{
		MsgType:  "markdown",
		MarkDown: workwx.BotMarkDown{Content: dailyReport}}
	wxbot.Send(markdown)
}
