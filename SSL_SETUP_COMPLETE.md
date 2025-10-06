# B25 Trading System - SSL Setup Complete

## ✅ Setup Status

**Domain**: mm.itziklerner.com
**Server IP**: 66.94.120.149
**SSL Certificate**: ✅ Active (ZeroSSL via acme.sh)
**Certificate Expiry**: January 4, 2026
**Auto-Renewal**: ✅ Configured (cron job)

---

## 🔐 SSL Certificate Details

### Certificate Information
- **Issuer**: ZeroSSL ECC Domain Secure Site CA
- **Type**: ECC (Elliptic Curve) - 256-bit
- **Certificate Path**: `/etc/nginx/ssl/mm.itziklerner.com.crt`
- **Private Key Path**: `/etc/nginx/ssl/mm.itziklerner.com.key`
- **Certificate Created**: October 6, 2025
- **Next Renewal**: December 4, 2025
- **Auto-Renewal**: Daily check at 14:10 via cron

### SSL Security Configuration
- **Protocols**: TLSv1.2, TLSv1.3
- **Cipher Suite**: Mozilla Intermediate Configuration
- **HSTS**: Enabled (max-age: 63072000 seconds / 2 years)
- **SSL Stapling**: Enabled
- **Session Cache**: 10m shared

---

## 🌐 Nginx Reverse Proxy Configuration

### File Location
`/etc/nginx/sites-available/mm.itziklerner.com`

### Proxy Routes

| Path | Backend | Description |
|------|---------|-------------|
| `/` | `http://localhost:3000` | Web Dashboard (React/Vite) |
| `/api` | `http://localhost:8000` | API Gateway |
| `/ws` | `http://localhost:8086/ws` | Dashboard WebSocket |
| `/grafana/` | `http://localhost:3001/` | Grafana Dashboard |
| `/health` | Nginx | Health check endpoint |
| `/nginx_status` | Nginx | Status (localhost only) |

### WebSocket Configuration
- **Upgrade Headers**: Configured for WebSocket protocol
- **Timeout**: 3600s (1 hour) for long-lived connections
- **Buffering**: Disabled for real-time communication

### Security Headers
```nginx
Strict-Transport-Security: max-age=63072000; includeSubDomains; preload
X-Frame-Options: SAMEORIGIN
X-Content-Type-Options: nosniff
X-XSS-Protection: 1; mode=block
Referrer-Policy: strict-origin-when-cross-origin
```

---

## 🔧 Configuration Files

### 1. Nginx Main Config
- **Path**: `/etc/nginx/sites-available/mm.itziklerner.com`
- **Status**: ✅ Active and symlinked
- **Features**:
  - HTTP to HTTPS redirect
  - SSL termination
  - Reverse proxy for all services
  - WebSocket support
  - CORS headers for API
  - Gzip compression

### 2. Web Dashboard .env
- **Path**: `/home/mm/dev/b25/ui/web/.env`
- **Configuration**:
```env
VITE_WS_URL=wss://mm.itziklerner.com/ws?type=web
VITE_API_URL=https://mm.itziklerner.com/api
VITE_AUTH_URL=https://mm.itziklerner.com/api/auth
NODE_ENV=production
```

### 3. Backup Files
- Environment backup: `/home/mm/dev/b25/ui/web/.env.backup.20251006_032710`

---

## 🚀 Access URLs

### Production URLs (HTTPS)
- **Web Dashboard**: https://mm.itziklerner.com
- **WebSocket**: wss://mm.itziklerner.com/ws?type=web
- **API Gateway**: https://mm.itziklerner.com/api
- **Grafana**: https://mm.itziklerner.com/grafana
- **Health Check**: https://mm.itziklerner.com/health

### HTTP (Redirects to HTTPS)
- http://mm.itziklerner.com → https://mm.itziklerner.com

---

## 📊 Service Status

### ✅ Running Services
- **Nginx**: ✅ Active (nginx/1.24.0)
- **Web Dashboard**: ✅ Port 3000 (Vite dev server)
- **Dashboard Server**: ✅ Port 8086 (WebSocket)
- **Grafana**: ✅ Port 3001

