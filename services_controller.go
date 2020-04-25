package main

import (
	"context"
	"errors"
	log "github.com/sirupsen/logrus"
	"sync"
)

type Service struct {
	Port  uint16 `json:"port" bson:"_id"`
	Name  string `json:"name" binding:"min=3" bson:"name"`
	Color string `json:"color" binding:"hexcolor" bson:"color"`
	Notes string `json:"notes" bson:"notes"`
}

type ServicesController struct {
	storage  Storage
	services map[uint16]Service
	mutex    sync.Mutex
}

func NewServicesController(storage Storage) *ServicesController {
	var result []Service
	if err := storage.Find(Services).All(&result); err != nil {
		log.WithError(err).Panic("failed to retrieve services")
		return nil
	}

	services := make(map[uint16]Service, len(result))
	for _, service := range result {
		services[service.Port] = service
	}

	return &ServicesController{
		storage:  storage,
		services: services,
	}
}

func (sc *ServicesController) SetService(c context.Context, service Service) error {
	sc.mutex.Lock()
	defer sc.mutex.Unlock()
	var upsert interface{}
	updated, err := sc.storage.Update(Services).Context(c).Filter(OrderedDocument{{"_id", service.Port}}).
		Upsert(&upsert).One(service)
	if err != nil {
		return errors.New("duplicate name")
	}
	if updated || upsert != nil {
		sc.services[service.Port] = service
	}
	return nil
}

func (sc *ServicesController) GetServices() map[uint16]Service {
	sc.mutex.Lock()
	services := make(map[uint16]Service, len(sc.services))
	for _, service := range sc.services {
		services[service.Port] = service
	}
	sc.mutex.Unlock()
	return services
}
