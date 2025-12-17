package bot

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// handleMessage handles incoming text messages
func (b *Bot) handleMessage(message *tgbotapi.Message) {
	session := b.getOrCreateSession(message.From.ID, message.From.UserName)
	chatID := message.Chat.ID
	text := message.Text

	b.logger.Infof("🎯 Processing command: %s from @%s", text, session.Username)

	// Handle file upload
	if message.Document != nil {
		b.handleFileUpload(chatID, message.Document)
		return
	}

	// Handle commands
	switch {
	case strings.HasPrefix(text, "/start"):
		b.handleStart(chatID, session)
	case strings.HasPrefix(text, "/help"):
		b.handleHelp(chatID)
	case strings.HasPrefix(text, "/status"):
		b.handleStatus(chatID)
	case strings.HasPrefix(text, "/info"):
		b.handleSystemInfo(chatID)
	case strings.HasPrefix(text, "/resources"):
		b.handleResources(chatID)
	case strings.HasPrefix(text, "/processes"):
		b.handleProcesses(chatID)
	case strings.HasPrefix(text, "/services"):
		b.handleServices(chatID)
	case strings.HasPrefix(text, "/ls"):
		b.handleListDirectory(chatID, session, text)
	case strings.HasPrefix(text, "/pwd"):
		b.handlePrintWorkingDirectory(chatID, session)
	case strings.HasPrefix(text, "/cd"):
		b.handleChangeDirectory(chatID, session, text)
	case strings.HasPrefix(text, "/cat"):
		b.handleCatFile(chatID, text)
	case strings.HasPrefix(text, "/download"):
		b.handleDownload(chatID, text)
	case strings.HasPrefix(text, "/exec"):
		b.handleExecCommand(chatID, session, text)
	case strings.HasPrefix(text, "/kill"):
		b.handleKillProcess(chatID, text)
	case strings.HasPrefix(text, "/logs"):
		b.handleLogs(chatID, text)
	case strings.HasPrefix(text, "/admin"):
		b.handleAdmin(chatID, session, text)
	default:
		b.sendMessage(chatID, "❓ Unknown command. Type /help for available commands.")
	}
}

// handleStart handles the /start command
func (b *Bot) handleStart(chatID int64, session *UserSession) {
	b.logger.Infof("👋 User @%s started the bot", session.Username)
	
	welcomeMsg := fmt.Sprintf(
		"🤖 **Remote Server Admin Bot**\n\n"+
			"Welcome @%s! You are now connected to:\n"+
			"🏷️ Server: **%s**\n"+
			"🕐 Time: %s\n\n"+
			"Choose an option below to get started:",
		session.Username,
		b.config.Server.Name,
		time.Now().Format("2006-01-02 15:04:05"),
	)

	msg := tgbotapi.NewMessage(chatID, welcomeMsg)
	msg.ParseMode = "Markdown"
	msg.ReplyMarkup = b.createMainKeyboard()
	b.api.Send(msg)
}

// handleHelp handles the /help command
func (b *Bot) handleHelp(chatID int64) {
	b.logger.Info("📖 Help command requested")
	
	helpText := `🤖 **Remote Server Admin Bot - Commands**

**System Information:**
/start - Show main menu
/status - System status overview
/info - Detailed system information
/resources - CPU, Memory, Disk usage

**Process Management:**
/processes - List running processes
/kill <pid> - Kill process by ID
/services - System services status

**File Management:**
/ls [path] - List directory contents
/pwd - Show current directory
/cd <path> - Change directory
/cat <file> - Display file contents
/download <file> - Download file

**System Control:**
/exec <command> - Execute shell command
/logs [lines] - Show system logs
/admin - Admin functions

**Tips:**
• Use inline buttons for easier navigation
• File paths can be relative or absolute
• Commands are logged for security
• Session timeout: ` + fmt.Sprintf("%d minutes", b.config.Security.SessionTimeout/60)

	msg := tgbotapi.NewMessage(chatID, helpText)
	msg.ParseMode = "Markdown"
	b.api.Send(msg)
}

