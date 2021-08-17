package service

import (
	"context"
	"github.com/gorilla/mux"
	"sync"
)

type Service interface {
	RegisterRoutes(*mux.Router)
	Run(ctx context.Context)
}

type Manager struct {
	services []Service
}

func (m Manager) RegisterService(router *mux.Router, service Service) {
	service.RegisterRoutes(router)
	m.services = append(m.services, service)
}

func (m Manager) Run(ctx context.Context) {
	wg := sync.WaitGroup{}

	wg.Add(len(m.services))

	for i := 0; i < len(m.services); i++ {
		svc := m.services[i]

		go func() {
			svc.Run(ctx)
			wg.Done()
		}()
	}

	wg.Wait()
}
