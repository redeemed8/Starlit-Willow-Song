package common

import (
	"context"
	"errors"
	"github.com/gin-gonic/gin"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func Run(r *gin.Engine, addr string, srvName string, stop func()) {

	srv := &http.Server{
		Addr:    addr,
		Handler: r,
	}
	//	保证下面的优雅启停
	go func() {
		log.Printf("%s running in %s \n", srvName, srv.Addr)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("%s failed to start , cause by : %v ... \n", srvName, err)
		}
	}()
	//	标记通道
	quit := make(chan os.Signal)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Printf("Shutting down project : %s ... \n", srvName)

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	//	关闭其余的 grpc服务
	if stop != nil {
		stop()
	}

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("project %s Shutdown , cause by : %v ... \n", srvName, err)
	}

	select {
	case <-ctx.Done():
		log.Printf("protect %s wait timeout ... \n", srvName)
	}

	log.Printf("project %s stop success ... \n", srvName)
}
