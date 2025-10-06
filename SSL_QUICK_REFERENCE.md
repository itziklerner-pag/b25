# B25 SSL Setup - Quick Reference Card

## ✅ Setup Complete!

**Domain**: `mm.itziklerner.com` → `66.94.120.149`
**SSL**: ✅ Active (ZeroSSL ECC-256, expires Jan 4, 2026)
**Auto-Renewal**: ✅ Daily cron at 14:10

---

## 🌐 Production URLs

```
Web Dashboard:  https://mm.itziklerner.com
WebSocket:      wss://mm.itziklerner.com/ws?type=web
API Gateway:    https://mm.itziklerner.com/api
Grafana:        https://mm.itziklerner.com/grafana
Health Check:   https://mm.itziklerner.com/health
```

---

## 🚀 Start Services

```bash
# Start all production services
bash /home/mm/dev/b25/start-production-services.sh

# Verify setup
sudo bash /home/mm/dev/b25/verify-ssl-setup.sh
```

---

## 🔧 Essential Commands

### Nginx
```bash
sudo systemctl reload nginx        # Reload config
sudo nginx -t                      # Test config
sudo tail -f /var/log/nginx/b25-access.log
```

### SSL Certificate
```bash
sudo /root/.acme.sh/acme.sh --info -d mm.itziklerner.com --ecc
sudo /root/.acme.sh/acme.sh --renew -d mm.itziklerner.com --force --ecc
```

### Services
```bash
# Check ports
sudo netstat -tlnp | grep -E "(80|443|3000|8000|8086)"

# Service status
systemctl status nginx
```

---

## 📁 Key Files

| File | Path |
|------|------|
| Nginx Config | `/etc/nginx/sites-available/mm.itziklerner.com` |
| SSL Cert | `/etc/nginx/ssl/mm.itziklerner.com.crt` |
| SSL Key | `/etc/nginx/ssl/mm.itziklerner.com.key` |
| Web .env | `/home/mm/dev/b25/ui/web/.env` |
| Access Log | `/var/log/nginx/b25-access.log` |
| Error Log | `/var/log/nginx/b25-error.log` |

---

## ⚠️ Action Required

### 1. Start API Gateway
```bash
cd /home/mm/dev/b25/services/api-gateway
npm start
```

### 2. Deploy Production Web
```bash
cd /home/mm/dev/b25/ui/web
npm run build
npm run preview
```

### 3. Enable Firewall (Optional)
```bash
sudo ufw enable
```

---

## 🔄 Auto-Renewal

- **Cron**: Daily at 14:10
- **Command**: `/root/.acme.sh/acme.sh --cron`
- **Next Renewal**: Dec 4, 2025
- **Hook**: Automatically reloads Nginx

---

## 📊 Current Status

| Component | Status |
|-----------|--------|
| SSL Certificate | ✅ Active |
| Nginx | ✅ Running |
| Web Dashboard | ✅ Running (dev mode) |
| Dashboard Server | ✅ Running |
| Grafana | ✅ Running |
| API Gateway | ❌ Not Running |

---

## 🧪 Quick Tests

```bash
# Test HTTPS
curl -I https://mm.itziklerner.com/health

# Test redirect
curl -I http://mm.itziklerner.com

# Test SSL
openssl s_client -connect mm.itziklerner.com:443 < /dev/null
```

---

## 📚 Documentation

Full docs in `/home/mm/dev/b25/`:
- `SSL_SETUP_COMPLETE.md` - Complete setup guide
- `NGINX_SSL_STATUS.md` - Detailed status report
- `verify-ssl-setup.sh` - Verification script

---

**Last Updated**: Oct 6, 2025
