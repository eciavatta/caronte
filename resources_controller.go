package main

import (
	"context"
	"github.com/gin-gonic/gin"
	"github.com/shirou/gopsutil/cpu"
	"github.com/shirou/gopsutil/disk"
	"github.com/shirou/gopsutil/mem"
	log "github.com/sirupsen/logrus"
	"sync"
	"time"
)

const (
	averageCPUPercentAlertThreshold = 90.0
	averageCPUPercentAlertMinInterval = 120.0
)

type SystemStats struct {
	VirtualMemory *mem.VirtualMemoryStat `json:"virtual_memory"`
	CPUTimes      []cpu.TimesStat        `json:"cpu_times"`
	CPUPercents   []float64              `json:"cpu_percents"`
	DiskUsage     *disk.UsageStat        `json:"disk_usage"`
}

type ResourcesController struct {
	notificationController *NotificationController
	lastCPUPercent         []float64
	mutex                  sync.Mutex
}

func NewResourcesController(notificationController *NotificationController) *ResourcesController {
	return &ResourcesController{
		notificationController: notificationController,
	}
}

func (csc *ResourcesController) GetProcessStats(c context.Context) interface{} {
	return nil
}

func (csc *ResourcesController) GetSystemStats(c context.Context) SystemStats {
	virtualMemory, err := mem.VirtualMemoryWithContext(c)
	if err != nil {
		log.WithError(err).Panic("failed to retrieve virtual memory")
	}
	cpuTimes, err := cpu.TimesWithContext(c, true)
	if err != nil {
		log.WithError(err).Panic("failed to retrieve cpu times")
	}
	diskUsage, err := disk.UsageWithContext(c, "/")
	if err != nil {
		log.WithError(err).Panic("failed to retrieve disk usage")
	}

	defer csc.mutex.Unlock()
	csc.mutex.Lock()

	return SystemStats{
		VirtualMemory: virtualMemory,
		CPUTimes:      cpuTimes,
		DiskUsage:     diskUsage,
		CPUPercents:   csc.lastCPUPercent,
	}
}

func (csc *ResourcesController) Run() {
	interval, _ := time.ParseDuration("3s")
	var lastAlertTime time.Time

	for {
		cpuPercent, err := cpu.Percent(interval, true)
		if err != nil {
			log.WithError(err).Error("failed to retrieve cpu percent")
			return
		}

		csc.mutex.Lock()
		csc.lastCPUPercent = cpuPercent
		csc.mutex.Unlock()

		avg := Average(cpuPercent)
		if avg > averageCPUPercentAlertThreshold && time.Now().Sub(lastAlertTime).Seconds() > averageCPUPercentAlertMinInterval {
			csc.notificationController.Notify("resources.cpu_alert", "alert", gin.H{
				"cpu_percent": cpuPercent,
			})
			log.WithField("cpu_percent", cpuPercent).Warn("cpu percent usage has exceeded the limit threshold")
			lastAlertTime = time.Now()
		}
	}
}
