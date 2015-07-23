package main

import (
	"time"

	"github.com/qiniu/log.v1"
)

type sched struct {
	Name           string
	CheckFrequency time.Duration
	DueCheck       time.Time
	LastCheck      time.Time
	Parent         *repoSched
}

type repoSched struct {
	Scheds map[string]*sched
	Nc     chan interface{}
}

func (s *sched) checkNotifications() (ret time.Duration) {
	log.Info("checkNotifications in ", s.Name)
	ret = time.Duration(time.Second * 3)
	return
}

func (s *sched) save() {
	log.Info("save in ", s.Name)
}

func (s *sched) check() {
	log.Info("check in ", s.Name)
	s.checkNotifications()
	time.Sleep(time.Duration(time.Microsecond * 200))
	s.Parent.Nc <- s
	log.Println(">>>>>>> send s.nc from ", s.Name)
	return
}

func (s *repoSched) Run() error {

	s.Nc = make(chan interface{}, 1)

	go s.Poll()
	for {
		scheds, nearest, err := GetNearestCheck(s.Scheds)
		if err != nil {
			return err
		}
		wait := time.After(nearest)

		//log.Println("starting check")
		//now := time.Now()
		for _, s := range scheds {
			s.check()
			s.LastCheck = s.DueCheck
			//s.DueCheck = s.LastCheck.Add(s.CheckFrequency)
			//display(s)
		}
		//log.Println("wait for: ", nearest)
		//log.Println("................................")
		<-wait

	}
	return nil
}

// GetNearestCheck checkes all repo checkFrequence. It returns the
// duration and corresponding repoid until the nearest check triggers.
func GetNearestCheck(scheds map[string]*sched) (ret []*sched, dur time.Duration, err error) {

	var nearest time.Time
	now := time.Now()
	ret = make([]*sched, 0)
	nearest = now.Add(time.Duration(time.Hour * 0x000fffff)) //
	for k, s := range scheds {
		//log.Info("===>", scheds[k].Name, scheds[k].DueCheck)

		scheds[k].DueCheck = scheds[k].LastCheck.Add(scheds[k].CheckFrequency)

		//log.Info(k, scheds[k].LastCheck, scheds[k].CheckFrequency, scheds[k].DueCheck)
		t := scheds[k].DueCheck
		if t.Before(nearest) {
			nearest = t
			ret = ret[:0]
			ret = append(ret, s)
		} else if t.Equal(nearest) {
			ret = append(ret, s)
		}
	}

	if len(scheds) == 0 {
		dur = time.Duration(time.Hour * 1)
	} else {
		dur = nearest.Sub(now)
	}

	log.Info("---get nearest---", len(ret), dur)

	return

}

func (s *repoSched) Poll() {
	log.Println("=========start poll, start init check=======")

	for _, v := range s.Scheds {
		v.checkNotifications()
		v.save()
	}
	log.Println("============= init check complete")

	timeout := time.Duration(time.Minute * 5)
	for {
		select {
		case <-time.After(timeout):
			log.Println(">>>>>>> timeout", timeout)
		case val := <-s.Nc:
			ss, ok := val.(*sched)
			log.Println(">>>>>>> recieve s.nc from ", ss.Name)

			if ok {
				timeout = ss.checkNotifications()
				ss.save()
			}
		}
	}
}

func main() {

	now := time.Now()

	one := &sched{
		Name:           "one",
		CheckFrequency: time.Duration(time.Second * 10),
		LastCheck:      now,
	}
	two := &sched{
		Name:           "two",
		CheckFrequency: time.Duration(time.Second * 7),
		LastCheck:      now,
	}
	three := &sched{
		Name:           "three",
		CheckFrequency: time.Duration(time.Second * 3),
		LastCheck:      now,
	}

	ss := make(map[string]*sched, 3)
	ss["one"] = one
	ss["two"] = two
	ss["three"] = three

	reposched := &repoSched{
		Scheds: ss,
		Nc:     make(chan interface{}),
	}

	one.Parent = reposched
	one.DueCheck = now.Add(one.CheckFrequency)
	two.Parent = reposched
	two.DueCheck = now.Add(two.CheckFrequency)

	three.Parent = reposched
	three.DueCheck = now.Add(three.CheckFrequency)

	reposched.Run()
	//GetNearestCheck(reposched.Scheds)

}

func display(s *sched) {
	log.Printf(">>>>>>>>>>>>>> last=%s,due=%s\n", s.LastCheck.String(), s.DueCheck.String())
}
