# Analytics Service - Deployment Checklist

## Pre-Deployment

### Infrastructure
- [ ] PostgreSQL 14+ database provisioned
- [ ] Redis 7+ instance available
- [ ] Kafka cluster configured (or alternative message queue)
- [ ] Network connectivity between services verified
- [ ] Firewall rules configured for ports 9097, 9098, 9099

### Configuration
- [ ] `config.yaml` created from `config.example.yaml`
- [ ] Database credentials configured
- [ ] Redis connection details set
- [ ] Kafka broker addresses configured
- [ ] Log level set appropriately (info for production)
- [ ] Retention policies configured for your needs
- [ ] CORS allowed origins configured
- [ ] Rate limiting enabled and configured

### Database Setup
- [ ] Database created: `CREATE DATABASE analytics;`
- [ ] User created with proper permissions
- [ ] Migrations executed: `migrations/001_initial_schema.sql`
- [ ] Indexes verified
- [ ] Partitions configured (or auto-partition script set up)
- [ ] Database backup strategy in place

### Security
- [ ] Database SSL/TLS enabled
- [ ] Redis password set (if applicable)
- [ ] API rate limiting configured
- [ ] Secrets stored securely (not in config files)
- [ ] Container runs as non-root user (default: UID 1000)
- [ ] Network policies configured (if using Kubernetes)

## Deployment Steps

### Docker Deployment
- [ ] Docker image built: `make docker-build`
- [ ] Image tagged with version
- [ ] Image pushed to registry (if using private registry)
- [ ] Volume mounts configured for config file
- [ ] Environment variables set
- [ ] Container started: `docker run` or `docker-compose up`

### Kubernetes Deployment
- [ ] ConfigMap created for configuration
- [ ] Secrets created for sensitive data
- [ ] PersistentVolumeClaim created (if needed)
- [ ] Deployment manifest applied
- [ ] Service manifest applied (ClusterIP/LoadBalancer)
- [ ] Ingress configured (if external access needed)
- [ ] HorizontalPodAutoscaler configured (optional)

### Kafka Topics
- [ ] Topics created: `./scripts/init-topics.sh`
  - [ ] trading.events
  - [ ] market.data
  - [ ] order.events
  - [ ] user.actions
- [ ] Topic partitions configured appropriately
- [ ] Retention policies set on topics
- [ ] Consumer group configured

## Post-Deployment Verification

### Health Checks
- [ ] Service is running: `curl http://localhost:9097/health`
- [ ] Database connected: Check health endpoint response
- [ ] Redis connected: Check health endpoint response
- [ ] Kafka consumer started: Check logs for "Event consumer started"

### Functional Tests
- [ ] Send test event: `./scripts/test-event.sh`
- [ ] Query events: `curl http://localhost:9097/api/v1/events`
- [ ] Check dashboard metrics: `curl http://localhost:9097/api/v1/dashboard/metrics`
- [ ] Verify Prometheus metrics: `curl http://localhost:9098/metrics`

### Performance Tests
- [ ] Send batch of 1000 events, verify ingestion
- [ ] Check database for inserted events
- [ ] Verify aggregations are running
- [ ] Check Redis cache is working (cache hit metrics)
- [ ] Monitor resource usage (CPU, memory)

### Monitoring Setup
- [ ] Prometheus scraping configured
- [ ] Grafana dashboards imported/created
- [ ] Alert rules configured
- [ ] Log aggregation configured (ELK, Loki, etc.)
- [ ] Error tracking configured (Sentry, etc.)

### Metrics to Monitor
- [ ] `analytics_events_ingested_total` - Should be increasing
- [ ] `analytics_events_failed_total` - Should be near zero
- [ ] `analytics_batches_processed_total` - Should be increasing
- [ ] `analytics_cache_hits_total` - Should increase over time
- [ ] Database connection pool usage
- [ ] API response times
- [ ] Memory and CPU usage

## Production Configuration

### Resource Limits (Kubernetes)
```yaml
resources:
  requests:
    memory: "256Mi"
    cpu: "250m"
  limits:
    memory: "512Mi"
    cpu: "500m"
```

### Scaling Configuration
- [ ] Horizontal scaling strategy defined
- [ ] Load balancer configured (if multiple instances)
- [ ] Kafka consumer group properly configured for scaling
- [ ] Database connection pool sized appropriately

### Backup & Recovery
- [ ] Database backup schedule configured (daily recommended)
- [ ] Point-in-time recovery tested
- [ ] Disaster recovery plan documented
- [ ] Redis persistence enabled (if needed)

### Logging
- [ ] Log level: `info` (not `debug` in production)
- [ ] Log format: `json` for structured logging
- [ ] Log aggregation destination configured
- [ ] Log retention policy set

## Operational Procedures

### Daily Tasks
- [ ] Check service health
- [ ] Review error logs
- [ ] Monitor resource usage
- [ ] Verify data ingestion

### Weekly Tasks
- [ ] Review performance metrics
- [ ] Check database partition status
- [ ] Verify backup completion
- [ ] Review and address any alerts

### Monthly Tasks
- [ ] Review data retention and cleanup
- [ ] Analyze query performance
- [ ] Update dependencies (security patches)
- [ ] Review and optimize database indexes

### Emergency Procedures
- [ ] Incident response plan documented
- [ ] On-call rotation configured
- [ ] Rollback procedure documented
- [ ] Emergency contact list updated

## Rollback Plan

If deployment fails:
1. [ ] Stop new service
2. [ ] Restore previous version
3. [ ] Verify database integrity
4. [ ] Check for any data loss
5. [ ] Document issues for future deployment

## Sign-off

- [ ] All checklist items completed
- [ ] Deployment documented
- [ ] Team notified of deployment
- [ ] Monitoring verified
- [ ] Production deployment approved

**Deployed By**: _______________
**Date**: _______________
**Version**: _______________
**Environment**: _______________

## Support Contacts

- **On-call Engineer**: _______________
- **Database Admin**: _______________
- **DevOps Team**: _______________
- **Documentation**: See README.md and IMPLEMENTATION_SUMMARY.md

---

For detailed information, see:
- [README.md](README.md) - Full documentation
- [QUICK_START.md](QUICK_START.md) - Quick start guide
- [IMPLEMENTATION_SUMMARY.md](IMPLEMENTATION_SUMMARY.md) - Technical details