// handleStatus handles the /status command
func (b *Bot) handleStatus(chatID int64) {
	b.logger.Info("📊 Status command requested")
	
	info, err := b.sysInfo.GetSystemStatus()
	if err != nil {
		b.logger.Errorf("❌ Failed to get system status: %v", err)
		b.sendMessage(chatID, fmt.Sprintf("❌ Error getting system status: %v", err))
		return
	}

	statusMsg := fmt.Sprintf(
		"📊 **System Status - %s**\n\n"+
			"🖥️ **Uptime:** %s\n"+
			"💾 **Memory:** %.1f%% used\n"+
			"🔥 **CPU:** %.1f%% used\n"+
			"💿 **Disk:** %.1f%% used\n"+
			"🌐 **Load:** %.2f, %.2f, %.2f\n"+
			"🕐 **Time:** %s",
		b.config.Server.Name,
		info.Uptime,
		info.MemoryUsagePercent,
		info.CPUUsagePercent,
		info.DiskUsagePercent,
		info.LoadAverage[0], info.LoadAverage[1], info.LoadAverage[2],
		time.Now().Format("2006-01-02 15:04:05"),
	)

	msg := tgbotapi.NewMessage(chatID, statusMsg)
	msg.ParseMode = "Markdown"
	b.api.Send(msg)
}

// handleSystemInfo handles the /info command
func (b *Bot) handleSystemInfo(chatID int64) {
	b.logger.Info("ℹ️ System info command requested")
	
	info, err := b.sysInfo.GetDetailedSystemInfo()
	if err != nil {
		b.logger.Errorf("❌ Failed to get system info: %v", err)
		b.sendMessage(chatID, fmt.Sprintf("❌ Error getting system info: %v", err))
		return
	}

	b.sendMessage(chatID, info)
}

// handleResources handles the /resources command
func (b *Bot) handleResources(chatID int64) {
	b.logger.Info("📈 Resources command requested")
	
	resources, err := b.sysInfo.GetResourceUsage()
	if err != nil {
		b.logger.Errorf("❌ Failed to get resources: %v", err)
		b.sendMessage(chatID, fmt.Sprintf("❌ Error getting resources: %v", err))
		return
	}

	b.sendMessage(chatID, resources)
}

// handleProcesses handles the /processes command
func (b *Bot) handleProcesses(chatID int64) {
	b.logger.Info("🔧 Processes command requested")
	
	processes, err := b.sysInfo.GetTopProcesses(10)
	if err != nil {
		b.logger.Errorf("❌ Failed to get processes: %v", err)
		b.sendMessage(chatID, fmt.Sprintf("❌ Error getting processes: %v", err))
		return
	}

	b.sendMessage(chatID, processes)
}

// handleServices handles the /services command
func (b *Bot) handleServices(chatID int64) {
	b.logger.Info("🚀 Services command requested")
	
	// This is a simplified version - you might want to implement proper service management
	cmd := exec.Command("systemctl", "list-units", "--type=service", "--state=running", "--no-pager")
	output, err := cmd.Output()
	if err != nil {
		b.logger.Errorf("❌ Failed to list services: %v", err)
		b.sendMessage(chatID, "❌ Error listing services")
		return
	}

	servicesText := fmt.Sprintf("🚀 **Running Services:**\n\n```\n%s\n```", string(output))
	msg := tgbotapi.NewMessage(chatID, servicesText)
	msg.ParseMode = "Markdown"
	b.api.Send(msg)
}

