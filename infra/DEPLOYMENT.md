# NullMail Deployment Guide

## 🚀 Real-World Deployment Options

### 1. **Local Docker Production Test**

```bash
# 1. Build and run production stack
cd infra
cp ../.env.example .env
# Edit .env with your configuration
docker-compose -f docker-compose.prod.yml up --build

# 2. Test SMTP
swaks --to test@nullmail.local --from sender@example.com --server localhost:2525

# 3. Test Web Interface
open http://localhost:3000/inbox/test@nullmail.local
```

### 2. **VPS/Cloud Server Deployment**

#### **DigitalOcean/Linode/Vultr ($5-10/month)**

```bash
# 1. Create VPS with Docker
ssh root@your-server-ip

# 2. Install Docker & Docker Compose
curl -fsSL https://get.docker.com -o get-docker.sh
sh get-docker.sh
curl -L "https://github.com/docker/compose/releases/latest/download/docker-compose-$(uname -s)-$(uname -m)" -o /usr/local/bin/docker-compose
chmod +x /usr/local/bin/docker-compose

# 3. Clone and deploy
git clone https://github.com/yourusername/nullmail.git
cd nullmail/infra
cp ../.env.example .env
nano .env  # Configure your settings

# 4. Deploy with SSL
docker-compose -f docker-compose.prod.yml up -d

# 5. Set up SSL (Let's Encrypt)
./setup-ssl.sh your-domain.com
```

### 3. **Cloud Platform Deployments**

#### **Railway.app (Easiest)**
1. Connect GitHub repository
2. Deploy SMTP server + Client + Redis
3. Configure environment variables
4. Get domain: `yourapp.railway.app`

#### **Fly.io**
```bash
# Install flyctl
curl -L https://fly.io/install.sh | sh

# Deploy SMTP server
fly launch --dockerfile Dockerfile --name nullmail-smtp
fly deploy

# Deploy client
cd client
fly launch --dockerfile Dockerfile --name nullmail-web
fly deploy
```

#### **Railway Deployment Files**

Create `railway.toml`:
```toml
[build]
builder = "DOCKERFILE"
dockerfilePath = "Dockerfile"

[deploy]
startCommand = "./nullmail"
restartPolicyType = "ON_FAILURE"
restartPolicyMaxRetries = 10

[[services]]
name = "smtp"
[services.variables]
REDIS_URL = "${{ REDIS_URL }}"
PORT = "2525"
```

### 4. **Domain & DNS Setup**

#### **For Email Testing (MX Records)**
```dns
# Your DNS records
MX    @    10 mail.yourdomain.com
A     mail    your-server-ip
A     @       your-server-ip
```

#### **For Web Interface**
```dns
A     nullmail    your-server-ip
CNAME www         nullmail.yourdomain.com
```

### 5. **Production Checklist**

#### **Security**
- [ ] Change default Redis password
- [ ] Set up SSL certificates
- [ ] Configure firewall (ports 80, 443, 25, 2525)
- [ ] Enable fail2ban for SMTP protection

#### **Monitoring**
- [ ] Set up health checks
- [ ] Configure log aggregation
- [ ] Monitor Redis memory usage
- [ ] Set up uptime monitoring

#### **Performance**
- [ ] Configure Redis persistence
- [ ] Set up Nginx caching
- [ ] Configure rate limiting
- [ ] Set TTL policies for emails

### 6. **Testing Commands**

```bash
# Test SMTP from anywhere
swaks --to user@yourdomain.com --from test@gmail.com --server yourdomain.com:25

# Test internal SMTP
swaks --to user@nullmail.local --from test@example.com --server localhost:2525

# Test web interface
curl https://yourdomain.com/api/emails/test@nullmail.local

# Check logs
docker-compose logs -f smtp-server
docker-compose logs -f client
docker-compose logs -f redis
```

### 7. **Scaling for Production**

#### **High Availability Setup**
```yaml
# docker-compose.scale.yml
services:
  smtp-server:
    deploy:
      replicas: 3
      
  client:
    deploy:
      replicas: 2
      
  redis:
    image: redis/redis-stack:latest  # Redis Cluster
```

#### **Load Balancer Configuration**
```nginx
upstream smtp_servers {
    server smtp1:2525;
    server smtp2:2525;
    server smtp3:2525;
}

upstream web_servers {
    server client1:3000;
    server client2:3000;
}
```

### 8. **Cost Estimates**

| Platform | Monthly Cost | Features |
|----------|-------------|----------|
| Railway | $5-20 | Easy deploy, auto-scaling |
| DigitalOcean | $5-10 | Full control, custom domain |
| Fly.io | $0-15 | Global edge, fast |
| Linode | $5-10 | Reliable, good docs |
| Vultr | $2.50-10 | Cheap, many locations |

### 9. **Maintenance Scripts**

Create `scripts/cleanup.sh`:
```bash
#!/bin/bash
# Clean up old emails (run daily)
docker-compose exec redis redis-cli -a $REDIS_PASSWORD --eval cleanup.lua
```

Create `scripts/backup.sh`:
```bash
#!/bin/bash
# Backup Redis data
docker-compose exec redis redis-cli -a $REDIS_PASSWORD BGSAVE
docker cp nullmail-redis-prod:/data/dump.rdb ./backups/
```

### 10. **Quick Start Commands**

```bash
# Development
make docker-up
make run-client

# Production
cd infra
docker-compose -f docker-compose.prod.yml up -d

# Monitor
docker-compose logs -f

# Scale
docker-compose up --scale smtp-server=3 --scale client=2

# Update
git pull
docker-compose -f docker-compose.prod.yml up -d --build
```

## 🎯 **Recommended Approach**

1. **Start with Railway.app** - Easiest deployment
2. **Test thoroughly** with real email providers
3. **Move to VPS** when you need custom domains
4. **Scale horizontally** as usage grows

The system is now production-ready! 🚀