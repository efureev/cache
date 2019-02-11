package cache

import "time"

type cg struct {
	Interval time.Duration
	stop     chan bool
}

func (j *cg) Run(c *Cache) {
	ticker := time.NewTicker(j.Interval)
	for {
		select {
		case <-ticker.C:
			c.DeleteExpired()
		case <-j.stop:
			ticker.Stop()
			return
		}
	}
}

func stopCG(c *Cache) {
	c.cg.stop <- true
}

func runCG(cache *Cache, ci time.Duration) {
	j := &cg{
		Interval: ci,
		stop:     make(chan bool),
	}
	cache.cg = j
	go j.Run(cache)
}
