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
