package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"time"

	"github.com/giant-stone/go/glogging"
	"github.com/giant-stone/go/gtime"
	"github.com/giant-stone/go/gutil"

	"github.com/giant-stone/gmq/gmq"
)

var (
	cmdPrintStats  bool
	cmdAddMsg      bool
	cmdGetMsg      bool
	cmdListMsg     bool
	cmdDelMsg      bool
	cmdStatsWeekly bool
	cmdDelQueue    bool
	cmdPauseq      string
	cmdResumeq     string
	dsnRedis       string
	msgId          string
	payloadStr     string
	queueName      string
	state          string

	limit  int64
	offset uint64

	loglevel string
)

func main() {
	flag.StringVar(&loglevel, "l", "", "loglevel debug,info,warn,error")
	flag.StringVar(&dsnRedis, "d", "redis://127.0.0.1:6379/0", "redis DSN")

	flag.BoolVar(&cmdPrintStats, "stats", false, "print queue stats")
	flag.BoolVar(&cmdAddMsg, "add", false, "append a message into queue")
	flag.BoolVar(&cmdGetMsg, "getmsg", false, "get a message detail")
	flag.BoolVar(&cmdListMsg, "listmsg", false, "get a message detail")
	flag.BoolVar(&cmdDelMsg, "delmsg", false, "delete a message from queue")
	flag.BoolVar(&cmdDelQueue, "delqueue", false, "delete a message from queue")
	flag.StringVar(&cmdPauseq, "pause", "", "queuename to pause")
	flag.StringVar(&cmdResumeq, "resume", "", "queuename to resume")

	flag.StringVar(&queueName, "q", gmq.DefaultQueueName, "queue name")
	flag.StringVar(&payloadStr, "p", "", "message payload in JSON")
	flag.StringVar(&msgId, "i", "", "message id, it is auto-generated by default")
	flag.StringVar(&state, "state", "", "must be one of pending|processing|failed, required for -listmsg, queue state to search")

	flag.Int64Var(&limit, "limit", -1, "use with -listmsg, maximum number of messages to display, default to display all, if limit <=0, display all the messages after offset ")
	flag.Uint64Var(&offset, "offset", 0, "use with -listmsg, first messages offset to display, start with 0")

	flag.Parse()

	// cmdDelQueue = true
	// queueName = "nx"
	// dsnRedis = "redis://127.0.0.1:6379/0"

	glogging.Init([]string{"stdout"}, loglevel)

	if !cmdPrintStats && !cmdAddMsg && !cmdGetMsg && !cmdListMsg && !cmdDelMsg && !cmdDelQueue && (cmdPauseq != "") && (cmdResumeq != "") && !cmdStatsWeekly {
		flag.PrintDefaults()
		os.Exit(1)
	}

	broker, err := gmq.NewBrokerRedis(dsnRedis)
	gutil.ExitOnErr(err)
	ctx := context.Background()
	if cmdPauseq != "" {
		pauseQueue(ctx, broker, cmdPauseq)
		os.Exit(0)
	} else if cmdResumeq != "" {
		resumeQueue(ctx, broker, cmdResumeq)
		os.Exit(0)
	}

	if cmdPrintStats {
		printStats(ctx, broker)
	} else if cmdStatsWeekly {
		printStatsWeekly(ctx, broker)
	} else if cmdAddMsg {
		addMsg(ctx, broker, queueName, payloadStr, msgId)
	} else if cmdGetMsg {
		getMsg(ctx, broker, queueName, msgId)
	} else if cmdListMsg {
		if state == "" {
			flag.PrintDefaults()
			os.Exit(1)
		}
		listMsg(ctx, broker, queueName)
	} else if cmdDelMsg {
		delMsg(ctx, broker, queueName, msgId)
	} else if cmdDelQueue {
		delQueue(ctx, broker, queueName)
	} else {
		flag.PrintDefaults()
		os.Exit(1)
	}

	fmt.Print("\n")
}

func pauseQueue(ctx context.Context, broker gmq.Broker, queuename string) {
	if err := broker.Pause(ctx, queuename); err != nil {
		fmt.Printf("Pausing queue %s  failed. errer(%s) \n", queuename, err.Error())
	} else {
		fmt.Printf("Pause queue %s \n", queuename)
	}
}
func resumeQueue(ctx context.Context, broker gmq.Broker, queuename string) {
	if err := broker.Resume(ctx, queuename); err != nil {
		fmt.Printf("Resuming queue %s failed. errer(%s) \n", cmdPauseq, err.Error())
	} else {
		fmt.Printf("Resume queue %s \n", queuename)
	}
}

