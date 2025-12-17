package security

import (
	"strings"
	"sync"
	"time"

	log "github.com/sirupsen/logrus"

	"github.com/yourusername/remote-server-admin-bot/internal/config"
)

// AuthManager handles authentication and authorization
type AuthManager struct {
	config      *config.Config
	logger      *log.Logger
	rateLimits  map[int64]*RateLimit
	mu          sync.RWMutex
}

// RateLimit represents rate limiting data for a user
type RateLimit struct {
	Commands    int
	WindowStart time.Time
}

// NewAuthManager creates a new authentication manager
func NewAuthManager(cfg *config.Config, logger *log.Logger) *AuthManager {
	return &AuthManager{
		config:     cfg,
		logger:     logger,
		rateLimits: make(map[int64]*RateLimit),
	}
}

// IsAuthorized checks if a user is authorized to use the bot
func (am *AuthManager) IsAuthorized(userID int64) bool {
	for _, allowedID := range am.config.Telegram.AllowedUsers {
		if userID == allowedID {
			am.logger.Debugf("✅ User %d is authorized", userID)
			return true
		}
	}
	am.logger.Warnf("❌ User %d is NOT authorized", userID)
	return false
}

// CheckRateLimit checks if user has exceeded rate limit
func (am *AuthManager) CheckRateLimit(userID int64) bool {
	am.mu.Lock()
	defer am.mu.Unlock()

	now := time.Now()
	windowDuration := time.Duration(am.config.Security.RateLimitWindow) * time.Second

	// Get or create rate limit entry
	rl, exists := am.rateLimits[userID]
	if !exists {
		am.rateLimits[userID] = &RateLimit{
			Commands:    1,
			WindowStart: now,
		}
		am.logger.Debugf("⚡ Created new rate limit for user %d", userID)
		return true
	}

	// Reset window if expired
	if now.Sub(rl.WindowStart) > windowDuration {
		rl.Commands = 1
		rl.WindowStart = now
		am.logger.Debugf("🔄 Reset rate limit window for user %d", userID)
		return true
	}

	// Check if within limit
	if rl.Commands >= am.config.Security.RateLimitCommands {
		am.logger.Warnf("⚡ Rate limit exceeded for user %d: %d/%d commands", 
			userID, rl.Commands, am.config.Security.RateLimitCommands)
		return false
	}

	// Increment counter
	rl.Commands++
	am.logger.Debugf("📊 Rate limit for user %d: %d/%d commands", 
		userID, rl.Commands, am.config.Security.RateLimitCommands)
	return true
}

// ValidateCommand validates if a command is safe to execute
func (am *AuthManager) ValidateCommand(command string) bool {
	command = strings.TrimSpace(strings.ToLower(command))
	am.logger.Debugf("🔍 Validating command: %s", command)

	// List of dangerous commands/patterns
	dangerousCommands := []string{
		"rm -rf /",
		"rm -rf /*",
		"mkfs",
		"dd if=/dev/zero",
		"dd if=/dev/random",
		":(){ :|:& };:", // Fork bomb
		"sudo dd",
		"sudo mkfs",
		"sudo rm -rf",
		">" + " /dev/sda",
		">" + " /dev/disk",
		"fdisk",
		"parted",
		"cfdisk",
	}

	dangerousPatterns := []string{
		"rm -rf",
		"rm -fr",
		"rm -r /",
		"rm -rf ~",
		"rm -rf $HOME",
		">/dev/sd",
		">/dev/hd",
		">/dev/disk",
		"curl | sh",
		"wget | sh",
		"curl | bash",
		"wget | bash",
		"sudo passwd",
		"sudo su",
		"chmod 777 /",
		"chown -R",
	}

	// Check exact matches
	for _, dangerous := range dangerousCommands {
		if command == dangerous {
			am.logger.Warnf("🚫 Blocked dangerous command: %s", command)
			return false
		}
	}

	// Check pattern matches
	for _, pattern := range dangerousPatterns {
		if strings.Contains(command, pattern) {
			am.logger.Warnf("🚫 Blocked dangerous pattern '%s' in command: %s", pattern, command)
			return false
		}
	}

	// Additional checks for shell operators
	if strings.Contains(command, "&&") && strings.Contains(command, "rm") {
		am.logger.Warnf("🚫 Blocked potentially dangerous chained command: %s", command)
		return false
	}

	if strings.Contains(command, ";") && strings.Contains(command, "rm") {
		am.logger.Warnf("🚫 Blocked potentially dangerous sequential command: %s", command)
		return false
	}

	am.logger.Debugf("✅ Command validated successfully: %s", command)
	return true
}

// ValidateFilePath validates if a file path is safe to access
func (am *AuthManager) ValidateFilePath(path string) bool {
	am.logger.Debugf("📂 Validating file path: %s", path)

	// Normalize path
	path = strings.TrimSpace(path)

	// Blocked paths
	blockedPaths := []string{
		"/etc/shadow",
		"/etc/passwd",
		"/etc/sudoers",
		"/root/.ssh",
		"/home/*/.ssh",
		"/var/log/auth.log",
		"/etc/ssh/ssh_host_rsa_key",
		"/etc/ssl/private",
		"/proc/kcore",
		"/dev/mem",
		"/dev/kmem",
	}

	blockedPatterns := []string{
		".ssh/id_rsa",
		".ssh/id_ed25519",
		".ssh/id_ecdsa",
		"private_key",
		"*.key",
		"*.pem",
		"*.p12",
		"*.pfx",
	}

	// Check blocked paths
	for _, blocked := range blockedPaths {
		if path == blocked {
			am.logger.Warnf("🚫 Blocked access to sensitive path: %s", path)
			return false
		}
	}

	// Check blocked patterns
	for _, pattern := range blockedPatterns {
		if strings.Contains(path, pattern) {
			am.logger.Warnf("🚫 Blocked access to sensitive pattern '%s' in path: %s", pattern, path)
			return false
		}
	}

	// Check for directory traversal
	if strings.Contains(path, "../") || strings.Contains(path, "..\\`) {
		am.logger.Warnf("🚫 Blocked directory traversal attempt: %s", path)
		return false
	}

	am.logger.Debugf("✅ File path validated successfully: %s", path)
	return true
}

// GetRateLimitStatus returns current rate limit status for a user
func (am *AuthManager) GetRateLimitStatus(userID int64) (int, int, time.Duration) {
	am.mu.RLock()
	defer am.mu.RUnlock()

	rl, exists := am.rateLimits[userID]
	if !exists {
		return 0, am.config.Security.RateLimitCommands, time.Duration(am.config.Security.RateLimitWindow) * time.Second
	}

	windowDuration := time.Duration(am.config.Security.RateLimitWindow) * time.Second
	remaining := windowDuration - time.Since(rl.WindowStart)
	
	return rl.Commands, am.config.Security.RateLimitCommands, remaining
}

// ClearRateLimit clears rate limit for a user (admin function)
func (am *AuthManager) ClearRateLimit(userID int64) {
	am.mu.Lock()
	defer am.mu.Unlock()

	delete(am.rateLimits, userID)
	am.logger.Infof("🗑️ Cleared rate limit for user %d", userID)
}