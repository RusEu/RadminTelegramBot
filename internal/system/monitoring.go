package system

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/disk"
	"github.com/shirou/gopsutil/v3/mem"
	log "github.com/sirupsen/logrus"

	"github.com/RusEu/RadminTelegramBot/internal/config"
)

// MonitoringManager handles system monitoring and alerts
type MonitoringManager struct {
	config      *config.Config
	logger      *log.Logger
	isRunning   bool
	lastAlerts  map[string]time.Time
	mu          sync.RWMutex
}

// AlertCallback is the function signature for alert callbacks
type AlertCallback func(alertType, message string)

// MonitoringData represents monitoring metrics
type MonitoringData struct {
	Timestamp        time.Time
	CPUUsage         float64
	MemoryUsage      float64
	DiskUsage        float64
	LoadAverage      []float64
}

// NewMonitoringManager creates a new monitoring manager
func NewMonitoringManager(cfg *config.Config, logger *log.Logger) *MonitoringManager {
	return &MonitoringManager{
		config:     cfg,
		logger:     logger,
		isRunning:  false,
		lastAlerts: make(map[string]time.Time),
	}
}

// StartMonitoring starts the monitoring loop
func (mm *MonitoringManager) StartMonitoring(ctx context.Context, alertCallback AlertCallback) {
	mm.mu.Lock()
	if mm.isRunning {
		mm.mu.Unlock()
		mm.logger.Warn("📊 Monitoring already running")
		return
	}
	mm.isRunning = true
	mm.mu.Unlock()

	mm.logger.Info("📊 Starting system monitoring...")
	mm.logger.Infof("📈 Monitoring interval: %d seconds", mm.config.Monitoring.MonitoringInterval)
	mm.logger.Infof("🚨 CPU alert threshold: %.1f%%", mm.config.Monitoring.CPUAlertThreshold)
	mm.logger.Infof("🚨 Memory alert threshold: %.1f%%", mm.config.Monitoring.MemoryAlertThreshold)
	mm.logger.Infof("🚨 Disk alert threshold: %.1f%%", mm.config.Monitoring.DiskAlertThreshold)

	ticker := time.NewTicker(time.Duration(mm.config.Monitoring.MonitoringInterval) * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			mm.logger.Info("📊 Monitoring stopped")
			mm.mu.Lock()
			mm.isRunning = false
			mm.mu.Unlock()
			return
		case <-ticker.C:
			mm.checkSystemHealth(alertCallback)
		}
	}
}

// checkSystemHealth checks system health and sends alerts if needed
func (mm *MonitoringManager) checkSystemHealth(alertCallback AlertCallback) {
	mm.logger.Debug("🔍 Checking system health...")

	// Get monitoring data
	data, err := mm.collectMonitoringData()
	if err != nil {
		mm.logger.Errorf("❌ Failed to collect monitoring data: %v", err)
		return
	}

	mm.logger.Debugf("📊 CPU: %.1f%%, Memory: %.1f%%, Disk: %.1f%%", 
		data.CPUUsage, data.MemoryUsage, data.DiskUsage)

	// Check CPU usage
	if data.CPUUsage > mm.config.Monitoring.CPUAlertThreshold {
		alertType := "cpu"
		if mm.shouldSendAlert(alertType) {
			message := fmt.Sprintf("CPU usage is high: %.1f%% (threshold: %.1f%%)", 
				data.CPUUsage, mm.config.Monitoring.CPUAlertThreshold)
			mm.logger.Warnf("🔥 %s", message)
			alertCallback(alertType, message)
			mm.updateLastAlert(alertType)
		}
	}

	// Check Memory usage
	if data.MemoryUsage > mm.config.Monitoring.MemoryAlertThreshold {
		alertType := "memory"
		if mm.shouldSendAlert(alertType) {
			message := fmt.Sprintf("Memory usage is high: %.1f%% (threshold: %.1f%%)", 
				data.MemoryUsage, mm.config.Monitoring.MemoryAlertThreshold)
			mm.logger.Warnf("💾 %s", message)
			alertCallback(alertType, message)
			mm.updateLastAlert(alertType)
		}
	}

	// Check Disk usage
	if data.DiskUsage > mm.config.Monitoring.DiskAlertThreshold {
		alertType := "disk"
		if mm.shouldSendAlert(alertType) {
			message := fmt.Sprintf("Disk usage is high: %.1f%% (threshold: %.1f%%)", 
				data.DiskUsage, mm.config.Monitoring.DiskAlertThreshold)
			mm.logger.Warnf("💿 %s", message)
			alertCallback(alertType, message)
			mm.updateLastAlert(alertType)
		}
	}

	// Check Load Average (if too high)
	if len(data.LoadAverage) > 0 && data.LoadAverage[0] > 10.0 {
		alertType := "load"
		if mm.shouldSendAlert(alertType) {
			message := fmt.Sprintf("System load is very high: %.2f (1min average)", data.LoadAverage[0])
			mm.logger.Warnf("⚖️ %s", message)
			alertCallback(alertType, message)
			mm.updateLastAlert(alertType)
		}
	}
}