func addMsg(ctx context.Context, broker gmq.Broker, queueName, payloadStr, id string) {
	if payloadStr == "" && id == "" {
		fmt.Println("payload or id is required")
		return
	}

	rs, err := broker.Enqueue(ctx, &gmq.Msg{
		Payload: []byte(payloadStr),
		Id:      id,
	}, gmq.OptQueueName(queueName))
	gutil.ExitOnErr(err)

	dat, _ := json.Marshal(rs)
	fmt.Println("reply", string(dat))
}

func printStats(ctx context.Context, broker gmq.Broker) {
	queues, err := broker.GetStats(ctx)
	gutil.ExitOnErr(err)

	fmt.Println("")
	fmt.Print("# gmq stats \n\n")
	if len(queues) == 0 {
		fmt.Println("Related info not found. Do consumer(s) have not start yet?")
	} else {
		for _, rsStat := range queues {
			fmt.Printf("queue=%s total=%d pending=%d waiting=%d processing=%d failed=%d \n",
				rsStat.Name,
				rsStat.Total,
				rsStat.Pending,
				rsStat.Waiting,
				rsStat.Processing,
				rsStat.Failed,
			)
		}
		printStatsWeekly(ctx, broker)
	}
}

func printStatsWeekly(ctx context.Context, broker gmq.Broker) {
	dayInfo, totalInfo, err := broker.GetStatsWeekly(ctx)
	gutil.ExitOnErr(err)
	now := time.Now()
	fmt.Printf("\n## Weekly Statistic: %s ~ %s \n\n",
		gtime.UnixTime2YyyymmddUtc(now.AddDate(0, 0, -7).Unix()),
		gtime.UnixTime2YyyymmddUtc(now.Unix()))
	for i := range *dayInfo {
		fmt.Printf("date:%s processed: %d, failed: %d, total: %d \n", (*dayInfo)[i].Date, (*dayInfo)[i].Processed, (*dayInfo)[i].Failed, (*dayInfo)[i].Processed+(*dayInfo)[i].Failed)
	}
	fmt.Printf("Total processed: %d, Total failed: %d, total: %d \n", totalInfo.Processed, totalInfo.Failed, totalInfo.Processed+totalInfo.Failed)
}

// ./gmqcli.bin -getmsg -i <msgId> -q <queueName>
func getMsg(ctx context.Context, broker gmq.Broker, queueName, msgId string) {
	msg, err := broker.GetMsg(ctx, queueName, msgId)
	if err != nil {
		if err == gmq.ErrNoMsg {
			fmt.Printf("message matched queue=%s id=%s not found", queueName, msgId)
			return
		}
		gutil.ExitOnErr(err)
	}

	fmt.Println("RAW\n", msg)

	dat, err := json.MarshalIndent(msg, "", "  ")
	gutil.ExitOnErr(err)
	fmt.Println("INTERNAL\n", string(dat))
}

// ./gmqcli.bin -listmsg <queueName> -state failed [-limit 10 -offset 10]
func listMsg(ctx context.Context, broker gmq.Broker, queueName string) {
	msgs, err := broker.ListMsg(ctx, queueName, state, int64(offset), limit)
	if err != nil {
		if err == gmq.ErrNoMsg {
			fmt.Printf("message matched queue=%s state=%s not found", queueName, state)
			return
		}
		gutil.ExitOnErr(err)
	}

	fmt.Printf("list queue: %s\n\n", gmq.NewKeyQueueState(gmq.Namespace, queueName, state))
	for _, msg := range msgs {
		fmt.Println(msg)
	}
}

func delMsg(ctx context.Context, broker gmq.Broker, queueName, msgId string) {
	err := broker.DeleteMsg(ctx, queueName, msgId)
	if err != nil {
		if err != gmq.ErrNoMsg {
			gutil.ExitOnErr(err)
		}
	}
	fmt.Printf("queue=%s msgId=%s deleted \n", queueName, msgId)
}

func delQueue(ctx context.Context, broker gmq.Broker, queueName string) {
	err := broker.DeleteQueue(ctx, queueName)
	if err != nil {
		if err != gmq.ErrNoMsg {
			gutil.ExitOnErr(err)
		}
	}
	fmt.Printf("queue %s cleared", queueName)
}
