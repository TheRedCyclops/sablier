package sablier

import (
	"context"
	"github.com/sablierapp/sablier/pkg/promise"
	"github.com/sablierapp/sablier/pkg/provider"
	"github.com/sablierapp/sablier/pkg/tinykv"
	log "github.com/sirupsen/logrus"
	"sync"
	"time"
)

type Sablier struct {
	Provider    provider.Provider
	promises    map[string]*promise.Promise[Instance]
	expirations tinykv.KV[string]

	lock *sync.RWMutex
}

func NewSablier(ctx context.Context, provider provider.Provider) *Sablier {
	lock := &sync.RWMutex{}
	promises := make(map[string]*promise.Promise[Instance])
	expirations := tinykv.New(time.Second, func(k string, _ string) {
		lock.Lock()
		defer lock.Unlock()
		log.Printf("instance [%s] expired - removing from promises", k)
		err := provider.Stop(ctx, k)
		if err != nil {
			log.Printf("error stopping instance [%s]: %v", k, err)
		}
		delete(promises, k)
	})
	go func() {
		<-ctx.Done()
		expirations.Stop()
	}()
	return &Sablier{
		Provider:    provider,
		promises:    promises,
		expirations: expirations,
		lock:        lock,
	}
}
