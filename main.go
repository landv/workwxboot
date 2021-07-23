package main

import (
	"database/sql"
	"fmt"
	_ "github.com/denisenkom/go-mssqldb"
	"github.com/kardianos/service"
	"github.com/robfig/cron/v3"
	"log"
	"os"
	"strconv"
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
	server   = "127.0.0.1"
	port     = "1437"
	user     = "sa"
	password = "abc123."
	database = "Galasys"
	db       *sql.DB
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

	// demo
	//aa := getGateData()
	//fmt.Println(aa)
	getDataFromTheDatabase()

	// 拦截panic 直接不输出
	defer func() {
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
			err := s.Install()
			if err != nil {
				log.Fatalf("Install service error:%s\n", err.Error())
			}
			return
		case "uninstall":
			err := s.Uninstall()
			if err != nil {
				log.Fatalf("Uninstall service error:%s\n", err.Error())
			}
			return
		case "start":
			err := s.Start()
			if err != nil {
				log.Fatalf("start service error:%s\n", err.Error())
			}
			return
		case "stop":
			err := s.Stop()
			if err != nil {
				log.Fatalf("stop service error:%s\n", err.Error())
			}
			return
		case "status":
			fmt.Println(s.Status())
			return
		case "restart":
			err := s.Restart()
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
func scheduledTasks() {
	//新建一个定时任务兑现
	//根据cron表达式进行时间调度，cron可以精确到秒，大部分表达式格式也是从秒开始
	//crontab = cron.New() // 默认从分开始进行时间调度
	crontab := cron.New(cron.WithSeconds()) // 精确到秒
	// 定义定时器调用的任务函数
	task := func() {
		// 获取数据并发送
		getDataFromTheDatabase()
		//fmt.Println("hello world", time.Now())
	}
	//定时任务 每天早上8:30执行
	spec := "0 30 8 1/1 * ? *" //cron表达式
	//spec:="*/5 * * * * ?" //cron表达式 每5秒执行一次
	// 添加定是任务
	crontab.AddFunc(spec, task)
	// 启动定时器
	crontab.Start()
	// 定时任务是另起协程执行的,这里使用 select 简答阻塞.实际开发中需要
	// 根据实际情况进行控制
	select {} //阻塞主线程停止
}

func connectToTheDatabase() {
	// 数据库链接
	connString := fmt.Sprintf("server=%s;port=%s;user id=%s;password=%s;database=%s;encrypt=disable",
		server, port, user, password, database)
	conn, err := sql.Open("sqlserver", connString)
	if err != nil {
		log.Println("数据库连接失败")
	}
	db = conn
}

func getDataFromTheDatabase() {

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

	count := 0
	for rows.Next() {
		var 项目名称 string
		var 项目编码 string
		err := rows.Scan(&项目名称, &项目编码)
		if err != nil {
			log.Fatal("Scan failed:", err.Error())
		}
		fmt.Println(项目名称, 项目编码)
		count++
	}

	// 获取数据

	// 发送数据
	qingqingRanchIncomeDaily()
}


/****
-- 报表清单 需要调用存储过程
select * from GS_RM_REPORTMANAGEMENT WHERE  rm_name in ('闸机入园报表')
select * from GS_RM_REPORTMANAGEMENT where RM_GROUP in ('线下票务报表')

exec PRC_RM_ALLSELLTOTAL
@v_start_time = '2021-07-22 00:00:00',
@v_end_Time = '2021-07-22 23:59:59',
@v_Ohterwhere = ''

 */

// 获取闸机数据
/***
青青牧场客户尊享券免费,数量:<font color="info">39</font>张,合计:<font color="info">0.00</font>元
青青牧场入园礼包38,数量:<font color="info">1</font>张,合计:<font color="info">38.00</font>元
总人数合计：<font color="info">40</font>人
*/
func getGateData() (sa string) {
	startingTime := time.Now().AddDate(0, 0, -1).Format("2006-01-02 00:00:00")
	endTime := time.Now().Format("2006-01-02 00:00:00")
	//fmt.Printf(`%v`,endTime)
	// 先解析
	stmt, err := db.Prepare(fmt.Sprintf(`select
	bbb.*,
	(bbb.价格 * bbb.数量)as 总价
from
(select abcd.项目名称,sum(abcd.价格)as 价格 ,COUNT(abcd.价格) as 数量  from (SELECT 
        CASE WHEN LEN(GFA.SREMARK) = 18 THEN GFA.SREMARK 
		WHEN LEN(GFTOD.BindCard) = 18 THEN GFTOD.BindCard 
		WHEN TWB.TWB_MSG = 'IDCARD' AND LEN(TWB.TWB_ID) = 18 THEN TWB.TWB_ID 
		ELSE NULL END AS 身份证号 ,

		CASE WHEN TWB.TWB_MSG='ADMIN' THEN DT.DROPUP_TYPENAME 
		WHEN TWB.TWB_MSG='OLINETICKET' AND (SELECT COUNT(*) FROM TWB_TRN_ONLINE WHERE ECODE=TWB.TWB_ID)>0 THEN  TTO.TICKETNAME
		WHEN  TWB.TWB_MSG='OLINETICKET' AND (SELECT COUNT(*) FROM TWB_TRN_ONLINE WHERE ECODE=TWB.TWB_ID)=0  THEN '未获取到名称'
		WHEN  TWB.TWB_MSG='AGINETICKET'  THEN '二次入园门票'
		WHEN  TWB.TWB_MSG='IDCARD'  THEN '身份证直接入园'
		ELSE TBI.STICKETNAMECH 
		END AS 项目名称,
		CONVERT(varchar(100),TWB.TWB_SYSTIME,23) 入园日期,
		CONVERT(varchar(100),TWB.TWB_SYSTIME,24) 入园时间,
	    Convert(int,Datename(hour,TWB.TWB_SYSTIME))  AS  小时段,
		 TWB.INPARKCOUNT 入园次数, TWB.PEOPLECOUNT 人数,
	   datename(w,TWB.TWB_SYSTIME) 星期,
	   DATENAME(YEAR,TWB.TWB_SYSTIME) AS 年,
	   DATENAME(MONTH,TWB.TWB_SYSTIME) AS 月,
	   DATENAME(DAY,TWB.TWB_SYSTIME) AS 日,
		
       CASE WHEN GFA.NINCOMETYPE=1 OR GFA.NINCOMETYPE=5 OR GFA.NINCOMETYPE=6 OR GFA.NINCOMETYPE=7 THEN TWB.TWB_ID  
		WHEN  GFA.NINCOMETYPE=2 THEN MFCRD_MANUALNO
		WHEN GFA.NINCOMETYPE= 3 OR GFA.NINCOMETYPE=4 THEN GFA.SREMARK
		WHEN TWB.TWB_MSG='ADMIN' THEN TWB.TWB_ID
		WHEN TWB.TWB_MSG='OLINETICKET' THEN TWB.TWB_ID
		WHEN TWB.TWB_MSG='AGINETICKET' THEN TWB.TWB_ID
		WHEN TWB.TWB_MSG='EMPCARD'  THEN TWB.TWB_ID
		WHEN TWB.TWB_MSG='IDCARD' THEN TWB.TWB_ID 
		END AS 项目编码,

		CASE WHEN GFA.NINCOMETYPE=1 THEN '门票' 
		WHEN GFA.NINCOMETYPE=2 THEN '年卡' 
		WHEN GFA.NINCOMETYPE=3 THEN
		 (SELECT DICT_VALUE FROM DICT_TABLE
		WHERE DICT_TYPE_ID='MAKETICKETMODE' AND DICT_KEY=2)
		WHEN GFA.NINCOMETYPE=4 THEN 
		 (SELECT DICT_VALUE FROM DICT_TABLE
		WHERE DICT_TYPE_ID='MAKETICKETMODE' AND DICT_KEY=3)
		WHEN GFA.NINCOMETYPE=5 THEN GFD.SPMODE
		 WHEN GFA.NINCOMETYPE=6  THEN GFD.SPMODE
		WHEN TWB.TWB_MSG='ADMIN' THEN '管理卡'
		WHEN TWB.TWB_MSG='OLINETICKET' THEN '线上门票'   
		WHEN TWB.TWB_MSG='AGINETICKET' THEN '二次入园' 
		WHEN TWB.TWB_MSG='EMPCARD'  THEN '员工卡' 
		WHEN GFA.NINCOMETYPE=7 THEN 'TVM售票'
		WHEN TWB.TWB_MSG='IDCARD' THEN '身份证入园' 
		END AS 项目类型,
		GFA.NDEALID 交易号,
		 CASE WHEN GFA.NINCOMETYPE=2 THEN CARD.MFCRD_NAME 
		WHEN DT.DROPUP_STATUS='ADMIN' THEN  EMP.EMP_NAME 
		WHEN DT.DROPUP_STATUS='Up' THEN  EMP.EMP_NAME 
		WHEN DT.DROPUP_STATUS='Drop' THEN  EMP.EMP_NAME 
		WHEN DT.DROPUP_STATUS='ACCREDIT' THEN  EMP.EMP_NAME 
		end 持卡人,
		 TBI.NGENERALPRICE AS 价格,
		 DEVICE.SDEVICENAME 检票终端,
        GZONE.SGZONENAME 检票点, A.SPARKNAME 景区, COM.SCOMPANYNAME 公司,
		ISNULL(GSFO.Travel,'线下窗口') 渠道, GSFO.OrderNo as 订单号
	 
	FROM  TWB_TRN TWB  WITH(NOLOCK) 
	LEFT JOIN TWB_TRN_ONLINE TTO  WITH(NOLOCK) ON TWB.TWB_ID =TTO.ECODE
	LEFT JOIN GS_F_ACCESS GFA  WITH(NOLOCK) ON TWB.TWB_ID =GFA.SBARCODE
	LEFT JOIN GS_F_DEALINFO GFD  WITH(NOLOCK) ON GFD.NDEALID =GFA.NDEALID
	LEFT JOIN MFYEARCRD_TBL CARD  WITH(NOLOCK)  ON CARD.SBARCODE =GFA.SBARCODE
	LEFT JOIN GS_T_TICKETBASEINFO TBI  WITH(NOLOCK) ON TBI.NTICKETID=GFA.NTICKETID
	LEFT JOIN GS_C_DEVICE DEVICE  WITH(NOLOCK) ON DEVICE.NDEVICEID = TWB.TWB_GATE
	LEFT JOIN GS_C_GZONE GZONE  WITH(NOLOCK) ON GZONE.NGZONEID =DEVICE.NGZONEID
	LEFT JOIN GS_C_PARK A  WITH(NOLOCK) ON A.NPARKID = GZONE.NPARKID 
	LEFT JOIN GS_C_COMPANY COM  WITH(NOLOCK) ON COM.NCOMPANYID = A.NCOMPANYID
	LEFT JOIN DROPUP_TBL DT  WITH(NOLOCK) ON DT.DROPUP_ID=TWB.TWB_ID
	LEFT JOIN SAC_employee EMP  WITH(NOLOCK) ON EMP.EMP_ID=DT.DROPUP_USER
	LEFT JOIN GS_F_ThirdOnline GSFO  WITH(NOLOCK) on GSFO.NDEALID = GFD.NDEALID
	LEFT JOIN GS_F_ThirdOnlineDetail GFTOD  WITH(NOLOCK) ON GFTOD.SBARCODE = GFA.SBARCODE

	WHERE twb_msg <>'OUT' and TWB_MSG <> 'EMPCARD'  and 
	TWB.TWB_SYSTIME BETWEEN '%v' AND '%v'  
	--order by  TWB.TWB_SYSTIME
	)abcd
	GROUP by abcd.项目名称
	)bbb`, startingTime, endTime))
	if err != nil {
		log.Printf("\nPrepare failed:%T %+v\n", err, err)
	}
	//QueryRow TMD是读取单条记录
	rows, err := stmt.Query()
	if err != nil {
		fmt.Println("Error reading rows: " + err.Error())
	}
	defer stmt.Close()

	count := 0
	var (
		项目名称 string
		价格   string
		数量   string
		总价   string
		总人数  int
	)

	for rows.Next() {

		err := rows.Scan(&项目名称, &价格, &数量, &总价)
		if err != nil {
			log.Fatal("Scan failed:", err.Error())
		}
		//fmt.Println(g)
		i, _ := strconv.Atoi(数量)
		//fmt.Println("i",i)
		总人数 += i
		sa += fmt.Sprintf(">%v—数量:<font color=\"info\">%v</font>张,合计:<font color=\"info\">%v</font>元 \n", 项目名称, 数量, 总价)
		count++
	}
	sa += fmt.Sprintf(">总人数合计：<font color=\"info\">%d</font>人 \n", 总人数)
	return
}

// 获取商品售卖数据
func getProductSalesData()  {
	
}

/***
qingqingRanchIncomeDaily 青青牧场收入日报
闸机人数 numberOfGates
门票售卖 ticketSales
商品售卖 merchandiseSale
*/
func qingqingRanchIncomeDaily() {
	// 获取上一天日期
	yesTime := time.Now().AddDate(0, 0, -1).Format("2006年01月02日")
	//fmt.Println(yesTime)

	webhook := "https://qyapi.weixin.qq.com/cgi-bin/webhook/send?key=cbeac157-32d9-4ec0-ab82-2e99d376ed96"
	wxbot := workwx.NewRobot(webhook)
	total:=1

	dailyReport := fmt.Sprintf("# %v 收入详情：\n", yesTime)
	//dailyReport += fmt.Sprintf("## 进闸机人数：<font color=\"info\">%v</font> \n", numberOfGates)
	dailyReport += fmt.Sprintf("## 门票售卖情况：\n")
	dailyReport += getGateData()
	dailyReport += "\n"
	dailyReport += fmt.Sprintf("## 商品售卖情况：\n")
	dailyReport += fmt.Sprintf("# 合计：**<font color=\"warning\">%v</font>** \n", total)

	fmt.Println(dailyReport)

	markdown := workwx.WxBotMessage{
		MsgType:  "markdown",
		MarkDown: workwx.BotMarkDown{Content: dailyReport}}
	wxbot.Send(markdown)
}
