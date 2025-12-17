# Remote Server Admin Bot

🤖 A lightweight Telegram bot for remote server administration. Install it on any server and control it remotely through Telegram.

## Features

- 🖥️ **System Information**: Get CPU, memory, disk usage, and system stats
- 📁 **File Management**: Browse directories, view files, download/upload files
- 🔧 **Process Management**: List, start, stop, and monitor processes
- 📊 **System Monitoring**: Real-time system metrics and alerts
- 🔐 **Secure Access**: User authentication and command authorization
- 📝 **Command Execution**: Run shell commands with output capture
- 🚨 **Service Management**: Start/stop/restart system services
- 📈 **Resource Monitoring**: Track system resources over time
- 🔄 **Auto-updates**: Keep the bot updated with latest features
- 📱 **Mobile-friendly**: Rich inline keyboards for easy mobile use

## Quick Start

### 1. Download and Install

```bash
# Download the latest release
wget https://github.com/yourusername/remote-server-admin-bot/releases/latest/download/server-admin-bot-linux-amd64.tar.gz

# Extract
tar -xzf server-admin-bot-linux-amd64.tar.gz

# Move to system path
sudo mv server-admin-bot /usr/local/bin/

# Make executable
sudo chmod +x /usr/local/bin/server-admin-bot
```

### 2. Configure

```bash
# Create config directory
sudo mkdir -p /etc/server-admin-bot

# Create configuration file
sudo nano /etc/server-admin-bot/config.yaml
```

Add your configuration:

```yaml
telegram:
  bot_token: "YOUR_BOT_TOKEN_HERE"
  allowed_users:
    - 123456789  # Your Telegram user ID
    - 987654321  # Additional user ID

security:
  admin_password: "your_secure_admin_password"
  session_timeout: 3600  # 1 hour
  max_file_size: 52428800  # 50MB

server:
  name: "Production Server 1"
  timezone: "UTC"
  log_level: "info"

monitoring:
  cpu_alert_threshold: 85
  memory_alert_threshold: 90
  disk_alert_threshold: 95
```

### 3. Create System Service

```bash
# Create service file
sudo nano /etc/systemd/system/server-admin-bot.service
```

Add service configuration:

```ini
[Unit]
Description=Remote Server Admin Bot
After=network.target

[Service]
Type=simple
User=root
WorkingDirectory=/etc/server-admin-bot
ExecStart=/usr/local/bin/server-admin-bot -config /etc/server-admin-bot/config.yaml
Restart=always
RestartSec=5

[Install]
WantedBy=multi-user.target
```

### 4. Start the Service

```bash
# Reload systemd
sudo systemctl daemon-reload

# Enable and start the service
sudo systemctl enable server-admin-bot
sudo systemctl start server-admin-bot

# Check status
sudo systemctl status server-admin-bot
```

## Getting Your Bot Token

