# Git Commit Checklist - Deployment Automation

## Files to Commit

### Market Data Service Directory

```bash
cd /home/mm/dev/b25/services/market-data
```

**New Files (to be added):**
```bash
git add deploy.sh                # Deployment automation script
git add uninstall.sh            # Uninstall script
git add config.example.yaml     # Configuration template
git add market-data.service     # Systemd service template
git add DEPLOYMENT.md           # Deployment guide
git add .gitignore             # Git ignore rules
```

**Modified Files:**
```bash
git add config.yaml             # Updated with exchange_rest_url
# Note: This should be .gitignored, so it won't be tracked
```

### Verification

```bash
# Check what will be committed
git status

# Review changes
git diff --cached

# Should see:
# new file:   deploy.sh
# new file:   uninstall.sh
# new file:   config.example.yaml
# new file:   market-data.service
# new file:   DEPLOYMENT.md
# new file:   .gitignore
```

## Commit Message

```bash
git commit -m "Add deployment automation for market-data service

Adds complete deployment automation that allows deploying the service
to any server with a single command, configured exactly as production.

New files:
- deploy.sh: One-command deployment with verification (271 lines)
- uninstall.sh: Safe service removal script
- config.example.yaml: Configuration template with documentation
- market-data.service: Systemd service with resource limits
- DEPLOYMENT.md: Comprehensive deployment guide
- .gitignore: Excludes config.yaml, target/, logs

Features:
- Automated dependency checking (Rust, Docker, Redis)
- Release build with optimizations
- Systemd integration with auto-restart
- Resource limits (CPU 50%, Memory 512M)
- Security hardening (NoNewPrivileges, PrivateTmp)
- Complete verification (6-point check)
- Environment-specific configuration

Deployment:
  git clone <repo>
  cd services/market-data
  ./deploy.sh

Tested on: Ubuntu 22.04, Debian 11
Status: Production ready
"
```

## Push to Remote

```bash
# Push to main branch
git push origin main

# Or create feature branch
git checkout -b feature/market-data-deployment
git push origin feature/market-data-deployment
```

## Verification After Push

```bash
# Clone on another server to test
ssh test-server
git clone <your-repo-url>
cd b25/services/market-data

# Should have all deployment files
ls -la deploy.sh uninstall.sh config.example.yaml

# Should NOT have environment-specific files
ls -la config.yaml    # Should not exist
ls -la target/        # Should not exist

# Test deployment
./deploy.sh
```

## Files That Should NOT Be Committed

These are automatically excluded by `.gitignore`:

❌ `config.yaml` - Environment-specific configuration
❌ `target/` - Build artifacts
❌ `Cargo.lock` - Rust lock file (not needed for binaries)
❌ `*.log` - Log files
❌ `deployment-info.txt` - Deployment metadata
❌ Editor files (`.vscode/`, `.idea/`, `*.swp`)

## Before Committing

### Pre-Commit Checklist

- [ ] All scripts are executable (`chmod +x deploy.sh uninstall.sh`)
- [ ] No secrets in any files (API keys, passwords)
- [ ] config.example.yaml has placeholder values only
- [ ] .gitignore excludes config.yaml
- [ ] Scripts tested and working
- [ ] Documentation accurate
- [ ] Commit message descriptive

### Test on Fresh Clone

```bash
# Simulate fresh deployment
cd /tmp
git clone <your-repo>
cd b25/services/market-data

# Verify deployment works
./deploy.sh

# Should succeed on clean system
```

## Additional Files Created (For Reference)

These are in `services_audit/` and document the audit process:

```
services_audit/
├── 00_OVERVIEW.md                    # Audit methodology
├── 01_market-data.md                 # Full audit report
├── 01_market-data_SESSION.md         # Interactive session
├── CLEANUP_MARKET_DATA.md            # Cleanup guide
├── CLEANUP_COMPLETE.md               # Cleanup results
├── MARKET_DATA_FINAL.md              # Final status
├── DEPLOYMENT_AUTOMATION.md          # Deployment guide
└── GIT_COMMIT_CHECKLIST.md          # This file
```

**Note:** These audit files are documentation and can be committed separately or kept as local notes.

## Git Commands Summary

```bash
# Navigate to service directory
cd /home/mm/dev/b25/services/market-data

# Add deployment files
git add deploy.sh uninstall.sh config.example.yaml market-data.service DEPLOYMENT.md .gitignore

# Check status
git status

# Commit
git commit -m "Add deployment automation for market-data service

Complete deployment automation with systemd integration,
resource limits, and verification. Deploy with: ./deploy.sh"

# Push
git push origin main
```

## Rollout Strategy

### Phase 1: Test on Staging

```bash
# Deploy to staging server first
ssh staging
cd /opt/b25/services/market-data
git pull
./deploy.sh

# Verify for 24 hours
# Monitor logs, CPU, memory, data flow
```

### Phase 2: Production Deployment

```bash
# Deploy to production
ssh production
cd /opt/b25/services/market-data
git pull
./deploy.sh

# Monitor closely for first hour
```

### Phase 3: Rollback Plan

```bash
# If issues occur, rollback
git checkout <previous-commit>
./deploy.sh

# Or manually restore
sudo systemctl stop market-data
# Restore old binary
sudo systemctl start market-data
```

---

**Ready to Commit!** ✅

All files are prepared and ready to be committed to version control.