### ⚠️ Services Requiring Attention
- **API Gateway**: ❌ Not running (port 8000)
  - **Action Required**: Start API Gateway service
  - **Command**: `cd /home/mm/dev/b25/services/api-gateway && npm start`

### 🛑 Disabled Services
- **Caddy**: Disabled (was using port 80/443)
  - **Reason**: Replaced by Nginx

---

## 🔄 Certificate Auto-Renewal

### Renewal System
- **Tool**: acme.sh
- **Schedule**: Daily check at 14:10 (cron job)
- **Cron Entry**: `10 14 * * * "/root/.acme.sh"/acme.sh --cron --home "/root/.acme.sh" > /dev/null`
- **Reload Hook**: `/root/.acme.sh/acme.sh` automatically reloads Nginx after renewal

### Manual Renewal Commands
```bash
# Force renewal
sudo /root/.acme.sh/acme.sh --renew -d mm.itziklerner.com --force --ecc

# Check certificate info
sudo /root/.acme.sh/acme.sh --info -d mm.itziklerner.com --ecc

# List all certificates
sudo /root/.acme.sh/acme.sh --list
```

---

## 🛠️ Management Scripts

### Setup Scripts (✅ Completed)
1. **`/home/mm/dev/b25/setup-nginx-ssl-acme.sh`**
   - Installs Nginx and acme.sh
   - Obtains SSL certificate
   - Configures reverse proxy

2. **`/home/mm/dev/b25/update-env-for-domain.sh`**
   - Updates web dashboard .env for domain
   - Creates backup of old config

### Verification & Monitoring
3. **`/home/mm/dev/b25/verify-ssl-setup.sh`**
   - Checks Nginx status
   - Verifies SSL certificate
   - Tests all endpoints
   - Validates WebSocket connectivity

### Usage
```bash
# Verify SSL setup
sudo bash /home/mm/dev/b25/verify-ssl-setup.sh

# Check Nginx config
sudo nginx -t

# Reload Nginx
sudo systemctl reload nginx

# View Nginx logs
sudo tail -f /var/log/nginx/b25-access.log
sudo tail -f /var/log/nginx/b25-error.log
```

---

## 🔥 Firewall Configuration

### Current Status
- **UFW**: Installed but not active
- **Open Ports**: 80, 443 (configured, awaiting UFW activation)

### Activate Firewall (Optional)
```bash
sudo ufw enable
sudo ufw status
```

---

## 📝 Pending Actions

### 1. Start API Gateway
```bash
cd /home/mm/dev/b25/services/api-gateway
npm install  # if needed
npm start
```

### 2. Build Production Web Dashboard
The web dashboard is currently running in development mode. For production:

```bash
cd /home/mm/dev/b25/ui/web
npm run build
npm run preview  # or serve dist/ folder
```

**Or use a production server:**
```bash
# Install serve globally
npm install -g serve

# Serve production build
serve -s dist -l 3000
```

### 3. Configure Vite for Proxying (Alternative)
If keeping Vite dev server, update `vite.config.ts` to allow external access:

```typescript
export default defineConfig({
  server: {
    host: '0.0.0.0',
    port: 3000,
    strictPort: true,
    hmr: {
      protocol: 'wss',
      host: 'mm.itziklerner.com'
    }
  }
})
```

### 4. Enable UFW Firewall
```bash
sudo ufw enable
sudo ufw allow 22/tcp  # SSH (if not already allowed)
sudo ufw allow 80/tcp
sudo ufw allow 443/tcp
```

---

## 🧪 Testing Checklist

### ✅ Completed Tests
- [x] DNS resolution (mm.itziklerner.com → 66.94.120.149)
- [x] SSL certificate obtained and installed
- [x] HTTP to HTTPS redirect (301)
- [x] HTTPS health endpoint (200 OK)
- [x] WebSocket TCP connectivity
- [x] Nginx configuration valid
- [x] SSL auto-renewal configured

