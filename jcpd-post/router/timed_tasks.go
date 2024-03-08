package router

import (
	"jcpd.cn/post/internal/constants"
	"jcpd.cn/post/internal/models"
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
	flushBloomFilter()
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

// makeTimerByHour 根据小时创建定时器 - 每天的指定时间执行
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

type TaskFunc func(t *myTimer)

// fillDealFunc 装填处理函数
func (myTimer *myTimer) fillDealFunc(taskfunc TaskFunc) {
	myTimer.TaskFunc = taskfunc
}

//	----------------------------------

const updateHotPostTime = 1 * time.Minute
const updateHotPostSign = "update_hot_post"

// updateHotPost 定时任务，将 redis中的点赞数，同步到redis，同时更新热点帖子id
func updateHotPost() {
	var myTimer_ myTimer
	myTimer_.makeTimerInterval(updateHotPostTime)
	taskFunc := TaskFunc(func(t *myTimer) {
		//  定时任务，将 redis中的点赞数，同步到redis，同时更新热点帖子 id

	})
	myTimer_.fillDealFunc(taskFunc)
	//	加入到定时任务列表
	TimerTasks.putTimer(updateHotPostSign, myTimer_)
	log.Println(constants.Hint("定时任务:更新热点帖子  --  状态：已开启"))
}

//	----------------------------------

const flushBloomFilterHour = 4
const flushBloomFilterSign = "flush_bloom_filter"

func flushBloomFilter() {
	var myTimer_ myTimer
	myTimer_.makeTimerByHour(flushBloomFilterHour)
	taskFunc := TaskFunc(func(t *myTimer) {
		//	为了减少一些可能的不必要的问题
		time.Sleep(time.Second)

		//	刷新过滤器组
		models.BloomFilters.Flush()

		//	重置定时器
		if t != nil {
			t.makeTimerByHour(flushBloomFilterHour)
		}
	})
	myTimer_.fillDealFunc(taskFunc)
	//	加入到定时任务列表
	TimerTasks.putTimer(flushBloomFilterSign, myTimer_)
	log.Println(constants.Hint("定时任务:刷新布隆过滤器  --  状态：已开启"))
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
		case <-tasks.myTimers[updateHotPostSign].Timer.C:
			{
				TimeTaskSign = Working
				tasks.myTimers[updateHotPostSign].TaskFunc(nil)
			}
		case <-tasks.myTimers[flushBloomFilterSign].Timer.C:
			{
				TimeTaskSign = Working
				timer := tasks.myTimers[flushBloomFilterSign]
				tasks.myTimers[flushBloomFilterSign].TaskFunc(&timer)
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
