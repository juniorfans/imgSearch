package dbOptions

import (
	"os"
	"os/signal"
	"syscall"
	"fmt"
	"sync/atomic"
)

type SignalListener struct {
	signalQuit                int32
	taskToResponseForUserQuit chan interface{}
}

func (this *SignalListener) HasRecvQuitSignal() bool {
	return 0 != atomic.LoadInt32(&this.signalQuit)
}

//由 task 调用，对当前用户发出的 signal 进行响应
func (this *SignalListener) ResponseForUserQuit(resp interface{})  {
	//如果没有接收到过 user 的 quit 信号则此处不应该处理
	if this.HasRecvQuitSignal(){
		this.taskToResponseForUserQuit <- resp
	}
}

func NewSignalListener() *SignalListener {
	ret := SignalListener{}
	ret.signalQuit = 0
	ret.taskToResponseForUserQuit = make(chan interface{}, 1)
	return &ret
}

func (this *SignalListener) quitSignal(signal os.Signal)  {
	fmt.Println("user trigger quit, wait task to finished. (", signal,")")
	atomic.StoreInt32(&this.signalQuit, 1)
	resp := <- this.taskToResponseForUserQuit        //等待 task 完成后往此 channel 上发送信号
	fmt.Println("recive task response: ", resp, ", now quit")
	os.Exit(0)
}

func (this *SignalListener) WaitForSignal()  {
	c := make(chan os.Signal)

	signal.Notify(c, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
	go func() {
		for s := range c {
			switch s {
			case syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT:
				this.quitSignal(s)
				break;
			default:
				fmt.Println("unsupported signal", s)
			}
		}
	}()
}