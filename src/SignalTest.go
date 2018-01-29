package main

import (
	"dbOptions"
	"fmt"
	"time"
)

func main()  {
	sigLisenter := dbOptions.NewSignalListener()
	sigLisenter.WaitForSignal()

	count := 0
	for{
		time.Sleep(time.Second)
		fmt.Println("wait user to quit")
		if(sigLisenter.HasRecvQuitSignal()){
			sigLisenter.ResponseForUserQuit(nil)
			fmt.Println("response for quit")
			break
		}
		count ++

		if count > 5{
			sigLisenter.StopWait()
		}
	}
}