// handleListDirectory handles the /ls command
func (b *Bot) handleListDirectory(chatID int64, session *UserSession, text string) {
	parts := strings.Fields(text)
	var path string
	
	if len(parts) > 1 {
		path = parts[1]
		if !filepath.IsAbs(path) {
			path = filepath.Join(session.CurrentPath, path)
		}
	} else {
		path = session.CurrentPath
	}

	b.logger.Infof("📁 Listing directory: %s for @%s", path, session.Username)

	files, err := b.sysInfo.ListDirectory(path)
	if err != nil {
		b.logger.Errorf("❌ Failed to list directory %s: %v", path, err)
		b.sendMessage(chatID, fmt.Sprintf("❌ Error listing directory: %v", err))
		return
	}

	response := fmt.Sprintf("📁 **Directory: %s**\n\n%s", path, files)
	b.sendMessage(chatID, response)
}

// handlePrintWorkingDirectory handles the /pwd command
func (b *Bot) handlePrintWorkingDirectory(chatID int64, session *UserSession) {
	b.logger.Infof("📂 PWD command requested by @%s", session.Username)
	response := fmt.Sprintf("📂 **Current Directory:**\n%s", session.CurrentPath)
	b.sendMessage(chatID, response)
}

// handleChangeDirectory handles the /cd command
func (b *Bot) handleChangeDirectory(chatID int64, session *UserSession, text string) {
	parts := strings.Fields(text)
	if len(parts) < 2 {
		b.sendMessage(chatID, "❌ Usage: /cd <directory>")
		return
	}

	newPath := parts[1]
	if !filepath.IsAbs(newPath) {
		newPath = filepath.Join(session.CurrentPath, newPath)
	}

	b.logger.Infof("📂 CD command: %s -> %s for @%s", session.CurrentPath, newPath, session.Username)

	if err := b.sysInfo.ValidateDirectory(newPath); err != nil {
		b.sendMessage(chatID, fmt.Sprintf("❌ Invalid directory: %v", err))
		return
	}

	session.CurrentPath = newPath
	response := fmt.Sprintf("📂 **Changed to:**\n%s", newPath)
	b.sendMessage(chatID, response)
}

// handleExecCommand handles the /exec command
func (b *Bot) handleExecCommand(chatID int64, session *UserSession, text string) {
	parts := strings.Fields(text)
	if len(parts) < 2 {
		b.sendMessage(chatID, "❌ Usage: /exec <command>")
		return
	}

	cmdText := strings.TrimPrefix(text, "/exec ")
	b.logger.Warnf("⚡ EXEC command: '%s' by @%s", cmdText, session.Username)

	// Security validation
	if !b.auth.ValidateCommand(cmdText) {
		b.logger.Warnf("🚫 Dangerous command blocked: %s", cmdText)
		b.sendMessage(chatID, "❌ Command blocked for security reasons")
		return
	}

	output, err := b.sysInfo.ExecuteCommand(cmdText, session.CurrentPath)
	if err != nil {
		b.logger.Errorf("❌ Command execution failed: %v", err)
		b.sendMessage(chatID, fmt.Sprintf("❌ Command failed: %v", err))
		return
	}

	response := fmt.Sprintf("⚡ **Command:** `%s`\n\n```\n%s\n```", cmdText, output)
	msg := tgbotapi.NewMessage(chatID, response)
	msg.ParseMode = "Markdown"
	b.api.Send(msg)
}

// handleFileUpload handles file uploads
func (b *Bot) handleFileUpload(chatID int64, document *tgbotapi.Document) {
	b.logger.Infof("📤 File upload request: %s (%s)", document.FileName, document.FileID)
	
	// Get file
	fileURL, err := b.api.GetFileDirectURL(document.FileID)
	if err != nil {
		b.logger.Errorf("❌ Failed to get file URL: %v", err)
		b.sendMessage(chatID, "❌ Failed to download file")
		return
	}

	b.logger.Infof("✅ File uploaded: %s, URL: %s", document.FileName, fileURL)
	b.sendMessage(chatID, fmt.Sprintf("✅ File received: **%s** (%d bytes)\n\nUse /ls to see uploaded files", 
		document.FileName, document.FileSize))
}

