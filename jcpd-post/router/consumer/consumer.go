package consumer

import (
	"context"
	"errors"
	"github.com/IBM/sarama"
	"jcpd.cn/post/internal/constants"
	"jcpd.cn/post/internal/options"
	"log"
	"os"
	"os/signal"
	"sync"
	"time"
)

var KafkaListener kafkaListener

type kafkaListener struct {
	Status string
}

const KafkaWork = "1"
const KafkaRest = "0"
const KafkaError = "2"

func (listener *kafkaListener) StartListen() {
	defer listener.close(options.C.Consumer)

	//	开启失败重试
	go Retry()

	//	等待协程关闭
	wg := &sync.WaitGroup{}
	wg.Add(1)

	//	开启
	go func() {
		err := listener.start(wg, Consumer{})
		//	监听结果
		if errors.Is(err, sarama.ErrBrokerNotAvailable) {
			//	标记kafka异常，让定时任务暂时关闭，并进行重试
			KafkaListenerFail <- struct{}{}
			listener.err()
			wg.Done()
		}
	}()

	//	等待关闭信号
	sigterm := make(chan os.Signal, 1)
	signal.Notify(sigterm, os.Interrupt)
	<-sigterm

	listener.close(options.C.Consumer)
	wg.Wait()

}

// start 启动消息监听
func (*kafkaListener) start(wg *sync.WaitGroup, consumer Consumer) error {
	defer func() {
		if wg != nil {
			wg.Done()
		}
	}()
	log.Println(constants.Info("kafka listener already started ..."))
	for {
		err := options.C.Consumer.Consume(context.Background(), options.C.KafKa.Consumer.Topics, &consumer)
		if err != nil {
			log.Println(constants.Hint("kafka listener closed , cause by : " + err.Error()))
			return err
		}
	}
}

// close 关闭连接
func (listener *kafkaListener) close(Consumer sarama.ConsumerGroup) {
	for listener.Status == KafkaWork {
		time.Sleep(100 * time.Microsecond) //	0.1 毫秒
	}

	if err := Consumer.Close(); err != nil {
		log.Println(constants.Hint("Application two failed to close consumer , cause by : " + err.Error()))
	}
}

func (listener *kafkaListener) rest() {
	listener.Status = KafkaRest
}

func (listener *kafkaListener) work() {
	listener.Status = KafkaWork
}

func (listener *kafkaListener) err() {
	listener.Status = KafkaError
}

func (listener *kafkaListener) IsErr() bool {
	return listener.Status == KafkaError
}

func (listener *kafkaListener) Check() {
	for listener.Status == KafkaWork {
		time.Sleep(500 * time.Millisecond)
	}
	defer options.C.Producer.AsyncClose()
}

// 	------------------------------------- 0.0

var FailRetryQueue = make(chan *FailMsg, 1000)

type FailMsg struct {
	Msg  *sarama.ProducerMessage
	Time int
}

func NewFailMsg(msg *sarama.ProducerMessage) *FailMsg {
	return &FailMsg{Msg: msg, Time: 0}
}

var KafkaListenerFail = make(chan struct{}, 1)

// Retry 失败重试
func Retry() {
	exit := make(chan os.Signal, 1)
	signal.Notify(exit, os.Interrupt)
	for {
		select {
		case failMsg := <-FailRetryQueue:
			{
				//	取出对其进行重试三次
				if err := Send(failMsg.Msg); err != nil && failMsg.Time < 3 {
					failMsg.Time++
					FailRetryQueue <- failMsg
				}
			}
		case <-KafkaListenerFail:
			{
				//	尝试重新连接kafka服务
				go func() {
					err := KafkaListener.start(nil, Consumer{})
					if errors.Is(err, sarama.ErrBrokerNotAvailable) {
						//	说明仍然连接不上, 等待一会继续重试 ...
						log.Println(constants.Err("Application two failed to reconnect kafka.. , cause by : " + err.Error()))
						time.Sleep(10 * time.Second)
						KafkaListenerFail <- struct{}{}
					} else {
						log.Println(constants.Info("Application two reconnect kafka successfully ..."))
					}
				}()
			}
		case <-exit:
			return
		}
	}
}

func Send(msg *sarama.ProducerMessage) error {
	//	发送消息
	successChan := options.C.Producer.Successes()
	errorChan := options.C.Producer.Errors()

	// 发送消息到 Kafka
	options.C.Producer.Input() <- msg

	// 设置超时时间
	timeout := time.After(time.Second)

	// 等待消息的确认
	select {
	case <-successChan:
		return nil
	case err := <-errorChan:
		log.Println(constants.Err("Failed to Send message:" + err.Error()))
		FailRetryQueue <- NewFailMsg(msg) // 	放入失败重试队列
		return err
	case <-timeout:
		log.Println(constants.Err("Timeout: failed to receive message confirmation"))
	}
	FailRetryQueue <- NewFailMsg(msg) // 	放入失败重试队列
	return TimeoutErr
}

var TimeoutErr = errors.New("time out")
