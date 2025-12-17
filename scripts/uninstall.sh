#!/bin/bash

# Remote Server Admin Bot Uninstallation Script
# This script removes the bot and all its components

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Configuration
BOT_NAME="server-admin-bot"
INSTALL_DIR="/usr/local/bin"
CONFIG_DIR="/etc/server-admin-bot"
LOG_DIR="/var/log"
SERVICE_NAME="server-admin-bot"
LOG_FILE="$LOG_DIR/server-admin-bot.log"

echo -e "${RED}🗑️  Remote Server Admin Bot Uninstaller${NC}"
echo -e "${RED}=======================================${NC}"
echo

# Check if running as root
if [[ $EUID -ne 0 ]]; then
   echo -e "${RED}❌ This script must be run as root (use sudo)${NC}"
   exit 1
fi

echo -e "${GREEN}✅ Running as root${NC}"

# Ask for confirmation
echo -e "${YELLOW}⚠️  This will completely remove the Remote Server Admin Bot${NC}"
echo -e "${YELLOW}⚠️  including all configuration files and logs.${NC}"
echo
read -p "Are you sure you want to continue? (yes/no): " -r
echo

if [[ ! $REPLY =~ ^[Yy][Ee][Ss]$ ]]; then
    echo -e "${BLUE}ℹ️  Uninstallation cancelled${NC}"
    exit 0
fi

echo -e "${BLUE}🔄 Starting uninstallation...${NC}"
echo

# Stop and disable service
if systemctl is-active --quiet $SERVICE_NAME; then
    echo -e "${YELLOW}⏹️  Stopping service...${NC}"
    systemctl stop $SERVICE_NAME
    echo -e "${GREEN}✅ Service stopped${NC}"
else
    echo -e "${BLUE}ℹ️  Service is not running${NC}"
fi

if systemctl is-enabled --quiet $SERVICE_NAME 2>/dev/null; then
    echo -e "${YELLOW}🔗 Disabling service...${NC}"
    systemctl disable $SERVICE_NAME
    echo -e "${GREEN}✅ Service disabled${NC}"
else
    echo -e "${BLUE}ℹ️  Service is not enabled${NC}"
fi

# Remove systemd service file
if [[ -f "/etc/systemd/system/$SERVICE_NAME.service" ]]; then
    echo -e "${YELLOW}🗑️  Removing systemd service file...${NC}"
    rm -f "/etc/systemd/system/$SERVICE_NAME.service"
    echo -e "${GREEN}✅ Service file removed${NC}"
else
    echo -e "${BLUE}ℹ️  Service file not found${NC}"
fi

# Reload systemd
echo -e "${BLUE}🔄 Reloading systemd...${NC}"
systemctl daemon-reload
systemctl reset-failed 2>/dev/null || true
echo -e "${GREEN}✅ Systemd reloaded${NC}"

# Remove binary
if [[ -f "$INSTALL_DIR/$BOT_NAME" ]]; then
    echo -e "${YELLOW}🗑️  Removing bot binary...${NC}"
    rm -f "$INSTALL_DIR/$BOT_NAME"
    echo -e "${GREEN}✅ Binary removed${NC}"
else
    echo -e "${BLUE}ℹ️  Binary not found${NC}"
fi

# Ask about configuration files
if [[ -d "$CONFIG_DIR" ]]; then
    echo
    echo -e "${YELLOW}📁 Configuration directory found: $CONFIG_DIR${NC}"
    read -p "Do you want to remove configuration files? (yes/no): " -r
    echo
    
    if [[ $REPLY =~ ^[Yy][Ee][Ss]$ ]]; then
        echo -e "${YELLOW}🗑️  Removing configuration directory...${NC}"
        rm -rf "$CONFIG_DIR"
        echo -e "${GREEN}✅ Configuration directory removed${NC}"
    else
        echo -e "${BLUE}ℹ️  Configuration files preserved${NC}"
    fi
else
    echo -e "${BLUE}ℹ️  Configuration directory not found${NC}"
fi

# Ask about log files
if [[ -f "$LOG_FILE" ]]; then
    echo
    echo -e "${YELLOW}📋 Log file found: $LOG_FILE${NC}"
    read -p "Do you want to remove log files? (yes/no): " -r
    echo
    
    if [[ $REPLY =~ ^[Yy][Ee][Ss]$ ]]; then
        echo -e "${YELLOW}🗑️  Removing log files...${NC}"
        rm -f "$LOG_FILE"*
        echo -e "${GREEN}✅ Log files removed${NC}"
    else
        echo -e "${BLUE}ℹ️  Log files preserved${NC}"
    fi
else
    echo -e "${BLUE}ℹ️  Log file not found${NC}"
fi

# Remove logrotate configuration
if [[ -f "/etc/logrotate.d/$SERVICE_NAME" ]]; then
    echo -e "${YELLOW}🗑️  Removing log rotation configuration...${NC}"
    rm -f "/etc/logrotate.d/$SERVICE_NAME"
    echo -e "${GREEN}✅ Log rotation configuration removed${NC}"
else
    echo -e "${BLUE}ℹ️  Log rotation configuration not found${NC}"
fi

# Clean up any remaining processes
echo -e "${BLUE}🔍 Checking for running processes...${NC}"
if pgrep -x "$BOT_NAME" > /dev/null; then
    echo -e "${YELLOW}⚠️  Found running bot processes, terminating...${NC}"
    pkill -x "$BOT_NAME" || true
    sleep 2
    pkill -9 -x "$BOT_NAME" 2>/dev/null || true
    echo -e "${GREEN}✅ Processes terminated${NC}"
else
    echo -e "${BLUE}ℹ️  No running processes found${NC}"
fi

echo
echo -e "${GREEN}🎉 Uninstallation completed successfully!${NC}"
echo
echo -e "${BLUE}📋 Summary:${NC}"
echo -e "${GREEN}✅ Service stopped and disabled${NC}"
echo -e "${GREEN}✅ Binary removed${NC}"
echo -e "${GREEN}✅ Systemd service file removed${NC}"
echo -e "${GREEN}✅ Log rotation configuration removed${NC}"

if [[ ! -d "$CONFIG_DIR" ]]; then
    echo -e "${GREEN}✅ Configuration directory removed${NC}"
else
    echo -e "${YELLOW}ℹ️  Configuration directory preserved${NC}"
fi

if [[ ! -f "$LOG_FILE" ]]; then
    echo -e "${GREEN}✅ Log files removed${NC}"
else
    echo -e "${YELLOW}ℹ️  Log files preserved${NC}"
fi

echo
echo -e "${BLUE}👋 Thank you for using Remote Server Admin Bot!${NC}"
