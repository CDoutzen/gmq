package gmq

import "fmt"

const (
	DefaultQueueName = "default"
)

const (
	QueueNameList = "queues"

	MsgStatePending    = "pending"
	MsgStateWaiting    = "waiting"
	MsgStateProcessing = "processing"
	MsgStateFailed     = "failed"

	QueueNamePaused = "paused"

	QueueNameDailyStatProcessed = "processed"
)

func NewKeyQueueList(ns string) string {
	return fmt.Sprintf("%s:%s", ns, QueueNameList)
}

func NewKeyQueuePending(ns, queueName string) string {
	return fmt.Sprintf("%s:%s:%s", ns, queueName, MsgStatePending)
}

func NewKeyMsgDetail(ns, queueName, msgId string) string {
	return fmt.Sprintf("%s:%s:msg:%s", ns, queueName, msgId)
}

func NewKeyQueuePaused(ns, queueName string) string {
	return fmt.Sprintf("%s:%s:%s", ns, queueName, QueueNamePaused)
}

func NewKeyQueueProcessing(ns, queueName string) string {
	return fmt.Sprintf("%s:%s:%s", ns, queueName, MsgStateProcessing)
}

func NewKeyQueueWaiting(ns, queueName string) string {
	return fmt.Sprintf("%s:%s:%s", ns, queueName, MsgStateWaiting)
}

func NewKeyQueueFailed(ns, queueName string) string {
	return fmt.Sprintf("%s:%s:%s", ns, queueName, MsgStateFailed)
}

// gmq:<queueName>:processed:<YYYY-MM-DD>
func NewKeyDailyStatProcessed(ns, queueName, YYYYMMDD string) string {
	return fmt.Sprintf("%s:%s:%s:%s", ns, queueName, QueueNameDailyStatProcessed, YYYYMMDD)
}