### ⏳ Pending Tests
- [ ] Web Dashboard access via HTTPS
- [ ] API Gateway endpoints
- [ ] WebSocket data flow
- [ ] Grafana dashboard access

### Test Commands
```bash
# Test HTTPS
curl -I https://mm.itziklerner.com/health

# Test redirect
curl -I http://mm.itziklerner.com

# Test WebSocket (requires websocat)
websocat wss://mm.itziklerner.com/ws?type=web

# Test SSL certificate
openssl s_client -connect mm.itziklerner.com:443 -servername mm.itziklerner.com
```

---

## 📋 Quick Reference

### Service Commands
```bash
# Nginx
sudo systemctl status nginx
sudo systemctl reload nginx
sudo systemctl restart nginx
sudo nginx -t

# View logs
sudo tail -f /var/log/nginx/b25-access.log
sudo tail -f /var/log/nginx/b25-error.log

# Certificate management
sudo /root/.acme.sh/acme.sh --list
sudo /root/.acme.sh/acme.sh --info -d mm.itziklerner.com --ecc
```

### Configuration Files
- Nginx config: `/etc/nginx/sites-available/mm.itziklerner.com`
- SSL cert: `/etc/nginx/ssl/mm.itziklerner.com.crt`
- SSL key: `/etc/nginx/ssl/mm.itziklerner.com.key`
- Web .env: `/home/mm/dev/b25/ui/web/.env`

### Important Paths
- acme.sh home: `/root/.acme.sh`
- Nginx sites: `/etc/nginx/sites-available/`
- Nginx logs: `/var/log/nginx/`
- Web root: `/var/www/html/`

---

## 🎯 Next Steps

1. **Start API Gateway** (port 8000)
2. **Build and deploy production web dashboard**
3. **Test all endpoints with HTTPS**
4. **Verify WebSocket data flow**
5. **Enable UFW firewall** (optional but recommended)
6. **Monitor SSL certificate auto-renewal**

---

## 📞 Support & Troubleshooting

### Common Issues

#### 1. 403 Forbidden on Web Dashboard
**Cause**: Vite dev server blocks non-localhost requests
**Solution**: Build for production or configure Vite to allow external access

#### 2. 502 Bad Gateway on API
**Cause**: API Gateway not running
**Solution**: Start the API Gateway service on port 8000

#### 3. WebSocket Connection Failed
**Cause**: Dashboard server not running or firewall blocking
**Solution**: Verify port 8086 is accessible and service is running

#### 4. Certificate Renewal Failed
**Cause**: Port 80 not accessible or DNS issues
**Solution**: Ensure Nginx is running and port 80 is open

### Debug Commands
```bash
# Check ports
sudo netstat -tlnp | grep -E "(80|443|3000|8000|8086)"

# Check Nginx config
sudo nginx -T

# Test local services
curl http://localhost:3000
curl http://localhost:8000
curl http://localhost:8086/health

# Check SSL
openssl s_client -connect mm.itziklerner.com:443 -servername mm.itziklerner.com < /dev/null
```

---

## ✨ Summary

**SSL Certificate Setup: COMPLETE** ✅

The B25 Trading System now has:
- ✅ Valid SSL certificate from ZeroSSL (expires Jan 4, 2026)
- ✅ Nginx reverse proxy with SSL termination
- ✅ HTTP to HTTPS redirect
- ✅ WebSocket support for real-time data
- ✅ Auto-renewal configured (daily checks)
- ✅ Security headers and best practices
- ✅ All domain URLs configured

**Production URLs**:
- https://mm.itziklerner.com (Web Dashboard)
- wss://mm.itziklerner.com/ws (WebSocket)
- https://mm.itziklerner.com/api (API Gateway)
- https://mm.itziklerner.com/grafana (Monitoring)

**Remaining Tasks**:
1. Start API Gateway service
2. Deploy production build of web dashboard
3. Test all HTTPS endpoints

---

*Last Updated: October 6, 2025*
