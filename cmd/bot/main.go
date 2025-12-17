package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/RusEu/RadminTelegramBot/internal/bot"
	"github.com/RusEu/RadminTelegramBot/internal/config"
	"github.com/RusEu/RadminTelegramBot/internal/utils"
	log "github.com/sirupsen/logrus"
)

var (
	version    = "1.0.0"
	buildTime  = "unknown"
	gitCommit  = "unknown"
	configPath = flag.String("config", "config.yaml", "Path to configuration file")
	showVersion = flag.Bool("version", false, "Show version information")
)

func main() {
	flag.Parse()

	// Show version and exit
	if *showVersion {
		fmt.Printf("Remote Server Admin Bot\n")
		fmt.Printf("Version: %s\n", version)
		fmt.Printf("Build Time: %s\n", buildTime)
		fmt.Printf("Git Commit: %s\n", gitCommit)
		os.Exit(0)
	}

	// Load configuration
	log.Info("🚀 Starting Remote Server Admin Bot...")
	log.Infof("📁 Loading configuration from: %s", *configPath)
	
	cfg, err := config.LoadConfig(*configPath)
	if err != nil {
		log.Fatalf("❌ Failed to load configuration: %v", err)
	}

	// Setup logging
	logger := utils.SetupLogger(cfg.Server.LogLevel)
	logger.Infof("✅ Configuration loaded successfully")
	logger.Infof("🏷️  Server Name: %s", cfg.Server.Name)
	logger.Infof("🕐 Timezone: %s", cfg.Server.Timezone)
	logger.Infof("📊 Log Level: %s", cfg.Server.LogLevel)

	// Validate configuration
	if err := cfg.Validate(); err != nil {
		logger.Fatalf("❌ Configuration validation failed: %v", err)
	}

	logger.Infof("👥 Authorized users: %d", len(cfg.Telegram.AllowedUsers))
	logger.Infof("🛡️  Security features enabled")

	// Create bot instance
	logger.Info("🤖 Initializing Telegram bot...")
	botInstance, err := bot.NewBot(cfg, logger)
	if err != nil {
		logger.Fatalf("❌ Failed to create bot: %v", err)
	}

	// Setup graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Handle shutdown signals
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	go func() {
		sig := <-sigChan
		logger.Infof("🛑 Received signal: %v", sig)
		logger.Info("🔄 Gracefully shutting down...")
		cancel()
	}()

	// Start bot
	logger.Info("🎯 Starting bot services...")
	logger.Info("✅ Remote Server Admin Bot is running!")
	logger.Info("📱 Send /start to your bot to begin")
	
	if err := botInstance.Start(ctx); err != nil {
		logger.Errorf("❌ Bot error: %v", err)
	}

	logger.Info("👋 Bot stopped gracefully")
}