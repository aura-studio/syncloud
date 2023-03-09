package pusher

import (
	"log"
	"net/url"
)

type Pusher struct {
	*TaskList
}

func New(taskList *TaskList) *Pusher {
	return &Pusher{
		TaskList: taskList,
	}
}

func (p *Pusher) Push() {
	for s, tasks := range p.Tasks {
		p.newRemote(s).Push(tasks)
	}
}

func (p *Pusher) newRemote(s string) Remote {
	u, err := url.Parse(s)
	if err != nil {
		log.Panicf("parsing remote url error: %v", err)
	}

	switch u.Scheme {
	case "s3":
		return NewS3Remote(u.Host)
	default:
		log.Panicf("unknown remote scheme: %s", u.Scheme)
	}

	return nil
}
