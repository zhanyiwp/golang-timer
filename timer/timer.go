package timer

import (
	"fmt"
	"sync"
	"sync/atomic"
	"time"
)

const (
	TVR_BITS = 8
	TVN_BITS = 6
	TVN_SIZE = 1 << TVN_BITS
	TVR_SIZE = 1 << TVR_BITS
	TVN_MASK = TVN_SIZE - 1
	TVR_MASK = TVR_SIZE - 1
)
const (
	ONCE_TIMER   = 0
	CIRCLE_TIMER = 1
)

func offset(n uint64) uint64 {
	return TVR_SIZE + n*TVN_SIZE
}

func index(v, n uint64) uint64 {
	return ((v >> (TVR_BITS + n*TVN_BITS)) & TVN_MASK)
}

type TimerId uint64

type Timer struct {
	interval uint64
	expire   uint64
	index    int32
	id       TimerId
	_type    int32
	f        func(interface{})
	// for test
	Str string
	I   int
}

type TimerManager struct {
	//采用hashmap 主要考虑到需要快速update 带来的影响是需要同时触发的定时器不是先进先出的。
	tv        [TVR_SIZE + 4*TVN_SIZE]map[TimerId]*Timer
	checktime uint64
	/* 定时器精度 ，确定了定时范围 (0-2^32)*tick.精度越高,定时越不准。
	* 5us以下的定时可能不准确,不支持 0间隔定时器 interval=0 会直接添加失败
	 */
	tick    time.Duration
	idIndex map[TimerId]int32
	mutex   sync.Mutex
}

var (
	id uint64
)

func NewTimerId() TimerId {
	atomic.AddUint64(&id, 1)
	return TimerId(id)
}

func NewTimer(timertype int32) (*Timer, TimerId) {
	_id := NewTimerId()
	timer := &Timer{id: _id, _type: timertype}
	return timer, _id
}

func NewTimerManager(tick_ time.Duration) *TimerManager {
	tm := &TimerManager{tick: tick_}
	tm.mutex.Lock()
	tm.checktime = uint64(time.Now().UnixNano() / int64(tick_))
	for i, _ := range tm.tv {
		tm.tv[i] = make(map[TimerId]*Timer)
	}
	tm.idIndex = make(map[TimerId]int32)
	tm.mutex.Unlock()
	return tm
}

func (t *Timer) Stop(tm *TimerManager) {
	if t.index != -1 {
		tm.RemoveTimerInLock(t)
		t.index = -1
	}
}

func (t *Timer) Start(interval uint64, f func(interface{}), tm *TimerManager) error {
	if interval == 0 {
		err := fmt.Errorf("invalid interval time! %d", interval)
		return err
	}
	t.Stop(tm)
	t.interval = interval
	t.f = f
	t.expire = interval + uint64(time.Now().UnixNano()/int64(tm.tick))
	tm.AddTimerInLock(t)
	return nil
}

func (t *Timer) Update(interval uint64, f func(interface{}), tm *TimerManager) {
	t.Start(interval, f, tm)
}

func (t *TimerManager) AddTimerInLock(timer *Timer) {
	t.mutex.Lock()
	t.addTimer(timer)
	t.mutex.Unlock()
}

func (t *TimerManager) addTimer(timer *Timer) {
	expires := timer.expire
	idx := expires - t.checktime
	if idx < TVR_SIZE {
		timer.index = int32(expires & TVR_MASK)
	} else if idx < 1<<(TVR_BITS+TVN_BITS) {
		timer.index = int32(offset(0) + index(expires, 0))
	} else if idx < 1<<(TVR_BITS+2*TVN_BITS) {
		timer.index = int32(offset(1) + index(expires, 1))
	} else if idx < 1<<(TVR_BITS+3*TVN_BITS) {
		timer.index = int32(offset(2) + index(expires, 2))
	} else if int64(idx) < 0 {
		timer.index = int32(t.checktime & TVR_MASK)
	} else {
		if idx > 0xffffffff {
			idx = 0xffffffff
			expires = idx + t.checktime
		}
		timer.index = int32(offset(3) + index(expires, 3))
	}
	t.idIndex[timer.id] = timer.index
	timermap := t.tv[timer.index]
	timermap[timer.id] = timer
}

func (t *TimerManager) removeTimer(timer *Timer) {
	timermap := t.tv[timer.index]
	delete(timermap, timer.id)
	delete(t.idIndex, timer.id)
}

func (t *TimerManager) RemoveTimerInLock(timer *Timer) {
	t.mutex.Lock()
	t.removeTimer(timer)
	t.mutex.Unlock()
}

func (t *TimerManager) detectTimer() {
	now := uint64(time.Now().UnixNano() / int64(t.tick))
	for tnow := now; t.checktime <= tnow; t.checktime++ {
		index1 := t.checktime & TVR_MASK
		if index1 == 0 &&
			t.cascade(offset(0), index(t.checktime, 0)) == 0 &&
			t.cascade(offset(1), index(t.checktime, 1)) == 0 &&
			t.cascade(offset(2), index(t.checktime, 2)) == 0 {
			t.cascade(offset(3), index(t.checktime, 3))
		}
		timermap := t.tv[index1]
		for k, v := range timermap {
			v.f(v.id)
			delete(timermap, k)
			if v._type == CIRCLE_TIMER {
				v.expire = v.interval + now
				t.addTimer(v)
			}
		}
	}
}

func (t *TimerManager) cascade(offset, index uint64) uint64 {
	timermap := t.tv[offset+index]
	for k, v := range timermap {
		t.addTimer(v)
		delete(timermap, k)
	}
	return index
}

func (t *TimerManager) DetectTimerInLock() {
	t.mutex.Lock()
	t.detectTimer()
	t.mutex.Unlock()
}

func (t *TimerManager) FindTimerById(Id TimerId) *Timer {
	t.mutex.Lock()
	defer t.mutex.Unlock()
	if v, ok := t.idIndex[Id]; ok {
		timermap := t.tv[v]
		if v1, ok := timermap[Id]; ok {
			return v1
		} else {
			return nil
		}
	} else {
		return nil
	}

}
