package consumer

import (
	"context"
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

func (listener *kafkaListener) StartListen() {
	defer listener.close(options.C.Consumer)

	//	等待协程关闭
	wg := &sync.WaitGroup{}
	wg.Add(1)

	//	开启
	go func() {
		listener.start(wg, Consumer{})
	}()

	//	等待关闭信号
	sigterm := make(chan os.Signal, 1)
	signal.Notify(sigterm, os.Interrupt)
	<-sigterm

	listener.close(options.C.Consumer)
	wg.Wait()

}

// start 启动消息监听
func (*kafkaListener) start(wg *sync.WaitGroup, consumer Consumer) {
	defer wg.Done()
	log.Println(constants.Info("kafka listener already started ..."))
	for {
		err := options.C.Consumer.Consume(context.Background(), options.C.KafKa.Consumer.Topics, &consumer)
		if err != nil {
			log.Println(constants.Hint("kafka listener closed , cause by : " + err.Error()))
			return
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

func (listener *kafkaListener) Check() {
	for listener.Status == KafkaWork {
		time.Sleep(500 * time.Millisecond)
	}
}
