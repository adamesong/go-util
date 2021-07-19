package cron_tool

import (
	"fmt"
	"time"

	"github.com/adamesong/go-util/color"

	"github.com/robfig/cron/v3"
)

var (
	OneOffCron   *cron.Cron
	PeriodicCron *cron.Cron
)

// https://stackoverflow.com/questions/45942266/is-there-a-way-to-invoke-a-cron-job-on-a-certain-date-time
// 一次性任务的Schedule
type OneOffScheduler struct {
	At time.Time // 在什么时间运行
}

// 返回在时间t之后的下一次执行时间
// Next returns the next activation time, later than the given time.
// Next is invoked initially, and then each time the job is run.
func (ss *OneOffScheduler) Next(t time.Time) time.Time {
	if t.Before(ss.At) {
		return ss.At
	} else {
		return time.Time{}
	}
}

// 创建一次性任务的队列，并开始运行。
func InitOneOffCron() {
	OneOffCron = cron.New()
	OneOffCron.Start()
	fmt.Println(color.Blue("一次性计划任务已经开始运行。"))
}

// 创建周期性任务，并开始运行。
func InitPeriodicCron() {
	PeriodicCron = cron.New()
	PeriodicCron.Start()
	fmt.Println(color.Blue("周期性计划任务已经开始运行。"))
}

// 提供任务将在未来执行的时间、任务function，添加至一次性计划任务队列
func AddOneOffTask(executionTime time.Time, taskFunc func()) (entryID int) {
	s := OneOffScheduler{At: executionTime}
	entryID = int(OneOffCron.Schedule(&s, cron.FuncJob(taskFunc))) // 返回值是EntryID
	return
}

// 清理已完成的一次性计划任务
func CleanUpCompletedOneOffTasks() {
	for _, entry := range OneOffCron.Entries() {
		//fmt.Println("entry id:", entry.ID)
		//fmt.Println("entry prv: ", entry.Prev)
		//fmt.Println("next time: ", entry.Next)
		//fmt.Println("entry.schedule.next: ", entry.Schedule.Next(time.Now()))
		//fmt.Println("job: ", entry.Job)
		//fmt.Println("valid: ", entry.Valid())

		// 如果下次执行时间等于0值，则认为执行完成
		if entry.Next.Equal(time.Time{}) {
			// 清理这个已完成的entry
			OneOffCron.Remove(entry.ID)
		}
	}
}

// 根据entry_id取出这个entry，返回这个entry的下次执行时间。由于是个一次性任务，如果返回值是零值time.Time{}，则说明已经完成。
// Next time the job will run, or the zero time if Cron has not been
// started or this entry's schedule is unsatisfiable
func GetOneOffEntryNextExcTime(entryID int) time.Time {
	return OneOffCron.Entry(cron.EntryID(entryID)).Next
}

// 清除一个一次性计划任务
func RemoveOneOffEntry(entryID int) {
	OneOffCron.Remove(cron.EntryID(entryID))
}
