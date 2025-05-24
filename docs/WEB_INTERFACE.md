# Web Interface Documentation

## Overview

The Export Trakt 4 Letterboxd web interface provides a modern, responsive web-based management system for the application. It offers real-time monitoring, configuration management, and export operations through an intuitive user interface.

## Features

### 🎯 Dashboard

- **Real-time System Health**: Live monitoring of API, database, and memory status
- **Quick Actions**: Start exports, view logs, and access configuration
- **System Metrics**: Display of key performance indicators
- **WebSocket Integration**: Real-time updates without page refresh

### ⚙️ Configuration Management

- **Trakt API Settings**: Configure client credentials and authentication
- **Export Options**: Set output directory, format, and timezone preferences
- **Logging Configuration**: Adjust log levels and output formats
- **Connection Testing**: Validate Trakt API connectivity

### 📦 Export Management

- **Export Listing**: View all available exports with metadata
- **Export Creation**: Start new exports with custom parameters
- **File Operations**: Download, delete, and manage export files
- **Status Tracking**: Monitor export progress and completion

### 📊 Monitoring & Logs

- **Health Checks**: Comprehensive system health monitoring
- **Metrics Display**: Performance and usage statistics
- **Log Viewing**: Real-time log streaming and filtering
- **System Stats**: Resource usage and application metrics

## Architecture

### Backend Components

#### Server (`pkg/webui/server.go`)

- HTTP server with Gorilla Mux router
- WebSocket support for real-time communication
- Embedded static file serving
- Graceful shutdown handling

#### Handlers (`pkg/webui/handlers/`)

- **Config Handler**: Configuration API endpoints
- **Export Handler**: Export management operations
- **Monitoring Handler**: Health and metrics endpoints
- **UI Handler**: HTML template rendering

#### Middleware (`pkg/webui/middleware/`)

- **Logging**: Request/response logging
- **CORS**: Cross-origin resource sharing
- **Security**: Security headers and CSP policies

### Frontend Components

#### Templates (`pkg/webui/templates/`)

- **Base Template**: Common layout and navigation
- **Dashboard**: Real-time monitoring interface
- **Configuration**: Settings management forms

#### Static Assets (`pkg/webui/static/`)

- **CSS**: Modern responsive styling with dark mode
- **JavaScript**: API integration and real-time features
- **Assets**: Icons and other static resources

## API Endpoints

### Configuration

- `GET /api/v1/config` - Get current configuration
- `POST /api/v1/config` - Update configuration
- `POST /api/v1/config/trakt/auth` - Trakt authentication
- `POST /api/v1/config/test` - Test API connection

### Exports

- `GET /api/v1/exports` - List all exports
- `POST /api/v1/exports` - Start new export
- `GET /api/v1/exports/{id}` - Get export details
- `DELETE /api/v1/exports/{id}` - Delete export
- `GET /api/v1/exports/{id}/download` - Download export file

### Monitoring

- `GET /api/v1/health` - System health check
- `GET /api/v1/metrics` - Application metrics
- `GET /api/v1/stats` - System statistics
- `GET /api/v1/logs` - Application logs

### Real-time

- `WS /api/v1/ws` - WebSocket connection for live updates

## Usage

### Starting the Web Server

```bash
# Build the webserver
go build -o webserver ./cmd/webserver

# Run with default settings
./webserver

# Run with custom port
./webserver -port 3000

# Run with custom config
./webserver -config /path/to/config.toml -port 8080
```

### Command Line Options

- `-config`: Path to configuration file (default: `./config/config.toml`)
- `-port`: Port to run the web server on (default: `8080`)
- `-help`: Show help message

### Environment Variables

- `EXPORT_TRAKT_CONFIG_PATH`: Path to configuration file
- `EXPORT_TRAKT_WEB_PORT`: Port for web server

### Accessing the Interface

Once started, the web interface is available at:

- **Main Interface**: `http://localhost:8080`
- **Health Check**: `http://localhost:8080/api/v1/health`
- **Metrics**: `http://localhost:8080/api/v1/metrics`

## Security Features

### HTTP Security Headers

- Content Security Policy (CSP)
- X-Frame-Options
- X-Content-Type-Options
- X-XSS-Protection
- Referrer-Policy

### CORS Support

- Configurable cross-origin resource sharing
- Development-friendly defaults
- Production-ready security options

### Input Validation

- Server-side validation for all API endpoints
- Client-side form validation
- Sanitized error messages

## Development

### Project Structure

```
pkg/webui/
├── server.go              # Main web server
├── handlers/              # HTTP handlers
│   ├── config.go         # Configuration endpoints
│   ├── export.go         # Export management
│   ├── monitoring.go     # Health and metrics
│   └── ui.go             # Template rendering
├── middleware/           # HTTP middleware
│   └── middleware.go     # Logging, CORS, security
├── static/              # Static assets
│   ├── css/
│   │   └── style.css    # Main stylesheet
│   ├── js/
│   │   └── app.js       # Frontend JavaScript
│   └── assets/          # Images and icons
└── templates/           # HTML templates
    ├── base.html        # Base layout
    ├── dashboard.html   # Dashboard page
    └── config.html      # Configuration page
```

### Adding New Features

1. **Backend**: Add new handlers in `pkg/webui/handlers/`
2. **Frontend**: Update templates and JavaScript in `static/`
3. **Routing**: Register new routes in `server.go`
4. **Middleware**: Add cross-cutting concerns in `middleware/`

### Building and Testing

```bash
# Build the webserver
go build -o webserver ./cmd/webserver

# Run tests
go test ./pkg/webui/...

# Start development server
./webserver -port 8080
```

## Troubleshooting

### Common Issues

1. **Port Already in Use**

   ```bash
   # Use a different port
   ./webserver -port 3000
   ```

2. **Configuration File Not Found**

   ```bash
   # Specify config path
   ./webserver -config /path/to/config.toml
   ```

3. **Static Files Not Loading**
   - Ensure the `static/` directory is properly embedded
   - Check file permissions and paths

### Logs and Debugging

- Check application logs for detailed error information
- Use the `/api/v1/health` endpoint to verify system status
- Monitor the `/api/v1/logs` endpoint for real-time log streaming

## Browser Compatibility

The web interface supports modern browsers with:

- ES6+ JavaScript features
- CSS Grid and Flexbox
- WebSocket connections
- Responsive design for mobile devices

### Tested Browsers

- Chrome 90+
- Firefox 88+
- Safari 14+
- Edge 90+

## Performance

### Optimization Features

- Embedded static assets for fast loading
- Efficient WebSocket connections
- Responsive design for mobile devices
- Minimal JavaScript dependencies

### Resource Usage

- Low memory footprint
- Efficient HTTP handling
- Graceful degradation under load
- Configurable timeouts and limits

## Contributing

When contributing to the web interface:

1. Follow the existing code structure and patterns
2. Add appropriate error handling and logging
3. Update documentation for new features
4. Test across different browsers and devices
5. Ensure security best practices are followed

## License

This web interface is part of the Export Trakt 4 Letterboxd project and follows the same license terms.
