#!/bin/bash

# Remote Server Admin Bot Installation Script
# This script installs the bot as a system service

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Configuration
BOT_NAME="server-admin-bot"
BOT_USER="root"  # Running as root for full system access
INSTALL_DIR="/usr/local/bin"
CONFIG_DIR="/etc/server-admin-bot"
LOG_DIR="/var/log"
SERVICE_NAME="server-admin-bot"

echo -e "${BLUE}🤖 Remote Server Admin Bot Installer${NC}"
echo -e "${BLUE}=====================================${NC}"
echo

# Check if running as root
if [[ $EUID -ne 0 ]]; then
   echo -e "${RED}❌ This script must be run as root (use sudo)${NC}"
   exit 1
fi

echo -e "${GREEN}✅ Running as root${NC}"

# Check system compatibility
echo -e "${BLUE}🔍 Checking system compatibility...${NC}"

if ! command -v systemctl &> /dev/null; then
    echo -e "${RED}❌ systemd is required but not found${NC}"
    exit 1
fi

echo -e "${GREEN}✅ systemd found${NC}"

# Detect architecture
ARCH=$(uname -m)
case $ARCH in
    x86_64)
        ARCH="amd64"
        ;;
    aarch64|arm64)
        ARCH="arm64"
        ;;
    armv7*)
        ARCH="arm"
        ;;
    *)
        echo -e "${RED}❌ Unsupported architecture: $ARCH${NC}"
        exit 1
        ;;
esac

echo -e "${GREEN}✅ Architecture: $ARCH${NC}"

# Stop existing service if running
if systemctl is-active --quiet $SERVICE_NAME; then
    echo -e "${YELLOW}⏹️  Stopping existing service...${NC}"
    systemctl stop $SERVICE_NAME
fi

# Create directories
echo -e "${BLUE}📁 Creating directories...${NC}"
mkdir -p $CONFIG_DIR
mkdir -p $LOG_DIR
echo -e "${GREEN}✅ Directories created${NC}"

# Check if binary exists
if [[ ! -f "./$BOT_NAME" ]]; then
    echo -e "${RED}❌ Bot binary '$BOT_NAME' not found in current directory${NC}"
    echo -e "${YELLOW}Please make sure you have the compiled binary in the current directory${NC}"
    exit 1
fi

# Install binary
echo -e "${BLUE}📦 Installing bot binary...${NC}"
cp "./$BOT_NAME" "$INSTALL_DIR/$BOT_NAME"
chmod +x "$INSTALL_DIR/$BOT_NAME"
echo -e "${GREEN}✅ Binary installed to $INSTALL_DIR/$BOT_NAME${NC}"

# Create config file if it doesn't exist
if [[ ! -f "$CONFIG_DIR/config.yaml" ]]; then
    echo -e "${BLUE}⚙️  Creating configuration file...${NC}"
    
    if [[ -f "./configs/config.example.yaml" ]]; then
        cp "./configs/config.example.yaml" "$CONFIG_DIR/config.yaml"
        echo -e "${GREEN}✅ Configuration template copied${NC}"
    else
        # Create basic config
        cat > "$CONFIG_DIR/config.yaml" << EOF
telegram:
  bot_token: "YOUR_BOT_TOKEN_HERE"
  allowed_users:
    - 123456789  # Replace with your Telegram user ID

security:
  admin_password: "change_this_password"
  session_timeout: 3600
  max_file_size: 52428800
  rate_limit_window: 60
  rate_limit_commands: 10

server:
  name: "Server"
  timezone: "UTC"
  log_level: "info"
  working_dir: "/tmp"
  log_file: "$LOG_DIR/server-admin-bot.log"

monitoring:
  cpu_alert_threshold: 85.0
  memory_alert_threshold: 90.0
  disk_alert_threshold: 95.0
  monitoring_interval: 300
  alert_cooldown: 1800
EOF
        echo -e "${GREEN}✅ Basic configuration created${NC}"
    fi
    
    chmod 600 "$CONFIG_DIR/config.yaml"
    echo -e "${YELLOW}⚠️  Please edit $CONFIG_DIR/config.yaml with your bot token and settings${NC}"
else
    echo -e "${GREEN}✅ Configuration file already exists${NC}"
fi

# Create systemd service
echo -e "${BLUE}🔧 Creating systemd service...${NC}"
cat > "/etc/systemd/system/$SERVICE_NAME.service" << EOF
[Unit]
Description=Remote Server Admin Bot
Documentation=https://github.com/RusEu/RadminTelegramBot
After=network.target network-online.target
Wants=network-online.target
Requires=network.target

[Service]
Type=simple
User=$BOT_USER
Group=$BOT_USER
WorkingDirectory=$CONFIG_DIR
ExecStart=$INSTALL_DIR/$BOT_NAME -config $CONFIG_DIR/config.yaml
ExecReload=/bin/kill -HUP \$MAINPID
Restart=always
RestartSec=5
TimeoutStopSec=30
KillMode=process
KillSignal=SIGTERM

# Security settings
NoNewPrivileges=false
PrivateTmp=true
ProtectSystem=false
ProtectHome=false

# Resource limits
LimitNOFILE=65536

# Logging
StandardOutput=journal
StandardError=journal
SyslogIdentifier=$SERVICE_NAME

[Install]
WantedBy=multi-user.target
EOF

echo -e "${GREEN}✅ Systemd service created${NC}"

# Create log rotation
echo -e "${BLUE}📋 Setting up log rotation...${NC}"
cat > "/etc/logrotate.d/$SERVICE_NAME" << EOF
$LOG_DIR/server-admin-bot.log {
    daily
    missingok
    rotate 30
    compress
    delaycompress
    notifempty
    sharedscripts
    postrotate
        systemctl reload $SERVICE_NAME > /dev/null 2>&1 || true
    endscript
}
EOF

echo -e "${GREEN}✅ Log rotation configured${NC}"

# Reload systemd
echo -e "${BLUE}🔄 Reloading systemd...${NC}"
systemctl daemon-reload
echo -e "${GREEN}✅ Systemd reloaded${NC}"

# Enable service
echo -e "${BLUE}🔗 Enabling service...${NC}"
systemctl enable $SERVICE_NAME
echo -e "${GREEN}✅ Service enabled${NC}"

echo
echo -e "${GREEN}🎉 Installation completed successfully!${NC}"
echo
echo -e "${BLUE}📋 Next steps:${NC}"
echo -e "1. Edit the configuration file: ${YELLOW}$CONFIG_DIR/config.yaml${NC}"
echo -e "2. Add your bot token and user IDs"
echo -e "3. Start the service: ${YELLOW}sudo systemctl start $SERVICE_NAME${NC}"
echo -e "4. Check status: ${YELLOW}sudo systemctl status $SERVICE_NAME${NC}"
echo -e "5. View logs: ${YELLOW}sudo journalctl -u $SERVICE_NAME -f${NC}"
echo
echo -e "${BLUE}📖 Configuration file location:${NC} $CONFIG_DIR/config.yaml"
echo -e "${BLUE}📊 Log file location:${NC} $LOG_DIR/server-admin-bot.log"
echo -e "${BLUE}🔧 Service management:${NC}"
echo -e "   Start:   sudo systemctl start $SERVICE_NAME"
echo -e "   Stop:    sudo systemctl stop $SERVICE_NAME"
echo -e "   Restart: sudo systemctl restart $SERVICE_NAME"
echo -e "   Status:  sudo systemctl status $SERVICE_NAME"
echo
echo -e "${GREEN}🤖 Ready to control your server via Telegram!${NC}"
