package taskmanager

import (
	"fmt"
	//"sync"
	"time"
	"timer"
)

//只精确到秒，小于1s的定时任务会添加失败
type TaskManager struct {
	tm          *timer.TimerManager
	taskTimerId map[uint64]timer.TimerId // task id 和timerid的映射 ，task id由上游传递，需要全局唯一，解决tiemrid不是全局唯一的问题
}

func NewTaskManager() *TaskManager {
	schedule := &TaskManager{tm: timer.NewTimerManager(time.Second)}
	return schedule
}

func (t *TaskManager) Serve() {
	go func() {
		for {
			t.tm.DetectTimerInLock()
			time.Sleep(20 * time.Millisecond)
		}
	}()
}

// 定时回调是跨协程的 回调函数需要注意协程安全
func (t *TaskManager) RunAt(unix int64, f func(interface{})) (timer.TimerId, error) {
	now := time.Now().Unix()
	interval := unix - now
	if interval < 0 {
		err := fmt.Errorf("can run at past time!")
		return 0, err
	}
	_timer, Id := timer.NewTimer(timer.ONCE_TIMER)
	_timer.Start(uint64(interval), f, t.tm)
	return Id, nil
}

func (t *TaskManager) RunAfter(d time.Duration, f func(interface{})) (timer.TimerId, error) {
	interval := d / time.Second
	if interval <= 0 {
		err := fmt.Errorf("invalid interval time! %d", interval)
		return 0, err
	}
	_timer, Id := timer.NewTimer(timer.ONCE_TIMER)
	_timer.Start(uint64(interval), f, t.tm)
	return Id, nil
}

func (t *TaskManager) RunEvery(d time.Duration, f func(interface{})) (timer.TimerId, error) {
	interval := d / time.Second
	if interval <= 0 {
		err := fmt.Errorf("invalid interval time! %d", interval)
		return 0, err
	}
	_timer, Id := timer.NewTimer(timer.CIRCLE_TIMER)
	_timer.Start(uint64(interval), f, t.tm)
	return Id, nil
}

func (t *TaskManager) Update(Id timer.TimerId, d time.Duration, f func(interface{})) error {
	interval := d / time.Second
	if interval <= 0 {
		err := fmt.Errorf("invalid interval time! %d", interval)
		return err
	}
	_timer := t.tm.FindTimerById(Id)
	if _timer == nil {
		err := fmt.Errorf("find timer failed by  %v", Id)
		return err
	}
	_timer.Update(uint64(interval), f, t.tm)
	return nil
}

func (t *TaskManager) Stop(Id timer.TimerId) error {
	_timer := t.tm.FindTimerById(Id)
	if _timer == nil {
		err := fmt.Errorf("find timer failed by  %v", Id)
		return err
	}
	_timer.Stop(t.tm)
	return nil
}
