package main

import (
	"math/rand"
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
	Scheds   map[string]*sched
	Nc       chan interface{}
	Timeouts map[string]time.Duration
}

var r = rand.New(rand.NewSource(43))

func (s *sched) checkNotifications() (ret time.Duration) {

	ret = time.Duration(time.Second * time.Duration(r.Intn(10)))
	log.Infof("checkNotifications in %s, return dur:%v \n", s.Name, ret)
	return
}

func (s *sched) save() {
	log.Info("save in ", s.Name)
}

func (s *sched) check() {
	log.Info("check in ", s.Name)
	//time.Sleep(time.Duration(time.Microsecond * 200))
	s.Parent.Nc <- s
	log.Println(">>>>>>> send s.nc from ", s.Name)
	return
}

func (s *repoSched) Run() error {
	now := time.Now()
	s.Nc = make(chan interface{}, 1)

	log.Println("========= start init check=======")

	for k, v := range s.Scheds {
		timeout := v.checkNotifications()
		v.LastCheck = now
		s.Timeouts[k] = timeout
		v.save()
	}

	log.Println("============= init check complete")

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
		if dur < 0 {
			log.Fatal("dur is nagtive", dur)
		}
	}

	log.Info("---get nearest---", len(ret), dur)

	return

}

func (s *repoSched) Poll() {

	go s.ContinuousCheckRepoNotifications()
	for {
		select {
		case val := <-s.Nc:
			ss, ok := val.(*sched)
			log.Println(">>>>>>> recieve s.nc from ", ss.Name)
			if ok {
				ss.checkNotifications()
				ss.save()
			}
		}
	}
}

func (s *repoSched) ContinuousCheckRepoNotifications() {

	for {
		repoids, timeout := s.getMinDuration()
		log.Printf("get min duration:%v ", timeout)
		log.Println("repoids ", repoids)
		select {
		case <-time.After(timeout):
			for _, repoid := range repoids {
				timeout := s.Scheds[repoid].checkNotifications()
				s.Scheds[repoid].save()
				s.Timeouts[repoid] = timeout
			}
		}
	}
}

func (s *repoSched) getMinDuration() (ret []string, timeout time.Duration) {

	//TODO 提高效率
	//get min
	timeout = time.Duration(time.Hour * 0xffff)
	for k, v := range s.Timeouts {
		if v < timeout {
			timeout = v
			ret = ret[:0]
			ret = append(ret, k)
		} else if v == timeout {
			ret = append(ret, k)
		}
	}

	//map[string]duration.sub(min)
	for k, v := range s.Timeouts {
		s.Timeouts[k] = v - timeout
	}
	return
}

func main() {

	now := time.Now()

	one := &sched{
		Name:           "one",
		CheckFrequency: time.Duration(time.Second * 1000),
		LastCheck:      now,
	}
	two := &sched{
		Name:           "two",
		CheckFrequency: time.Duration(time.Second * 700),
		LastCheck:      now,
	}
	three := &sched{
		Name:           "three",
		CheckFrequency: time.Duration(time.Second * 300),
		LastCheck:      now,
	}

	ss := make(map[string]*sched, 3)
	ss["one"] = one
	ss["two"] = two
	ss["three"] = three

	reposched := &repoSched{
		Scheds:   ss,
		Nc:       make(chan interface{}),
		Timeouts: make(map[string]time.Duration),
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
