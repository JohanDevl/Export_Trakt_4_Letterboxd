# Security Policy

## 🔒 Supported Versions

We are committed to maintaining security across our supported versions. The following table shows which versions currently receive security updates:

| Version | Support Status     | End of Life | Notes                               |
| ------- | ------------------ | ----------- | ----------------------------------- |
| 2.x     | ✅ **Supported**   | TBD         | Current stable branch               |
| latest  | ✅ **Supported**   | Rolling     | Latest Docker images                |
| main    | ✅ **Supported**   | Rolling     | Main branch, same as latest         |
| develop | ⚠️ **Development** | N/A         | Unstable, for development only      |
| 1.x     | ❌ **End of Life** | 2025-05-23  | Legacy version, no longer supported |

### Docker Images

We maintain security updates for:

- `johandevl/export-trakt-4-letterboxd:latest`
- `johandevl/export-trakt-4-letterboxd:v2.x.x` (specific versions)
- `ghcr.io/johandevl/export_trakt_4_letterboxd:latest`

## 🚨 Reporting a Vulnerability

**Please do not report security vulnerabilities through public GitHub issues.** Instead, please follow our responsible disclosure process.

### How to Report

We encourage responsible disclosure of security vulnerabilities. Please report security issues via one of the following methods:

1. **Primary**: Email the maintainer directly at: **[Create an issue with `@JohanDevl` mention]**
2. **GitHub Security Advisory**: Use GitHub's [private vulnerability reporting](https://github.com/JohanDevl/Export_Trakt_4_Letterboxd/security/advisories/new)
3. **Twitter DM**: [@0xUta](https://twitter.com/0xUta) for urgent issues

### Information to Include

When reporting a vulnerability, please include:

- **Vulnerability Type**: What kind of security issue is it?
- **Impact Assessment**: What's the potential impact?
- **Affected Versions**: Which versions are affected?
- **Reproduction Steps**: Detailed steps to reproduce the issue
- **Proof of Concept**: If available (but avoid actual exploitation)
- **Suggested Fix**: If you have ideas for remediation
- **Disclosure Timeline**: Your preferred timeline for public disclosure

### Example Report Template

```
Subject: [SECURITY] Vulnerability in Export Trakt 4 Letterboxd

Vulnerability Type: [e.g., Authentication bypass, Injection, etc.]
Affected Versions: [e.g., 2.0.0 - 2.1.5]
Severity: [Critical/High/Medium/Low]

Description:
[Detailed description of the vulnerability]

Impact:
[What could an attacker accomplish with this vulnerability?]

Steps to Reproduce:
1. [Step 1]
2. [Step 2]
3. [Vulnerability is triggered]

Suggested Mitigation:
[Your suggestions for fixing the issue]
```

## 📋 Security Response Process

### Our Commitment

- **Initial Response**: We will acknowledge receipt of your vulnerability report within **48 hours**
- **Assessment**: We will assess and validate the reported vulnerability within **5 business days**
- **Resolution**: We will work on a fix and coordinate disclosure timing with you
- **Communication**: We will keep you informed throughout the process

### Response Timeline

1. **0-48 hours**: Acknowledgment of report
2. **2-5 days**: Initial assessment and validation
3. **5-14 days**: Development of fix (depending on complexity)
4. **14-30 days**: Testing and release preparation
5. **30+ days**: Public disclosure (coordinated with reporter)

### Disclosure Policy

- We believe in **coordinated disclosure**
- We will work with you to determine an appropriate disclosure timeline
- We will credit you in our security advisory (unless you prefer anonymity)
- We will not take legal action against security researchers acting in good faith

## 🛡️ Security Best Practices

### For Users

When deploying Export Trakt 4 Letterboxd:

#### 🔑 **API Security**

- **Rotate your Trakt.tv API tokens regularly**
- **Use environment variables** or secure secret management for API credentials
- **Never commit API keys** to version control
- **Limit API token permissions** to the minimum required scope

#### 🐳 **Docker Security**

```bash
# Use specific version tags, not 'latest' in production
docker pull johandevl/export-trakt-4-letterboxd:v2.1.0

# Run with limited privileges
docker run --user 1000:1000 --read-only \
  -v $(pwd)/config:/app/config:ro \
  -v $(pwd)/exports:/app/exports \
  johandevl/export-trakt-4-letterboxd:v2.1.0

# Use Docker secrets for sensitive data
echo "your-api-token" | docker secret create trakt-token -
```

#### 🔒 **Configuration Security**

```toml
# config.toml - Secure configuration example
[trakt]
client_id = "${TRAKT_CLIENT_ID}"      # Use environment variables
client_secret = "${TRAKT_CLIENT_SECRET}"
access_token = "${TRAKT_ACCESS_TOKEN}"

[logging]
level = "info"                        # Avoid debug logs in production
file = "logs/export.log"
max_size = "10MB"                     # Limit log file sizes

[export]
output_dir = "./exports"              # Use relative paths when possible
```

#### 🌐 **Network Security**

- **Use HTTPS** for all API communications (default)
- **Consider using a VPN** for sensitive environments
- **Firewall rules** to limit outbound connections if needed
- **Monitor network traffic** for unexpected connections

#### 📁 **File System Security**

```bash
# Set appropriate file permissions
chmod 600 config/config.toml          # Config file readable only by owner
chmod 755 exports/                    # Export directory
chmod 644 exports/*.csv               # Export files

# Use dedicated user for running the application
useradd -r -s /bin/false export-user
chown -R export-user:export-user /app/
```

### For Developers

#### 🔍 **Code Security**

- **Input Validation**: Validate all user inputs and API responses
- **Error Handling**: Don't expose sensitive information in error messages
- **Dependency Management**: Keep dependencies updated and audit regularly
- **Secrets Management**: Never hardcode secrets in source code

#### 🧪 **Security Testing**

```bash
# Static analysis
go vet ./...
golangci-lint run

# Dependency vulnerability scanning
go list -json -m all | nancy sleuth

# Security audit
gosec ./...

# Container scanning
docker scan johandevl/export-trakt-4-letterboxd:latest
```

#### 📦 **Build Security**

- **Reproducible builds** with pinned dependencies
- **Signed releases** for binary distributions
- **Multi-stage Docker builds** to minimize attack surface
- **Regular base image updates**

## 🔍 Security Monitoring

### Automated Security

- **GitHub Security Advisories**: Automated dependency vulnerability scanning
- **CodeQL Analysis**: Static code analysis for security issues
- **Container Scanning**: Regular Docker image vulnerability scans
- **Dependency Updates**: Automated security updates via Dependabot

### Manual Reviews

- **Code Reviews**: All code changes undergo security-focused review
- **Security Audits**: Regular manual security assessments
- **Penetration Testing**: Periodic security testing of the application
- **Threat Modeling**: Regular assessment of potential attack vectors

## 📚 Security Resources

### Documentation

- [OWASP Go Security Cheat Sheet](https://cheatsheetseries.owasp.org/cheatsheets/Go_SG_Cheat_Sheet.html)
- [Docker Security Best Practices](https://docs.docker.com/engine/security/)
- [Trakt.tv API Security](https://trakt.docs.apiary.io/#introduction/required-headers)

### Tools

- [gosec](https://github.com/securecodewarrior/gosec) - Go security analyzer
- [nancy](https://github.com/sonatypeoss/nancy) - Dependency vulnerability scanner
- [docker-bench-security](https://github.com/docker/docker-bench-security) - Docker security audit

## 🏆 Security Recognition

We appreciate security researchers who help keep our project secure. Contributors who report valid security vulnerabilities will be:

- **Acknowledged** in our security advisories (with permission)
- **Listed** in our hall of fame (if desired)
- **Invited** to test future releases for security issues
- **Considered** for bug bounty rewards (when available)

## 📞 Contact

For non-security related issues, please use our regular [issue templates](https://github.com/JohanDevl/Export_Trakt_4_Letterboxd/issues/new/choose).

**Security Contact**: [@JohanDevl](https://github.com/JohanDevl) | [@0xUta](https://twitter.com/0xUta)

---

Thank you for helping keep Export Trakt 4 Letterboxd secure! 🔒

# Security Guide

Export Trakt 4 Letterboxd implements comprehensive security measures to protect user data and ensure secure operation. This document outlines the security features and best practices.

## Security Architecture

### Overview

The application follows a defense-in-depth security approach with multiple layers of protection:

1. **Credential Management** - Secure storage and handling of API credentials
2. **Data Protection** - Encryption of sensitive data at rest and in transit
3. **Access Control** - File permission enforcement and path validation
4. **Network Security** - HTTPS enforcement and secure HTTP client configuration
5. **Rate Limiting** - Protection against API abuse and DoS attacks
6. **Audit Logging** - Comprehensive security event monitoring
7. **Input Validation** - Protection against injection attacks

## Security Components

### 1. Credential Management

#### Features

- **AES-256 encryption** for stored API credentials
- **Multiple storage backends**: system keyring, environment variables, encrypted files
- **Automatic credential rotation** support
- **Credential validation** on startup
- **Secure credential retrieval** with audit logging

#### Configuration

```toml
[security]
encryption_enabled = true
keyring_backend = "system"  # system, env, file

[security.keyring]
service_name = "export_trakt_4_letterboxd"
username = "default"
```

#### Environment Variables

```bash
export TRAKT_CLIENT_ID="your_client_id"
export TRAKT_CLIENT_SECRET="your_client_secret"
export ENCRYPTION_KEY="base64_encoded_key"
```

#### Usage

The application automatically manages credentials through the security manager:

- Credentials are encrypted before storage
- Access is logged and monitored
- Invalid or expired credentials trigger rotation

### 2. Data Protection

#### File Permission Enforcement

- **Config files**: 0600 (owner read/write only)
- **Data files**: 0644 (owner read/write, group/others read)
- **Directories**: 0750 (owner full, group read/execute)
- **Automatic permission validation** and correction

#### Encryption

- **AES-256-GCM** encryption for sensitive data
- **Secure key generation** using crypto/rand
- **Key derivation** from user-provided or auto-generated keys
- **Encrypted storage** for export files containing sensitive data

#### Configuration

```toml
[security.filesystem]
enforce_permissions = true
config_file_mode = 0600
data_file_mode = 0644
directory_mode = 0750
max_file_size = 104857600  # 100MB
check_symlinks = true
```

### 3. Access Control

#### Path Validation

- **Path traversal protection** - Blocks ../ and ..\ patterns
- **Allowed path enforcement** - Restricts access to specified directories
- **Restricted path blocking** - Prevents access to system directories
- **Symlink attack prevention** - Validates symlink targets

#### Secure File Operations

```go
// Create file with secure permissions
file, err := securityManager.SecureCreateFile(path, 0600)

// Write data with automatic permission enforcement
err := securityManager.SecureWriteFile(path, data, true) // isConfig=true

// Validate file permissions
err := securityManager.ValidateFilePermissions(path)
```

### 4. Network Security

#### HTTPS Enforcement

- **Mandatory HTTPS** for all external API calls
- **TLS 1.2 minimum** version requirement
- **Strong cipher suites** only
- **Certificate validation** (no insecure skip verify in production)
- **HTTP Strict Transport Security (HSTS)** support

#### Secure HTTP Client

```go
// Create secure HTTP client
client := httpsEnforcer.CreateSecureClient()

// Validate URLs before requests
err := httpsEnforcer.ValidateURL("https://api.trakt.tv")

// Secure request with security headers
err := httpsEnforcer.SecureRequest(req)
```

#### Configuration

```toml
[security.https]
require_https = true
allow_insecure = false      # Only for development
tls_min_version = 771       # TLS 1.2
timeout = "30s"
max_redirects = 5
allowed_hosts = ["api.trakt.tv", "api.themoviedb.org"]
blocked_hosts = ["localhost", "127.0.0.1"]
enable_hsts = true
```

### 5. Rate Limiting

#### Token Bucket Algorithm

- **Per-service rate limits** with configurable parameters
- **Burst capacity** for handling traffic spikes
- **Automatic token refill** based on configured rates
- **Context-aware waiting** with timeout support

#### Service-Specific Limits

```toml
[security.rate_limit]
enabled = true
default_limit = 60          # requests per minute
burst_limit = 10            # burst capacity
cleanup_interval = "5m"

[security.rate_limit.limits.trakt_api]
requests_per_minute = 40
burst_capacity = 5
window = "1m"

[security.rate_limit.limits.auth]
requests_per_minute = 10
burst_capacity = 3
window = "1m"
```

#### Usage

```go
// Check if request is allowed
if !securityManager.AllowRequest("trakt_api") {
    return ErrRateLimited
}

// Wait for permission (with context timeout)
ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
defer cancel()
err := securityManager.WaitForRequest(ctx, "trakt_api")
```

### 6. Audit Logging

#### Comprehensive Event Tracking

- **Authentication events** - Login, logout, failures
- **Credential operations** - Access, storage, rotation
- **Data operations** - Export, encryption, decryption
- **Security violations** - Unauthorized access attempts
- **System events** - Startup, shutdown, errors

#### Structured JSON Logging

```json
{
  "timestamp": "2024-01-15T10:30:00Z",
  "event_type": "credential_access",
  "severity": "medium",
  "source": "security_manager",
  "action": "retrieve_credentials",
  "result": "success",
  "message": "API credentials retrieved successfully",
  "details": {
    "credential_type": "api_credentials"
  }
}
```

#### Configuration

```toml
[security.audit]
log_level = "info"
retention_days = 90
include_sensitive = false
output_format = "json"
```

### 7. Input Validation

#### Protection Against Common Attacks

- **SQL injection prevention** - Input sanitization
- **XSS protection** - HTML encoding of outputs
- **Path traversal prevention** - Path validation
- **Command injection prevention** - Input filtering

#### Validation Rules

```go
// Validate export path
err := validator.ValidateExportPath(path)

// Validate configuration value
err := validator.ValidateConfigValue(field, value)

// Sanitize input for safe processing
clean := validator.SanitizeInput(userInput)

// Sanitize input for logging
logSafe := validator.SanitizeForLog(userInput)
```

## Security Configuration

### Complete Configuration Example

```toml
[security]
encryption_enabled = true
keyring_backend = "system"
audit_logging = true
rate_limit_enabled = true
require_https = true

[security.audit]
log_level = "info"
retention_days = 90
include_sensitive = false
output_format = "json"

[security.rate_limit]
enabled = true
default_limit = 60
burst_limit = 10
window_duration = "1m"
cleanup_interval = "5m"

[security.filesystem]
enforce_permissions = true
config_file_mode = 0600
data_file_mode = 0644
directory_mode = 0750
allowed_base_paths = ["./config", "./exports", "./logs", "./temp"]
restricted_paths = ["/etc", "/var", "/usr", "/sys", "/proc", "/dev"]
max_file_size = 104857600
check_symlinks = true

[security.https]
require_https = true
allow_insecure = false
tls_min_version = 771
timeout = "30s"
max_redirects = 5
allowed_hosts = ["api.trakt.tv", "api.themoviedb.org"]
blocked_hosts = ["localhost", "127.0.0.1"]
enable_hsts = true
```

## Security Best Practices

### For Users

1. **Use strong API credentials** from Trakt.tv
2. **Keep software updated** to latest version
3. **Secure your config directory** with appropriate permissions
4. **Monitor audit logs** for suspicious activity
5. **Use HTTPS URLs** for all API endpoints
6. **Don't disable security features** in production

### For Developers

1. **Never hardcode credentials** in source code
2. **Use the security manager** for all credential operations
3. **Validate all user inputs** before processing
4. **Log security events** appropriately
5. **Follow secure coding practices**
6. **Test security features** thoroughly

### For Deployment

1. **Use secure file permissions** (0600 for config files)
2. **Enable all security features** in production
3. **Monitor audit logs** regularly
4. **Use strong encryption keys**
5. **Implement backup strategies** for credentials
6. **Keep dependencies updated**

## Security Monitoring

### Audit Log Analysis

Monitor audit logs for:

- **Failed authentication attempts**
- **Unusual credential access patterns**
- **Security violation events**
- **Rate limit violations**
- **Unauthorized file access attempts**

### Security Metrics

The application provides security metrics:

```go
metrics := securityManager.GetSecurityMetrics()
// Returns: encryption status, audit metrics, rate limit stats
```

### Log Retention

- **Default retention**: 90 days
- **Automatic cleanup** of old logs
- **Configurable retention periods**
- **Secure log file permissions**

## Incident Response

### Security Violations

1. **Automatic blocking** of suspicious requests
2. **Detailed audit logging** of security events
3. **Alert generation** for critical violations
4. **Graceful degradation** when possible

### Credential Compromise

1. **Immediate credential rotation**
2. **Audit log analysis** for unauthorized access
3. **Security event notifications**
4. **Recovery procedures** documentation

## Compliance Considerations

### Data Protection

- **GDPR compliance** ready features
- **Data minimization** practices
- **Secure data handling** procedures
- **User consent** mechanisms

### Industry Standards

- **OWASP** security guidelines compliance
- **NIST** cybersecurity framework alignment
- **ISO 27001** security controls implementation
- **SOC 2** compliance readiness

## Troubleshooting

### Common Issues

1. **Permission denied errors** - Check file permissions
2. **Credential access failures** - Verify keyring setup
3. **Rate limit violations** - Adjust rate limit configuration
4. **HTTPS validation errors** - Check allowed hosts configuration

### Debug Mode

For troubleshooting, temporarily enable debug logging:

```toml
[security.audit]
log_level = "debug"
include_sensitive = true  # Only for debugging, never in production
```

## Updates and Maintenance

### Security Updates

- **Regular dependency updates** for security patches
- **Security feature enhancements** based on threat landscape
- **Vulnerability assessments** and remediation
- **Security configuration reviews**

### Monitoring Tools

Consider integrating with:

- **SIEM systems** for log analysis
- **Vulnerability scanners** for dependency checks
- **Intrusion detection systems** for real-time monitoring
- **Security information dashboards** for visibility

---

For technical support or security concerns, please refer to our [security policy](SECURITY.md#reporting-vulnerabilities) or open an issue on GitHub.
