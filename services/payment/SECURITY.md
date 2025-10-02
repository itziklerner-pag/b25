# Security and PCI Compliance Guide

## Overview

This document outlines the security measures and PCI compliance considerations for the Payment Service.

## PCI DSS Compliance

### Scope Reduction Strategy

The payment service is designed to minimize PCI DSS scope by:

1. **Never Storing Card Data**: All card information is handled by Stripe
2. **Using Stripe Elements**: Card data is collected directly by Stripe's PCI-compliant forms
3. **Token-Based Processing**: Only payment method tokens are stored

### SAQ A Compliance

This service is designed for **SAQ A** (Self-Assessment Questionnaire A) compliance when:

- Using Stripe Elements for all card data collection
- Never transmitting cardholder data through your servers
- Using HTTPS for all communications
- Not storing, processing, or transmitting cardholder data

### Security Controls

#### 1. Data Protection

**Encryption at Rest**
- Sensitive database fields encrypted using AES-256
- Encryption keys stored separately from data
- 32-character encryption key required

**Encryption in Transit**
- TLS 1.3 enforced for all connections
- Certificate validation required
- Secure cipher suites only

**Data Retention**
- Payment tokens stored indefinitely (no card data)
- Transaction records retained per business requirements
- Webhook events retained for audit purposes

#### 2. Access Control

**Authentication**
- JWT-based authentication required for all API endpoints
- Token expiration enforced
- Refresh token rotation supported

**Authorization**
- User-level access control
- Users can only access their own data
- Admin endpoints protected separately

**Rate Limiting**
- 100 requests per minute per IP (configurable)
- Webhook endpoints have separate limits
- DDoS protection via rate limiting

#### 3. Secure Development Practices

**Input Validation**
- All inputs validated using struct tags
- SQL injection prevention via parameterized queries
- XSS prevention via proper encoding

**Error Handling**
- No sensitive data in error messages
- Structured error logging
- Error tracking for audit purposes

**Dependency Management**
- Regular dependency updates
- Vulnerability scanning
- Minimal dependency footprint

#### 4. Audit and Monitoring

**Logging**
- All payment operations logged
- Webhook events stored for replay
- Failed authentication attempts logged
- No card data or CVV in logs

**Monitoring**
- Prometheus metrics for all operations
- Alert on failed payments
- Alert on webhook processing failures
- Performance monitoring

## Security Best Practices

### 1. Environment Configuration

**Required Secrets**
```env
# Strong JWT secret (minimum 32 characters)
JWT_SECRET=your_very_strong_jwt_secret_here_32_chars_minimum

# Encryption key (exactly 32 characters)
ENCRYPTION_KEY=12345678901234567890123456789012

# Stripe keys (never commit to version control)
STRIPE_SECRET_KEY=sk_live_...
STRIPE_WEBHOOK_SECRET=whsec_...
```

**Key Management**
- Use environment variables for secrets
- Never commit secrets to version control
- Rotate keys regularly (every 90 days)
- Use different keys per environment

### 2. Database Security

**Connection Security**
```
✓ Use SSL/TLS for database connections
✓ Configure firewall rules
✓ Use strong passwords
✓ Enable database auditing
✓ Regular backups with encryption
```

**Access Control**
```
✓ Least privilege principle
✓ Separate read/write users if possible
✓ No direct database access for applications
✓ Use connection pooling
```

### 3. API Security

**Request Validation**
- Validate content type
- Enforce size limits
- Validate JSON structure
- Sanitize inputs

**Response Security**
- Never expose internal errors
- Use consistent error responses
- Implement CORS properly
- Set security headers

### 4. Webhook Security

**Verification**
```go
// Always verify webhook signatures
signature := c.GetHeader("Stripe-Signature")
event, err := webhook.ConstructEvent(payload, signature, webhookSecret)
```

**Idempotency**
- Store webhook events to prevent duplicates
- Check event ID before processing
- Handle replay attacks

### 5. Production Deployment

**Infrastructure Security**
```
✓ Use HTTPS only
✓ Enable WAF (Web Application Firewall)
✓ Configure security groups/firewall rules
✓ Use private subnets for databases
✓ Enable VPC peering if needed
✓ Implement DDoS protection
```

