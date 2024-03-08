package router

import (
	"errors"
	"gorm.io/gorm"
	"jcpd.cn/user/internal/constants"
	"jcpd.cn/user/internal/models"
	"jcpd.cn/user/utils"
	"log"
	"time"
)

// TimerTasks 对外开放使用的 timerTasks_单例
var TimerTasks timerTasks_

// timerTasks_ 定时任务列表
type timerTasks_ struct {
	myTimers map[string]myTimer
}

// new 新建定时任务列表
func (tasks *timerTasks_) new() {
	tasks.myTimers = make(map[string]myTimer)
}

// init 初始化定时任务列表
func (tasks *timerTasks_) init() {
	tasks.new()
	cleanUsedApply()
	cleanDeletedGroup()
}

// putTimer 向定时任务列表里添加任务
func (tasks *timerTasks_) putTimer(timerSign string, timer myTimer) {
	tasks.myTimers[timerSign] = timer
}

// myTimer 定时器 - 定时任务
type myTimer struct {
	Timer    *time.Ticker
	TaskFunc TaskFunc
	Hour     int
}

// makeTimerByHour 根据小时创建定时器
func (myTimer *myTimer) makeTimerByHour(hour int) {
	curTimePeriod := time.Now()
	nextTimePeriod := time.Date(curTimePeriod.Year(), curTimePeriod.Month(), curTimePeriod.Day(), hour, 0, 0, 0, curTimePeriod.Location())
	if curTimePeriod.After(nextTimePeriod) {
		nextTimePeriod = nextTimePeriod.Add(24 * time.Hour)
	}
	durationUtilPeriod := nextTimePeriod.Sub(curTimePeriod)
	myTimer.Timer = time.NewTicker(durationUtilPeriod)
	myTimer.Hour = hour
}

// makeTimerInterval  创建一个定时器 - 间隔时间执行
func (myTimer *myTimer) makeTimerInterval(interval time.Duration) {
	myTimer.Timer = time.NewTicker(interval)
}

type TaskFunc func(*myTimer)

// fillDealFunc 装填处理函数
func (myTimer *myTimer) fillDealFunc(taskfunc TaskFunc) {
	myTimer.TaskFunc = taskfunc
}

//	----------------------------------

const cleanApplyHour = 12
const cleanApplySign = "clean_apply"

// cleanUsedApply 设置一个定时任务在协程中开启，定时清理一些已经通过或拒绝了的申请
func cleanUsedApply() {
	var myTimer_ myTimer
	myTimer_.makeTimerByHour(cleanApplyHour)
	taskFunc := TaskFunc(func(t *myTimer) {
		// 清理一些已经通过或拒绝了的申请
		err := models.JoinApplyDao.DeleteApplyByNotStatus(models.JoinApplyUtil.GetPendStatus())
		if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
			constants.MysqlErr("删除过期apply过程出错", err)
		}
		//	重置定时器
		t.makeTimerByHour(cleanApplyHour)
	})
	myTimer_.fillDealFunc(taskFunc)
	//	加入到定时任务列表
	TimerTasks.putTimer(cleanApplySign, myTimer_)
	log.Println(constants.Hint("定时任务:清理已审核的申请信息  --  状态：已开启"))
}

//	----------------------------------

const cleanGroupHour = 3
const cleanGroupSign = "clean_group"

// cleanDeletedGroup 删除已被解散的群聊，并且从用户的群列表中删除
func cleanDeletedGroup() {
	var myTimer_ myTimer
	myTimer_.makeTimerByHour(cleanGroupHour)
	taskFunc := TaskFunc(func(t *myTimer) {
		// 删除已被解散的群聊，并且从用户的群列表中删除

		//	获取所有已解散的群
		groups, err1 := models.GroupInfoDao.GetGroupsByMap(map[string]interface{}{"status": models.GroupDeleted})
		if err1 != nil {
			log.Println(constants.Hint("获取已删除的群聊失败,err = " + err1.Error()))
			return
		}

		//	在用户群列表中删除
		for _, group := range groups {
			err2 := models.UserInfoDao.GroupListUpdates(group.Id, utils.ParseListToUint(group.MemberIds))
			if err2 != nil {
				log.Println(constants.Hint("在用户群列表中删除群id出错,err = " + err2.Error()))
				return
			}
		}

		//	删除群
		err3 := models.GroupInfoDao.DeleteGroupById(groups.Ids())
		if err3 != nil {
			log.Println(constants.Hint("删除群信息失败,err = " + err3.Error()))
			return
		}

		//	重置定时器
		t.makeTimerByHour(cleanApplyHour)
	})
	myTimer_.fillDealFunc(taskFunc)
	//	加入到定时任务列表
	TimerTasks.putTimer(cleanGroupSign, myTimer_)
	log.Println(constants.Hint("定时任务:清理已被解散的群聊  --  状态：已开启"))
}

//	----------------------------------

const Working = "working"
const Resting = "resting"

var TimeTaskSign string

var TaskCloseChan = make(chan struct{}, 1)

// Check  检查是否有定时任务正在执行中
func (tasks *timerTasks_) Check() {
	//	检查定时任务是否正在工作
	if TimeTaskSign == Working {
		log.Println(constants.Hint("等待定时任务结束...."))
		for {
			time.Sleep(25 * time.Millisecond)
			if TimeTaskSign == Resting {
				break
			}
		}
		log.Println(constants.Info("定时任务已经结束...."))
	}
	//	发送停止信号
	TaskCloseChan <- struct{}{}
	return
}

// Start 开启定时任务
func (tasks *timerTasks_) Start() {
	tasks.init()

	time.Sleep(time.Minute)

	for {
		select {
		case <-tasks.myTimers[cleanApplySign].Timer.C:
			{
				TimeTaskSign = Working
				timer := tasks.myTimers[cleanApplySign]
				tasks.myTimers[cleanApplySign].TaskFunc(&timer)
			}
		case <-tasks.myTimers[cleanGroupSign].Timer.C:
			{
				TimeTaskSign = Working
				timer := tasks.myTimers[cleanGroupSign]
				tasks.myTimers[cleanGroupSign].TaskFunc(&timer)
			}
		case <-TaskCloseChan:
			{
				log.Println(constants.Info("定时任务已关闭..."))
				return
			}
		}
		TimeTaskSign = Resting
	}

}
