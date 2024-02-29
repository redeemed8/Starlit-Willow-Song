package service

import (
	"errors"
	"gorm.io/gorm"
	"jcpd.cn/user/internal/constants"
	"jcpd.cn/user/internal/models"
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
	Timer    *time.Timer
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
	myTimer.Timer = time.NewTimer(durationUtilPeriod)
	myTimer.Hour = hour
}

type TaskFunc func()

// fillDealFunc 装填处理函数
func (myTimer *myTimer) fillDealFunc(taskfunc TaskFunc) {
	myTimer.TaskFunc = taskfunc
}

const cleanApplyHour = 12
const cleanApplySign = "clean_apply"

// cleanUsedApply 设置一个定时任务在协程中开启，定时清理一些已经通过或拒绝了的申请
func cleanUsedApply() {
	var myTimer_ myTimer
	myTimer_.makeTimerByHour(cleanApplyHour)
	taskFunc := TaskFunc(func() {
		// 清理一些已经通过或拒绝了的申请
		err := models.JoinApplyDao.DeleteApplyByNotStatus(models.JoinApplyUtil.GetPendStatus())
		if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
			constants.MysqlErr("删除过期apply过程出错", err)
		}
	})
	myTimer_.fillDealFunc(taskFunc)
	//	加入到定时任务列表
	TimerTasks.putTimer(cleanApplySign, myTimer_)
	log.Println("定时任务:清理已审核的申请信息  --  状态：已开启")
}

const cleanGroupHour = 3
const cleanGroupSign = "clean_group"

// cleanDeletedGroup 删除已被解散的群聊，并且从用户的群列表中删除
func cleanDeletedGroup() {
	var myTimer_ myTimer
	myTimer_.makeTimerByHour(cleanGroupHour)
	taskFunc := TaskFunc(func() {
		// 删除已被解散的群聊，并且从用户的群列表中删除

	})
	myTimer_.fillDealFunc(taskFunc)
	//	加入到定时任务列表
	TimerTasks.putTimer(cleanGroupSign, myTimer_)
	log.Println("定时任务:清理已被解散的群聊  --  状态：已开启")
}

// Start 开启定时任务
func (tasks *timerTasks_) Start() {
	tasks.init()
	go func() {
		for {
			select {
			case <-tasks.myTimers[cleanApplySign].Timer.C:
				tasks.myTimers[cleanApplySign].TaskFunc()
			case <-tasks.myTimers[cleanGroupSign].Timer.C:
				tasks.myTimers[cleanGroupSign].TaskFunc()
			}
		}
	}()
}
