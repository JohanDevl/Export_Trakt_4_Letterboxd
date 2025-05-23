# Release Notes

## 🚀 Version 2.0.0 - Major Rewrite & Modernization

**Release Date**: December 15, 2024  
**Status**: ✅ Ready for Production

### 🌟 **What's New**

Version 2.0.0 represents a **complete rewrite** of Export Trakt 4 Letterboxd, transforming it from a collection of shell scripts into a robust, enterprise-grade Go application with modern tooling and comprehensive documentation.

---

## 🎯 **Major Features**

### **🔄 New Execution Modes**
- **`--run` Mode**: Immediate one-time execution for testing and quick exports
- **`--schedule` Mode**: Built-in cron scheduler with comprehensive validation
- **Flexible Scheduling**: Support for complex cron expressions with helpful error messages

```bash
# Quick one-time export
./export_trakt --run --export all --mode complete

# Schedule automated exports every 6 hours
./export_trakt --schedule "0 */6 * * *" --export all --mode complete
```

### **🌍 Internationalization (i18n)**
- **Multi-language Support**: English, French, German, Spanish
- **Configurable Language**: Easy language switching via configuration
- **Localized Messages**: All user-facing text supports translation

### **🐳 Production-Ready Docker**
- **Multi-architecture Images**: Support for amd64, arm64, armv7
- **Docker Compose Profiles**: Organized workflows for development and production
- **Optimized Images**: Minimal attack surface with multi-stage builds

### **🧪 Comprehensive Testing**
- **High Test Coverage**: 78%+ coverage across all packages
- **Integration Tests**: Real API interaction testing with mocks
- **Automated CI/CD**: GitHub Actions pipeline with multi-platform testing

---

## 💫 **Enhanced User Experience**

### **📚 Documentation Overhaul**
- **Enterprise-Grade README**: Professional layout with comprehensive guides
- **Modern GitHub Templates**: YAML-based issue forms with structured validation
- **Troubleshooting Guide**: Detailed solutions for common problems
- **Development Setup**: Complete contributor onboarding documentation

### **⚙️ Improved Configuration**
- **TOML Configuration**: Structured, human-readable configuration files
- **Environment Variables**: Support for containerized deployments
- **Validation**: Comprehensive config validation with helpful error messages

### **📊 Better Export Options**
- **Multiple Export Types**: watched, watchlist, collection, shows, all
- **Export Modes**: normal, complete, initial
- **Enhanced CSV Format**: Optimized for Letterboxd import compatibility

---

## 🔧 **Technical Improvements**

### **⚡ Performance**
- **Go Implementation**: Significant performance improvements over shell scripts
- **Concurrent Processing**: Parallel API requests where appropriate
- **Memory Efficiency**: Optimized data structures and processing

### **🛡️ Security**
- **Secure Token Handling**: Best practices for API credential management
- **Input Validation**: Comprehensive validation of all user inputs
- **Error Handling**: Secure error messages that don't leak sensitive information

### **📝 Logging**
- **Structured Logging**: JSON and text format support
- **Configurable Levels**: Debug, info, warn, error levels
- **File Rotation**: Automatic log file management

---

## 🔄 **Migration Guide**

### **From v1.x to v2.0**

#### **Configuration Migration**
```bash
# Old: Environment variables
export TRAKT_CLIENT_ID="your-id"
export TRAKT_CLIENT_SECRET="your-secret"

# New: config.toml file
[trakt]
client_id = "your-id"
client_secret = "your-secret"
```

#### **Command Changes**
```bash
# Old: Shell script
./export.sh

# New: Go binary with flags
./export_trakt --run --export all --mode complete
```

#### **Docker Updates**
```bash
# Pull latest multi-arch image
docker pull johandevl/export-trakt-4-letterboxd:v2.0.0

# Use new Docker Compose profiles
docker compose --profile run-all up
```

---

## 📦 **Installation Options**