// collectMonitoringData collects current system metrics
func (mm *MonitoringManager) collectMonitoringData() (*MonitoringData, error) {
	data := &MonitoringData{
		Timestamp: time.Now(),
	}

	// Get CPU usage
	cpuPercent, err := cpu.Percent(time.Second, false)
	if err != nil {
		return nil, fmt.Errorf("failed to get CPU usage: %w", err)
	}
	if len(cpuPercent) > 0 {
		data.CPUUsage = cpuPercent[0]
	}

	// Get Memory usage
	memStat, err := mem.VirtualMemory()
	if err != nil {
		return nil, fmt.Errorf("failed to get memory usage: %w", err)
	}
	data.MemoryUsage = memStat.UsedPercent

	// Get Disk usage
	diskStat, err := disk.Usage("/")
	if err != nil {
		return nil, fmt.Errorf("failed to get disk usage: %w", err)
	}
	data.DiskUsage = diskStat.UsedPercent

	return data, nil
}

// shouldSendAlert checks if an alert should be sent based on cooldown
func (mm *MonitoringManager) shouldSendAlert(alertType string) bool {
	mm.mu.RLock()
	lastAlert, exists := mm.lastAlerts[alertType]
	mm.mu.RUnlock()

	if !exists {
		return true
	}

	cooldown := time.Duration(mm.config.Monitoring.AlertCooldown) * time.Second
	return time.Since(lastAlert) > cooldown
}

// updateLastAlert updates the last alert time for a specific alert type
func (mm *MonitoringManager) updateLastAlert(alertType string) {
	mm.mu.Lock()
	mm.lastAlerts[alertType] = time.Now()
	mm.mu.Unlock()
}

// GetCurrentMetrics returns current system metrics
func (mm *MonitoringManager) GetCurrentMetrics() (*MonitoringData, error) {
	mm.logger.Debug("📊 Getting current metrics...")
	return mm.collectMonitoringData()
}

// IsRunning returns whether monitoring is currently running
func (mm *MonitoringManager) IsRunning() bool {
	mm.mu.RLock()
	defer mm.mu.RUnlock()
	return mm.isRunning
}

// GetAlertHistory returns the last alert times
func (mm *MonitoringManager) GetAlertHistory() map[string]time.Time {
	mm.mu.RLock()
	defer mm.mu.RUnlock()
	
	history := make(map[string]time.Time)
	for alertType, lastTime := range mm.lastAlerts {
		history[alertType] = lastTime
	}
	return history
}

// FormatMetrics formats monitoring data for display
func (mm *MonitoringManager) FormatMetrics(data *MonitoringData) string {
	return fmt.Sprintf(`📊 **Current System Metrics**

🕐 **Timestamp:** %s
🔥 **CPU Usage:** %.1f%%
💾 **Memory Usage:** %.1f%%
💿 **Disk Usage:** %.1f%%
⚖️ **Load Average:** %.2f, %.2f, %.2f

**Status:**
%s %s
%s %s
%s %s`,
		data.Timestamp.Format("2006-01-02 15:04:05"),
		data.CPUUsage,
		data.MemoryUsage,
		data.DiskUsage,
		data.LoadAverage[0], data.LoadAverage[1], data.LoadAverage[2],
		mm.getStatusIcon(data.CPUUsage, mm.config.Monitoring.CPUAlertThreshold),
		fmt.Sprintf("CPU: %.1f%% / %.1f%%", data.CPUUsage, mm.config.Monitoring.CPUAlertThreshold),
		mm.getStatusIcon(data.MemoryUsage, mm.config.Monitoring.MemoryAlertThreshold),
		fmt.Sprintf("Memory: %.1f%% / %.1f%%", data.MemoryUsage, mm.config.Monitoring.MemoryAlertThreshold),
		mm.getStatusIcon(data.DiskUsage, mm.config.Monitoring.DiskAlertThreshold),
		fmt.Sprintf("Disk: %.1f%% / %.1f%%", data.DiskUsage, mm.config.Monitoring.DiskAlertThreshold))
}

// getStatusIcon returns appropriate status icon based on usage vs threshold
func (mm *MonitoringManager) getStatusIcon(usage, threshold float64) string {
	if usage > threshold {
		return "🔴"
	} else if usage > threshold*0.8 {
		return "🟡"
	}
	return "🟢"
}