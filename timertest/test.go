package main

import (
	"fmt"
	"taskmanager"
	"time"
)

func ontimer(v interface{}) {
	fmt.Println("time ", time.Now().Unix(), "-----------------hello------------------------------------", "id ", v)
}

func ontime2r(v interface{}) {
	fmt.Println("time ", time.Now().Unix(), "-----------------ontime2r hello------------------------------------", "id ", v)
}

func main() {
	sch := taskmanager.NewTaskManager()
	sch.Serve()
	for i := 1; i < 10000; i++ {
		_, err := sch.RunAt(time.Now().Unix()+int64(i), ontimer)
		if err != nil {
			fmt.Println(err)
		}
		//err = sch.RunAt(time.Second, ontime2r)
		//if err != nil {
		//	fmt.Println(err)
		//}
	}
	fmt.Println("add finish")
	/*tm := timer.NewTimerManager(time.Second)
	go func() {
		for {
			tm.DetectTimerInLock()
			//time.Sleep(time.Nanosecond)
		}
	}()

	for i := 0; i < 2; i++ {
		//go func() {
		//	timer1, _ := timer.NewTimer(timer.ONCE_TIMER)
		//	timer1.Start(uint64(i%5)+1, ontimer, tm)
		//}()
		//go func() {
		timer2, _ := timer.NewTimer(timer.CIRCLE_TIMER)
		timer2.I = i
		err := timer2.Start(5, ontime2r, tm)
		if err != nil {
			fmt.Println(err)
		}

		//}()
	}
	time.Sleep(500 * time.Millisecond)
	timer3 := tm.FindTimerById(2)
	fmt.Println(timer3)
	//timer3.Update(8, ontimer, tm)*/
	for {
		time.Sleep(500 * time.Millisecond)
	}

}