### **Docker (Recommended)**
```bash
# Quick start
docker run --rm -it \
  -v $(pwd)/config:/app/config \
  -v $(pwd)/exports:/app/exports \
  johandevl/export-trakt-4-letterboxd:v2.0.0

# Docker Compose
git clone https://github.com/JohanDevl/Export_Trakt_4_Letterboxd.git
cd Export_Trakt_4_Letterboxd
docker compose --profile run-all up
```

### **Binary Download**
```bash
# Download for your platform
wget https://github.com/JohanDevl/Export_Trakt_4_Letterboxd/releases/download/v2.0.0/export_trakt_linux_amd64
chmod +x export_trakt_linux_amd64
./export_trakt_linux_amd64 --help
```

### **Build from Source**
```bash
git clone https://github.com/JohanDevl/Export_Trakt_4_Letterboxd.git
cd Export_Trakt_4_Letterboxd
go build -o export_trakt ./cmd/export_trakt/
```

---

## 🔍 **Breaking Changes**

### **Configuration Format**
- Environment variables are no longer the primary configuration method
- TOML configuration file is now required
- Some configuration keys have changed names

### **Command Line Interface**
- New CLI structure with explicit flags
- Scheduling is now built-in rather than external cron
- Export types and modes are more explicitly defined

### **Docker Image Structure**
- New base image and file structure
- Different volume mount points
- New environment variable names

---

## 🚀 **Getting Started**

### **Quick Setup**
1. **Download the application** (Docker or binary)
2. **Copy the configuration template**: `cp config/config.example.toml config/config.toml`
3. **Edit your configuration** with Trakt.tv API credentials
4. **Run your first export**: `./export_trakt --run --export watched --mode normal`

### **Production Deployment**
1. **Use specific version tags** in production: `v2.0.0`
2. **Set up automated scheduling** with the built-in scheduler
3. **Configure logging** for monitoring and debugging
4. **Set up volume mounts** for persistent data

---

## 🤝 **Community & Support**

### **Getting Help**
- **📖 Documentation**: [Project Wiki](https://github.com/JohanDevl/Export_Trakt_4_Letterboxd/wiki)
- **❓ Questions**: Use our structured [issue templates](https://github.com/JohanDevl/Export_Trakt_4_Letterboxd/issues/new/choose)
- **🐳 Docker Hub**: [johandevl/export-trakt-4-letterboxd](https://hub.docker.com/r/johandevl/export-trakt-4-letterboxd)

### **Contributing**
- **🔧 Bug Reports**: Use the Bug Report template with complete information
- **✨ Feature Requests**: Share your ideas with the Feature Request template
- **📚 Documentation**: Help improve our documentation
- **💻 Code**: Submit pull requests following our contribution guidelines

---

## 🙏 **Acknowledgments**

- **Original Creator**: [u2pitchjami](https://github.com/u2pitchjami) for the original concept and implementation
- **Current Maintainer**: [JohanDevl](https://github.com/JohanDevl) for the v2.0 rewrite
- **Community**: All users who provided feedback and suggestions
- **Beta Testers**: Contributors who helped test and improve the application

---

## 🔮 **What's Next**

### **Upcoming Features (v2.1)**
- **Enhanced Letterboxd Integration**: Better format compatibility
- **Additional Export Formats**: JSON, XML support
- **Advanced Filtering**: More granular export controls
- **Web UI**: Optional web interface for easier configuration

### **Long-term Roadmap**
- **Plugin System**: Extensible architecture for custom exporters
- **Real-time Sync**: Continuous synchronization options
- **Mobile App**: Companion mobile application
- **Cloud Deployment**: One-click cloud deployment options

---

**Ready to upgrade? Check out our [comprehensive migration guide](https://github.com/JohanDevl/Export_Trakt_4_Letterboxd/wiki/Migration-Guide) and start enjoying the enhanced Export Trakt 4 Letterboxd experience!** 🎬✨