1. Message [@BotFather](https://t.me/BotFather) on Telegram
2. Send `/newbot` command
3. Choose a name and username for your bot
4. Copy the bot token from BotFather's response
5. Add the token to your `config.yaml`

## Getting Your User ID

1. Message [@userinfobot](https://t.me/userinfobot) on Telegram
2. The bot will reply with your user ID
3. Add your user ID to the `allowed_users` list in `config.yaml`

## Available Commands

### System Information
- `/start` - Initialize the bot and show main menu
- `/status` - Show system status overview
- `/info` - Detailed system information
- `/uptime` - System uptime and load
- `/resources` - CPU, Memory, Disk usage

### Process Management
- `/processes` - List running processes
- `/kill <pid>` - Kill a process by ID
- `/services` - Manage system services
- `/top` - Show top processes by resource usage

### File Management
- `/ls [path]` - List directory contents
- `/pwd` - Show current directory
- `/cat <file>` - Display file contents
- `/download <file>` - Download file from server
- `/upload` - Upload file to server
- `/df` - Show disk usage

### System Control
- `/reboot` - Reboot the server (requires confirmation)
- `/shutdown` - Shutdown the server (requires confirmation)
- `/logs [lines]` - Show system logs
- `/exec <command>` - Execute shell command

### Monitoring
- `/alerts` - Show active alerts
- `/monitor` - Start/stop real-time monitoring
- `/stats` - Show performance statistics
- `/network` - Network interface information

## Configuration Options

### Telegram Settings
- `bot_token`: Your Telegram bot token
- `allowed_users`: List of authorized user IDs

### Security Settings
- `admin_password`: Password for sensitive operations
- `session_timeout`: Session timeout in seconds
- `max_file_size`: Maximum file upload size in bytes

### Server Settings
- `name`: Server name for identification
- `timezone`: Server timezone
- `log_level`: Logging level (debug, info, warn, error)

### Monitoring Settings
- `cpu_alert_threshold`: CPU usage alert threshold (%)
- `memory_alert_threshold`: Memory usage alert threshold (%)
- `disk_alert_threshold`: Disk usage alert threshold (%)

## Security Features

- ✅ **User Authentication**: Only authorized users can access the bot
- ✅ **Command Authorization**: Sensitive commands require additional confirmation
- ✅ **Session Management**: Automatic session timeouts
- ✅ **Input Validation**: All inputs are validated and sanitized
- ✅ **Rate Limiting**: Prevents command spam and abuse
- ✅ **Audit Logging**: All commands and actions are logged

## Building from Source

### Prerequisites
- Go 1.21 or later
- Git

### Clone and Build

```bash
# Clone the repository
git clone https://github.com/yourusername/remote-server-admin-bot.git
cd remote-server-admin-bot

# Install dependencies
go mod tidy

# Build for current platform
go build -o server-admin-bot cmd/bot/main.go

# Build for multiple platforms
make build-all
```

### Cross-compilation

```bash
# Linux AMD64
GOOS=linux GOARCH=amd64 go build -o server-admin-bot-linux-amd64 cmd/bot/main.go

# Linux ARM64
GOOS=linux GOARCH=arm64 go build -o server-admin-bot-linux-arm64 cmd/bot/main.go

# Windows AMD64
GOOS=windows GOARCH=amd64 go build -o server-admin-bot-windows-amd64.exe cmd/bot/main.go
```

## Logs and Troubleshooting

### View Logs
```bash
# Service logs
sudo journalctl -u server-admin-bot -f

# Log file (if configured)
tail -f /var/log/server-admin-bot.log
```

### Common Issues

**Bot not responding:**
- Check if the service is running: `systemctl status server-admin-bot`
- Verify bot token is correct
- Ensure your user ID is in the allowed_users list

**Permission errors:**
- Run bot as root for full system access
- Check file permissions on config and executable

**Network issues:**
- Verify internet connectivity
- Check firewall rules for outbound HTTPS (443)

## Development

### Project Structure
```
.
├── cmd/
│   └── bot/
│       └── main.go          # Application entry point
├── internal/
│   ├── bot/
│   │   ├── bot.go          # Main bot logic
│   │   ├── handlers.go     # Command handlers
│   │   └── middleware.go   # Authentication middleware
│   ├── config/
│   │   └── config.go       # Configuration management
│   ├── system/
│   │   ├── info.go         # System information
│   │   ├── process.go      # Process management
│   │   ├── files.go        # File operations
│   │   └── monitoring.go   # System monitoring
│   ├── security/
│   │   ├── auth.go         # Authentication
│   │   └── validation.go   # Input validation
│   └── utils/
│       ├── logger.go       # Logging utilities
│       └── helpers.go      # Helper functions
├── configs/
│   └── config.example.yaml # Example configuration
├── scripts/
│   ├── install.sh          # Installation script
│   └── uninstall.sh        # Uninstallation script
├── go.mod
├── go.sum
├── Makefile
└── README.md
```

### Contributing

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Add tests
5. Submit a pull request

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## Support

If you encounter any issues or have questions:

1. Check the [Issues](https://github.com/yourusername/remote-server-admin-bot/issues) page
2. Create a new issue with detailed information
3. Join our [Telegram group](https://t.me/your_support_group) for community support

## Changelog

See [CHANGELOG.md](CHANGELOG.md) for a list of changes in each version.

---

⚡ **Made with Go** - Fast, reliable, and lightweight server administration at your fingertips!