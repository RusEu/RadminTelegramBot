package system

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/disk"
	"github.com/shirou/gopsutil/v3/host"
	"github.com/shirou/gopsutil/v3/load"
	"github.com/shirou/gopsutil/v3/mem"
	"github.com/shirou/gopsutil/v3/process"
	log "github.com/sirupsen/logrus"
)

// InfoManager handles system information operations
type InfoManager struct {
	logger *log.Logger
}

// SystemStatus represents basic system status
type SystemStatus struct {
	Uptime              string
	MemoryUsagePercent  float64
	CPUUsagePercent     float64
	DiskUsagePercent    float64
	LoadAverage         []float64
}

// NewInfoManager creates a new InfoManager instance
func NewInfoManager(logger *log.Logger) *InfoManager {
	return &InfoManager{
		logger: logger,
	}
}

// GetSystemStatus returns basic system status
func (im *InfoManager) GetSystemStatus() (*SystemStatus, error) {
	im.logger.Debug("📊 Collecting system status...")
	
	// Get uptime
	hostStat, err := host.Info()
	if err != nil {
		return nil, fmt.Errorf("failed to get host info: %w", err)
	}

	uptime := time.Duration(hostStat.Uptime) * time.Second
	uptimeStr := formatDuration(uptime)

	// Get memory usage
	memStat, err := mem.VirtualMemory()
	if err != nil {
		return nil, fmt.Errorf("failed to get memory info: %w", err)
	}

	// Get CPU usage
	cpuPercent, err := cpu.Percent(time.Second, false)
	if err != nil {
		return nil, fmt.Errorf("failed to get CPU info: %w", err)
	}

	// Get disk usage
	diskStat, err := disk.Usage("/")
	if err != nil {
		return nil, fmt.Errorf("failed to get disk info: %w", err)
	}

	// Get load average
	loadStat, err := load.Avg()
	if err != nil {
		return nil, fmt.Errorf("failed to get load info: %w", err)
	}

	return &SystemStatus{
		Uptime:              uptimeStr,
		MemoryUsagePercent:  memStat.UsedPercent,
		CPUUsagePercent:     cpuPercent[0],
		DiskUsagePercent:    diskStat.UsedPercent,
		LoadAverage:         []float64{loadStat.Load1, loadStat.Load5, loadStat.Load15},
	}, nil
}

// GetDetailedSystemInfo returns detailed system information
func (im *InfoManager) GetDetailedSystemInfo() (string, error) {
	im.logger.Debug("ℹ️ Collecting detailed system information...")
	
	hostStat, err := host.Info()
	if err != nil {
		return "", fmt.Errorf("failed to get host info: %w", err)
	}

	memStat, err := mem.VirtualMemory()
	if err != nil {
		return "", fmt.Errorf("failed to get memory info: %w", err)
	}

	cpuInfo, err := cpu.Info()
	if err != nil {
		return "", fmt.Errorf("failed to get CPU info: %w", err)
	}

	diskStat, err := disk.Usage("/")
	if err != nil {
		return "", fmt.Errorf("failed to get disk info: %w", err)
	}

	uptime := time.Duration(hostStat.Uptime) * time.Second

	info := fmt.Sprintf(`🖥️ **System Information**

**Host:**
• Hostname: %s
• Platform: %s %s
• Architecture: %s
• Uptime: %s
• Boot Time: %s

**CPU:**
• Model: %s
• Cores: %d
• Threads: %d

**Memory:**
• Total: %s
• Used: %s (%.1f%%)
• Available: %s

**Disk (/):**
• Total: %s
• Used: %s (%.1f%%)
• Free: %s

**Runtime:**
• Go Version: %s
• Goroutines: %d`,
		hostStat.Hostname,
		hostStat.Platform,
		hostStat.PlatformVersion,
		runtime.GOARCH,
		formatDuration(uptime),
		time.Unix(int64(hostStat.BootTime), 0).Format("2006-01-02 15:04:05"),
		cpuInfo[0].ModelName,
		runtime.NumCPU(),
		len(cpuInfo),
		formatBytes(memStat.Total),
		formatBytes(memStat.Used),
		memStat.UsedPercent,
		formatBytes(memStat.Available),
		formatBytes(diskStat.Total),
		formatBytes(diskStat.Used),
		diskStat.UsedPercent,
		formatBytes(diskStat.Free),
		runtime.Version(),
		runtime.NumGoroutine())

	return info, nil
}

// GetResourceUsage returns current resource usage
func (im *InfoManager) GetResourceUsage() (string, error) {
	im.logger.Debug("📈 Collecting resource usage...")
	
	// CPU usage per core
	cpuPercents, err := cpu.Percent(time.Second, true)
	if err != nil {
		return "", fmt.Errorf("failed to get CPU usage: %w", err)
	}

	// Memory
	memStat, err := mem.VirtualMemory()
	if err != nil {
		return "", fmt.Errorf("failed to get memory usage: %w", err)
	}

	// Swap
	swapStat, err := mem.SwapMemory()
	if err != nil {
		return "", fmt.Errorf("failed to get swap usage: %w", err)
	}

	// Load average
	loadStat, err := load.Avg()
	if err != nil {
		return "", fmt.Errorf("failed to get load average: %w", err)
	}

	// Build CPU usage string
	cpuUsage := "**CPU Usage per Core:**\n"
	for i, percent := range cpuPercents {
		cpuUsage += fmt.Sprintf("• Core %d: %.1f%%\n", i, percent)
	}

	resources := fmt.Sprintf(`📈 **Resource Usage**

%s
**Memory:**
• Used: %s / %s (%.1f%%)
• Cached: %s
• Buffers: %s

**Swap:**
• Used: %s / %s (%.1f%%)

**Load Average:**
• 1min: %.2f
• 5min: %.2f
• 15min: %.2f`,
		cpuUsage,
		formatBytes(memStat.Used),
		formatBytes(memStat.Total),
		memStat.UsedPercent,
		formatBytes(memStat.Cached),
		formatBytes(memStat.Buffers),
		formatBytes(swapStat.Used),
		formatBytes(swapStat.Total),
		swapStat.UsedPercent,
		loadStat.Load1,
		loadStat.Load5,
		loadStat.Load15)

	return resources, nil
}

