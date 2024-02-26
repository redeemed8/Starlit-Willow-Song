package service

import (
	"errors"
	"gorm.io/gorm"
	"jcpd.cn/user/internal/constants"
	"jcpd.cn/user/internal/models"
	"log"
	"time"
)

type myTimer struct {
	Timer    *time.Timer
	TaskFunc TaskFunc
	Hour     int
}

// MakeTimerByHour 根据小时创建定时器
func (myTimer *myTimer) MakeTimerByHour(hour int) {
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

// FillDealFunc 装填处理函数
func (myTimer *myTimer) FillDealFunc(taskfunc TaskFunc) {
	myTimer.TaskFunc = taskfunc
}

func (myTimer *myTimer) Start() {
	for {
		<-myTimer.Timer.C
		myTimer.TaskFunc()
		time.Sleep(10 * time.Second)
		myTimer.MakeTimerByHour(myTimer.Hour) //	重置定时器
	}
}

// CleanUsedApply 设置一个定时任务在协程中开启，定时清理一些已经通过或拒绝了的申请
func (h *ApplyHandler) CleanUsedApply(hour int) {
	var myTimer myTimer
	myTimer.MakeTimerByHour(hour)
	taskFunc := TaskFunc(func() {
		err := models.JoinApplyDao.DeleteApplyByNotStatus(models.JoinApplyUtil.GetPendStatus())
		if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
			constants.MysqlErr("删除过期apply失败", err)
		}
	})
	myTimer.FillDealFunc(taskFunc)
	myTimer.Start()
	log.Println("定时任务:清理已审核的申请信息  --  状态：已开启")
}