// handleCatFile handles the /cat command to display file contents
func (b *Bot) handleCatFile(chatID int64, text string) {
	parts := strings.Fields(text)
	if len(parts) < 2 {
		b.sendMessage(chatID, "❌ Usage: /cat <file>")
		return
	}

	filePath := parts[1]
	b.logger.Infof("📄 Cat file request: %s", filePath)

	// Read file contents
	content, err := os.ReadFile(filePath)
	if err != nil {
		b.logger.Errorf("❌ Failed to read file %s: %v", filePath, err)
		b.sendMessage(chatID, fmt.Sprintf("❌ Error reading file: %v", err))
		return
	}

	// Limit content size
	maxSize := 4000
	contentStr := string(content)
	if len(contentStr) > maxSize {
		contentStr = contentStr[:maxSize] + "\n...(truncated)"
	}

	response := fmt.Sprintf("📄 **File: %s**\n\n```\n%s\n```", filePath, contentStr)
	msg := tgbotapi.NewMessage(chatID, response)
	msg.ParseMode = "Markdown"
	b.api.Send(msg)
}

// handleDownload handles the /download command to send files to user
func (b *Bot) handleDownload(chatID int64, text string) {
	parts := strings.Fields(text)
	if len(parts) < 2 {
		b.sendMessage(chatID, "❌ Usage: /download <file>")
		return
	}

	filePath := parts[1]
	b.logger.Infof("📥 Download request: %s", filePath)

	// Check if file exists
	fileInfo, err := os.Stat(filePath)
	if err != nil {
		b.logger.Errorf("❌ File not found: %s - %v", filePath, err)
		b.sendMessage(chatID, fmt.Sprintf("❌ File not found: %v", err))
		return
	}

	if fileInfo.IsDir() {
		b.sendMessage(chatID, "❌ Cannot download directories. Please specify a file.")
		return
	}

	// Send the file
	caption := fmt.Sprintf("📥 %s (%s)", filepath.Base(filePath), formatBytes(uint64(fileInfo.Size())))
	if err := b.sendDocument(chatID, filePath, caption); err != nil {
		b.logger.Errorf("❌ Failed to send file: %v", err)
		b.sendMessage(chatID, fmt.Sprintf("❌ Failed to send file: %v", err))
		return
	}

	b.logger.Infof("✅ File sent: %s", filePath)
}

// handleKillProcess handles the /kill command to terminate processes
func (b *Bot) handleKillProcess(chatID int64, text string) {
	parts := strings.Fields(text)
	if len(parts) < 2 {
		b.sendMessage(chatID, "❌ Usage: /kill <pid>")
		return
	}

	pidStr := parts[1]
	pid, err := strconv.Atoi(pidStr)
	if err != nil {
		b.sendMessage(chatID, "❌ Invalid PID. Must be a number.")
		return
	}

	b.logger.Warnf("⚠️ Kill process request: PID %d", pid)

	// Find and kill the process
	proc, err := os.FindProcess(pid)
	if err != nil {
		b.logger.Errorf("❌ Process not found: PID %d - %v", pid, err)
		b.sendMessage(chatID, fmt.Sprintf("❌ Process not found: %v", err))
		return
	}

	if err := proc.Kill(); err != nil {
		b.logger.Errorf("❌ Failed to kill process %d: %v", pid, err)
		b.sendMessage(chatID, fmt.Sprintf("❌ Failed to kill process: %v", err))
		return
	}

	b.logger.Infof("✅ Process killed: PID %d", pid)
	b.sendMessage(chatID, fmt.Sprintf("✅ Process killed: PID %d", pid))
}