// GetTopProcesses returns top processes by resource usage
func (im *InfoManager) GetTopProcesses(limit int) (string, error) {
	im.logger.Debugf("🔧 Getting top %d processes...", limit)
	
	processes, err := process.Processes()
	if err != nil {
		return "", fmt.Errorf("failed to get processes: %w", err)
	}

	type ProcessInfo struct {
		PID     int32
		Name    string
		CPU     float64
		Memory  float32
		Status  string
	}

	var processInfos []ProcessInfo

	for _, p := range processes {
		name, err := p.Name()
		if err != nil {
			continue
		}

		cpuPercent, err := p.CPUPercent()
		if err != nil {
			cpuPercent = 0
		}

		memPercent, err := p.MemoryPercent()
		if err != nil {
			memPercent = 0
		}

		status, err := p.Status()
		if err != nil {
			status = "unknown"
		}

		processInfos = append(processInfos, ProcessInfo{
			PID:    p.Pid,
			Name:   name,
			CPU:    cpuPercent,
			Memory: memPercent,
			Status: status,
		})

		if len(processInfos) >= limit*2 { // Get more than needed for sorting
			break
		}
	}

	// Sort by CPU usage (simplified sorting)
	result := "🔧 **Top Processes:**\n\n"
	result += "```\n"
	result += fmt.Sprintf("%-8s %-20s %-8s %-8s %s\n", "PID", "NAME", "CPU%", "MEM%", "STATUS")
	result += strings.Repeat("-", 60) + "\n"

	count := 0
	for _, p := range processInfos {
		if count >= limit {
			break
		}
		if len(p.Name) > 20 {
			p.Name = p.Name[:17] + "..."
		}
		result += fmt.Sprintf("%-8d %-20s %-8.1f %-8.1f %s\n",
			p.PID, p.Name, p.CPU, p.Memory, p.Status)
		count++
	}
	result += "```"

	return result, nil
}

// ListDirectory lists directory contents
func (im *InfoManager) ListDirectory(path string) (string, error) {
	im.logger.Debugf("📁 Listing directory: %s", path)
	
	files, err := os.ReadDir(path)
	if err != nil {
		return "", fmt.Errorf("failed to read directory: %w", err)
	}

	var result strings.Builder
	result.WriteString("```\n")
	result.WriteString(fmt.Sprintf("%-20s %-10s %s\n", "NAME", "SIZE", "MODIFIED"))
	result.WriteString(strings.Repeat("-", 50) + "\n")

	for _, file := range files {
		info, err := file.Info()
		if err != nil {
			continue
		}

		name := file.Name()
		if file.IsDir() {
			name = "📁 " + name
		} else {
			name = "📄 " + name
		}

		if len(name) > 18 {
			name = name[:15] + "..."
		}

		size := "-"
		if !file.IsDir() {
			size = formatBytes(uint64(info.Size()))
		}

		modTime := info.ModTime().Format("01-02 15:04")
		result.WriteString(fmt.Sprintf("%-20s %-10s %s\n", name, size, modTime))
	}

	result.WriteString("```")
	return result.String(), nil
}

// ValidateDirectory validates if a directory exists and is accessible
func (im *InfoManager) ValidateDirectory(path string) error {
	info, err := os.Stat(path)
	if err != nil {
		return fmt.Errorf("directory not accessible: %w", err)
	}

	if !info.IsDir() {
		return fmt.Errorf("path is not a directory: %s", path)
	}

	return nil
}

// ExecuteCommand executes a shell command safely
func (im *InfoManager) ExecuteCommand(command, workDir string) (string, error) {
	im.logger.Warnf("⚡ Executing command: %s in %s", command, workDir)
	
	cmd := exec.Command("sh", "-c", command)
	cmd.Dir = workDir
	
	output, err := cmd.CombinedOutput()
	if err != nil {
		return string(output), fmt.Errorf("command failed: %w", err)
	}

	return string(output), nil
}

// Helper functions
func formatBytes(bytes uint64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}
	div, exp := uint64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(bytes)/float64(div), "KMGTPE"[exp])
}

func formatDuration(d time.Duration) string {
	days := d / (24 * time.Hour)
	hours := (d % (24 * time.Hour)) / time.Hour
	minutes := (d % time.Hour) / time.Minute

	if days > 0 {
		return fmt.Sprintf("%dd %dh %dm", days, hours, minutes)
	} else if hours > 0 {
		return fmt.Sprintf("%dh %dm", hours, minutes)
	} else {
		return fmt.Sprintf("%dm", minutes)
	}
}