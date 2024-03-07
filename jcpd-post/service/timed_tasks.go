package service

import (
	"jcpd.cn/post/internal/constants"
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
	//	在这里 初始化定时任务
	updateHotPost()
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

// makeTimerInterval  创建

type TaskFunc func(t *myTimer)

// fillDealFunc 装填处理函数
func (myTimer *myTimer) fillDealFunc(taskfunc TaskFunc) {
	myTimer.TaskFunc = taskfunc
}

//	----------------------------------

const updateHotPostHour = 1
const updateHotPostSign = "update_hot_post"

// updateHotPost 定时任务，将 redis中的点赞数，同步到redis，同时更新热点帖子id
func updateHotPost() {
	var myTimer_ myTimer
	myTimer_.makeTimerByHour(updateHotPostHour)
	taskFunc := TaskFunc(func(t *myTimer) {
		//  定时任务，将 redis中的点赞数，同步到redis，同时更新热点帖子 id

		//	重置定时器
		t.makeTimerByHour(updateHotPostHour)
	})
	myTimer_.fillDealFunc(taskFunc)
	//	加入到定时任务列表
	TimerTasks.putTimer(updateHotPostSign, myTimer_)
	log.Println(constants.Hint("定时任务:更新热点帖子  --  状态：已开启"))
}

//	----------------------------------

// Start 开启定时任务
func (tasks *timerTasks_) Start() {
	tasks.init()
	go func() {
		for {
			select {
			case <-tasks.myTimers[updateHotPostSign].Timer.C:
				timer := tasks.myTimers[updateHotPostSign]
				tasks.myTimers[updateHotPostSign].TaskFunc(&timer)
			}
		}
	}()
}