// handleLogs handles the /logs command to show system logs
func (b *Bot) handleLogs(chatID int64, text string) {
	parts := strings.Fields(text)
	lines := 50 // Default number of lines
	
	if len(parts) > 1 {
		if n, err := strconv.Atoi(parts[1]); err == nil && n > 0 {
			lines = n
			if lines > 200 {
				lines = 200 // Limit to 200 lines
			}
		}
	}

	b.logger.Infof("📝 Logs request: %d lines", lines)

	// Try to read system logs
	cmd := exec.Command("journalctl", "-n", strconv.Itoa(lines), "--no-pager")
	output, err := cmd.Output()
	
	// Fallback to other log sources if journalctl is not available
	if err != nil {
		// Try reading from /var/log/syslog
		cmd = exec.Command("tail", "-n", strconv.Itoa(lines), "/var/log/syslog")
		output, err = cmd.Output()
		
		if err != nil {
			b.logger.Errorf("❌ Failed to read logs: %v", err)
			b.sendMessage(chatID, "❌ Failed to read system logs. Journalctl or syslog not accessible.")
			return
		}
	}

	logsText := string(output)
	if len(logsText) > 4000 {
		logsText = logsText[len(logsText)-4000:]
	}

	response := fmt.Sprintf("📝 **System Logs (last %d lines):**\n\n```\n%s\n```", lines, logsText)
	msg := tgbotapi.NewMessage(chatID, response)
	msg.ParseMode = "Markdown"
	b.api.Send(msg)
}

// handleAdmin handles the /admin command with subcommands
func (b *Bot) handleAdmin(chatID int64, session *UserSession, text string) {
	parts := strings.Fields(text)
	
	// If no subcommand, show admin menu
	if len(parts) < 2 {
		b.handleAdminMenu(chatID, session)
		return
	}

	subcommand := parts[1]
	b.logger.Infof("⚙️ Admin command: %s by @%s", subcommand, session.Username)

	switch subcommand {
	case "reboot":
		b.sendMessage(chatID, "⚠️ Reboot command disabled for safety. Please use your system's management console.")
	case "shutdown":
		b.sendMessage(chatID, "⚠️ Shutdown command disabled for safety. Please use your system's management console.")
	case "update":
		b.sendMessage(chatID, "🔄 System update initiated... (This is a placeholder - implement actual update logic)")
	default:
		b.sendMessage(chatID, "❓ Unknown admin command. Available: reboot, shutdown, update")
	}
}

// handleAdminMenu shows the admin menu
func (b *Bot) handleAdminMenu(chatID int64, session *UserSession) {
	b.logger.Infof("⚙️ Admin menu requested by @%s", session.Username)
	
	adminText := `⚙️ **Admin Functions**

**Available Commands:**
• /admin reboot - Reboot system
• /admin shutdown - Shutdown system
• /admin update - Update system

**System Control:**
Use these commands with caution!

**Current Session:**
• User: @` + session.Username + `
• Session started: ` + session.StartTime.Format("2006-01-02 15:04:05")

	msg := tgbotapi.NewMessage(chatID, adminText)
	msg.ParseMode = "Markdown"
	b.api.Send(msg)
}

// Helper function for formatting bytes (moved here to use in handlers)
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

// handleCallbackQuery handles inline keyboard callbacks
func (b *Bot) handleCallbackQuery(callback *tgbotapi.CallbackQuery) {
	session := b.getOrCreateSession(callback.From.ID, callback.From.UserName)
	chatID := callback.Message.Chat.ID
	data := callback.Data

	b.logger.Infof("🖱️ Callback query: %s from @%s", data, session.Username)

	// Answer callback to remove loading state
	callbackConfig := tgbotapi.NewCallback(callback.ID, "")
	b.api.Request(callbackConfig)

	switch data {
	case "system_info":
		b.handleSystemInfo(chatID)
	case "resources":
		b.handleResources(chatID)
	case "processes":
		b.handleProcesses(chatID)
	case "services":
		b.handleServices(chatID)
	case "files":
		b.handleListDirectory(chatID, session, "/ls")
	case "logs":
		b.handleLogs(chatID, "/logs")
	case "admin":
		b.handleAdminMenu(chatID, session)
	case "help":
		b.handleHelp(chatID)
	}
}