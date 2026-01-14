# CI/CD Setup Guide

This document explains how to set up and configure the CI/CD pipeline for Export Trakt 4 Letterboxd.

## GitHub Secrets Configuration

### Required Secrets

#### 1. PAT_TOKEN (Personal Access Token) - **REQUIRED**

**Purpose:** The `PAT_TOKEN` is required for the auto-tag workflow to trigger downstream CI/CD workflows.

**Why it's needed:**
- GitHub Actions workflows cannot trigger other workflows when using the default `GITHUB_TOKEN`
- The auto-tag workflow needs to create tags that trigger the CI/CD pipeline
- Without `PAT_TOKEN`, the Docker build workflow will NOT be triggered automatically

**How to create:**

1. Go to GitHub Settings → Developer settings → Personal access tokens → Tokens (classic)
2. Click "Generate new token (classic)"
3. Give it a descriptive name: "Export Trakt CI/CD Automation"
4. Set expiration (recommended: 90 days with renewal reminder)
5. Select the following scopes:
   - `repo` (Full control of private repositories)
   - `workflow` (Update GitHub Action workflows)
6. Click "Generate token"
7. **Copy the token immediately** (you won't see it again)

**How to add to repository:**

1. Go to your repository → Settings → Secrets and variables → Actions
2. Click "New repository secret"
3. Name: `PAT_TOKEN`
4. Value: Paste your personal access token
5. Click "Add secret"

**Important Notes:**
- The `PAT_TOKEN` must have `repo` and `workflow` scopes
- If `PAT_TOKEN` is not configured, the auto-tag workflow will fall back to `GITHUB_TOKEN`, but **this will NOT trigger CI/CD workflows**
- You need to renew the token before expiration to maintain automation
- Store the token securely - treat it like a password

#### 2. DOCKERHUB_USERNAME and DOCKERHUB_TOKEN - Optional

**Purpose:** Publishing Docker images to Docker Hub (in addition to GitHub Container Registry).

**How to create:**

1. Go to Docker Hub → Account Settings → Security
2. Click "New Access Token"
3. Give it a description: "Export Trakt CI/CD"
4. Set permissions: Read, Write, Delete
5. Click "Generate"
6. Copy the token

**How to add to repository:**

1. Add `DOCKERHUB_USERNAME` secret with your Docker Hub username
2. Add `DOCKERHUB_TOKEN` secret with your access token

**Note:** If these are not configured, images will only be published to GitHub Container Registry (ghcr.io).

## Workflow Overview

### Workflow Trigger Chain

```
PR Merged to main
    ↓
auto-tag.yml (creates v2.0.x tag using PAT_TOKEN)
    ↓
ci-cd.yml (triggered by tag push)
    ↓
Docker images built and published
    ↓
docker-tag-monitor.yml (verifies images exist, runs daily)
```

### Workflows and Their Dependencies

#### 1. **go-tests.yml** - Go Testing
- **Triggers:** Push to main/develop/feature branches, PRs
- **Dependencies:** None
- **Purpose:** Quick test validation

#### 2. **ci-cd.yml** - Main CI/CD Pipeline
- **Triggers:** Push to main/develop, tags (v*), PRs, releases
- **Dependencies:** None
- **Purpose:** Build, test, Docker build, security scan

#### 3. **auto-tag.yml** - Automatic Versioning
- **Triggers:** PR merge to main
- **Dependencies:** Requires `PAT_TOKEN` secret
- **Purpose:** Create semantic version tags automatically

#### 4. **release.yml** - Release Creation
- **Triggers:** Git tags (v*)
- **Dependencies:** None
- **Purpose:** Build multi-platform binaries and create GitHub release

#### 5. **docker-tag-monitor.yml** - Docker Image Monitoring
- **Triggers:** Daily at 6 AM UTC, manual dispatch
- **Dependencies:** Requires `PAT_TOKEN` to trigger ci-cd.yml
- **Purpose:** Verify Docker images exist for all tags

#### 6. **docker-cleanup.yml** - Docker Image Cleanup
- **Triggers:** PR closure, daily at 2 AM UTC, manual dispatch
- **Dependencies:** None
- **Purpose:** Clean up obsolete Docker images

#### 7. **security-scan.yml** - Security Scanning
- **Triggers:** Push to main/develop, PRs, daily at 2 AM UTC
- **Dependencies:** None
- **Purpose:** Security vulnerability scanning

## Troubleshooting

### Issue: Auto-tag creates tag but CI/CD doesn't run

**Cause:** `PAT_TOKEN` is not configured or lacks required permissions.

**Solution:**
1. Verify `PAT_TOKEN` secret exists in repository settings
2. Ensure the token has `repo` and `workflow` scopes
3. Check if the token has expired
4. Review the auto-tag workflow run logs

### Issue: Docker images not published

**Cause:** Authentication failure or missing credentials.

**Solution:**
1. For GitHub Container Registry: Verify GitHub Actions permissions
2. For Docker Hub: Check `DOCKERHUB_USERNAME` and `DOCKERHUB_TOKEN` secrets
3. Review ci-cd.yml workflow logs for authentication errors

### Issue: Coverage check fails

**Cause:** Test coverage below 70% threshold.

**Solution:**
1. Run tests locally: `go test -coverprofile=coverage.out ./pkg/...`
2. Check coverage: `go tool cover -func=coverage.out | grep total`
3. Add tests to improve coverage
4. Ensure new code includes tests

### Issue: Security scan fails

**Cause:** High or critical vulnerabilities detected.

**Solution:**
1. Review security-scan.yml workflow logs
2. Check gosec, Trivy, and dependency scan results
3. Update vulnerable dependencies: `go get -u ./...`
4. Fix security issues in code
5. Re-run tests and scans

## Best Practices

### 1. Token Management
- Set token expiration to 90 days
- Set up calendar reminders to renew tokens
- Use organization-level tokens for organization repositories
- Rotate tokens regularly

### 2. Branch Protection
- Require status checks to pass before merging
- Require pull request reviews
- Enable "Require branches to be up to date before merging"

### 3. Workflow Optimization
- Use caching for Go modules to speed up builds
- Run expensive jobs (Docker builds) only when necessary
- Use matrix builds for parallel execution

### 4. Security
- Enable Dependabot for automatic dependency updates
- Review security scan results regularly
- Keep Docker base images up to date
- Use minimal privileges for tokens

## Monitoring

### GitHub Actions Dashboard
- Monitor workflow runs at: `https://github.com/YOUR_USERNAME/YOUR_REPO/actions`
- Set up notifications for workflow failures
- Review workflow run times and optimize slow jobs

### Docker Registry Monitoring
- Check Docker Hub: `https://hub.docker.com/r/YOUR_USERNAME/YOUR_REPO/tags`
- Check GHCR: `https://github.com/YOUR_USERNAME/YOUR_REPO/pkgs/container/YOUR_REPO`
- Verify image tags match git tags

## Additional Resources

- [GitHub Actions Documentation](https://docs.github.com/en/actions)
- [Creating a Personal Access Token](https://docs.github.com/en/authentication/keeping-your-account-and-data-secure/creating-a-personal-access-token)
- [Docker Hub Access Tokens](https://docs.docker.com/docker-hub/access-tokens/)
- [Dependabot Documentation](https://docs.github.com/en/code-security/dependabot)

## Support

For issues or questions:
1. Check existing GitHub issues
2. Review workflow logs
3. Create a new issue with:
   - Workflow name
   - Error message
   - Steps to reproduce
   - Relevant logs
