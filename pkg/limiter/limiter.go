package limiter

import (
	"log"
	"time"

	"golang.org/x/time/rate"
)

type Limiter struct {
	Clients map[string]*Client
	window  rate.Limit
	limit   int
}

type Client struct {
	limiter  *rate.Limiter
	lastSeen time.Time
}

func NewLimiter() *Limiter {
	limit := 10.0
	duration := 24 * time.Hour
	window := limit / duration.Seconds()

	l := &Limiter{
		Clients: make(map[string]*Client),
		limit:   10,
		window:  rate.Limit(window),
	}
	return l
}

func (l *Limiter) AddClient(key string) {
	_, exists := l.Clients[key]
	if exists {
		log.Printf("Client %s already exists. Updating last seen time\n", key)
		l.UpdateClient(key)
		return
	}

	c := &Client{
		limiter:  rate.NewLimiter(l.window, l.limit),
		lastSeen: time.Now(),
	}

	l.Clients[key] = c
	log.Printf("Client %s added to limiter\n", key)
}

func (l *Limiter) RemoveClients() {
	log.Println("Removing clients")
	for k, v := range l.Clients {
		dur := time.Since(v.lastSeen)
		if dur > 30*time.Second {
			delete(l.Clients, k)
		}
	}
}

func (l *Limiter) UpdateClient(key string) {
	c, ok := l.Clients[key]
	if !ok {
		return
	}
	c.lastSeen = time.Now()
}

func (l *Limiter) ClientAllowed(key string) bool {
	l.AddClient(key)

	allowed := l.Clients[key].limiter.Allow()
	return allowed
}