**Application Security**
```
✓ Run as non-root user
✓ Minimize container attack surface
✓ Use security scanning in CI/CD
✓ Enable health checks
✓ Implement graceful shutdown
```

## Incident Response

### Security Incident Procedure

1. **Detection**
   - Monitor logs for anomalies
   - Set up alerts for suspicious activity
   - Review webhook failures

2. **Response**
   - Isolate affected systems
   - Preserve evidence (logs, database snapshots)
   - Notify stakeholders

3. **Recovery**
   - Patch vulnerabilities
   - Rotate compromised credentials
   - Restore from clean backups if needed

4. **Post-Incident**
   - Document incident
   - Update security procedures
   - Implement preventive measures

### Common Security Events

**Failed Authentication Attempts**
```
Action: Monitor and alert on repeated failures
Threshold: 5 failures in 5 minutes
Response: Temporary IP block
```

**Unusual Payment Patterns**
```
Action: Monitor for anomalies
Examples: Large amounts, rapid succession, foreign countries
Response: Additional verification
```

**Webhook Signature Failures**
```
Action: Alert immediately
Possible Cause: Man-in-the-middle attack
Response: Investigate and verify Stripe configuration
```

## Data Privacy

### GDPR Compliance

**Right to Access**
- API endpoints to retrieve user data
- Export functionality for user data

**Right to Deletion**
- Hard delete user payment methods
- Soft delete transaction records (retain for accounting)
- Anonymize user data while keeping transactions

**Data Minimization**
- Only store necessary payment data
- Use Stripe tokens instead of raw card data
- Regular data cleanup

### Data Handling

**Personal Information**
- User IDs (pseudonymized)
- Email addresses (if stored)
- Payment method metadata

**Sensitive Information**
- Payment method tokens (encrypted)
- Transaction amounts
- Billing addresses (if collected)

**Prohibited Information**
- Full card numbers
- CVV/CVC codes
- PIN numbers
- Track data

## Security Checklist

### Development

- [ ] No hard-coded secrets
- [ ] Input validation on all endpoints
- [ ] SQL injection prevention
- [ ] XSS prevention
- [ ] CSRF protection (if using sessions)
- [ ] Dependency vulnerability scanning
- [ ] Code review for security issues

### Deployment

- [ ] HTTPS/TLS enabled
- [ ] Strong JWT secret configured
- [ ] Encryption key set (32 characters)
- [ ] Database credentials secured
- [ ] Stripe webhook secret configured
- [ ] Rate limiting enabled
- [ ] CORS configured properly
- [ ] Security headers set
- [ ] Health checks enabled

### Operations

- [ ] Monitoring and alerting configured
- [ ] Log aggregation set up
- [ ] Backup strategy implemented
- [ ] Incident response plan documented
- [ ] Key rotation schedule defined
- [ ] Access control reviewed
- [ ] Security audit scheduled
- [ ] Penetration testing planned

## Compliance Documentation

### Required Documentation

1. **System Architecture Diagram**
   - Network topology
   - Data flow diagrams
   - Security boundaries

2. **Data Flow Documentation**
   - How payment data moves through system
   - Where data is stored
   - Data retention policies

3. **Access Control Policies**
   - Who has access to what
   - Authentication mechanisms
   - Authorization rules

4. **Incident Response Plan**
   - Detection procedures
   - Response procedures
   - Communication plan

5. **Security Policies**
   - Password policies
   - Encryption policies
   - Key management policies
   - Data retention policies

## Third-Party Security

### Stripe Integration

- Use latest Stripe SDK version
- Verify webhook signatures always
- Handle webhook events idempotently
- Monitor Stripe status page
- Review Stripe security updates

### Dependencies

- Regular dependency updates
- Automated vulnerability scanning
- Security patches applied promptly
- Minimal dependency footprint

## Contact

For security concerns or to report vulnerabilities:
- Email: security@yourcompany.com
- PGP Key: [Your PGP Key]

---

**Last Updated**: 2025-10-03
**Review Schedule**: Quarterly
