package components

import (
	"time"
)

type Channel struct {
	ID           string
	Name         string
	Topic        string
	Type         string
	UserID       string
	Presence     string
	Notification bool
}

type Messages []Message

func (a Messages) Len() int           { return len(a) }
func (a Messages) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a Messages) Less(i, j int) bool { return a[i].Time.After(a[j].Time) }

type Message struct {
	ThreadTimestamp string
	Time            time.Time
	Channel         *Channel
	Name            string
	Content         string
	Attachments     []string
	IsReply bool
}
