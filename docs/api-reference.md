# WeChat Service Account API Reference

## Table of Contents

1. [API Overview](#1-api-overview)
2. [Authentication & Credentials](#2-authentication--credentials)
3. [AccessToken Management](#3-accesstoken-management)
4. [API Domains & Endpoints](#4-api-domains--endpoints)
5. [Rate Limits](#5-rate-limits)
6. [Error Handling](#6-error-handling)
7. [Monitoring & Alerts](#7-monitoring--alerts)
8. [Troubleshooting](#8-troubleshooting)

---

## 1. API Overview

### Base Information

- **API Base URL:** `https://api.weixin.qq.com/`
- **Protocol:** HTTPS (required)
- **Data Format:** JSON/XML depending on endpoint
- **Encoding:** UTF-8

### API Categories

| Category | Description |
|----------|-------------|
| Base APIs | AccessToken, server IP retrieval |
| Message APIs | Sending/receiving messages |
| User APIs | User management, tags, notes |
| Menu APIs | Custom menu creation/management |
| Material APIs | Media upload/download |
| Group APIs | User grouping |
| QR Code APIs | Temporary/permanent QR codes |

---

## 2. Authentication & Credentials

### 2.1 Required Credentials

| Credential | Description | Location |
|------------|-------------|----------|
| **AppID** | Service account unique identifier | WeChat Developer Platform → Service Account → Basic Info |
| **AppSecret** | Secret key for API authentication | WeChat Developer Platform → Service Account → Development Keys |
| **Token** | Server verification token | WeChat Developer Platform → Service Account → Development Information |
| **EncodingAESKey** | Message encryption key (optional) | WeChat Developer Platform → Service Account → Development Information |

### 2.2 AppSecret Management

**Operations Supported:**
- Enable/Disable
- Reset (regenerate new secret)
- Freeze/Unfreeze (takes 10 minutes to take effect)

**Security Recommendations:**
- Freeze AppSecret if not in use for extended periods
- Platform does not save AppSecret - if lost, must reset
- Use WeChat Cloud Hosting for "免鉴权调用" (authentication-free calls)

**Error After Freezing:**
```
Error Code: 40243
Message: "AppSecret is frozen"
```

### 2.3 IP Whitelist

- Only IPs in whitelist can call AccessToken APIs
- Configure at: WeChat Developer Platform → Service Account → Development Keys
- **Error if IP not in whitelist:** `61004`

### 2.4 Security Features

**Risk-Based Access Control:**
- High-risk operations trigger confirmation flow
- Process:
  1. Developer initiates call → Returns error `89503`
  2. WeChat sends template message to admin
  3. Admin confirms IP → Call succeeds
- If admin rejects: IP blocked for 1 hour
- Multiple rejections → Long-term blocking

**Recommendation:** Communicate with admin before calls or enable IP whitelist.

---

## 3. AccessToken Management

### 3.1 Overview

AccessToken is required for all API calls.

**Properties:**
- **Type:** String (512+ characters)
- **Validity:** 2 hours (7200 seconds)
- **Scope:** Service account global unique credential
- **Renewal:** Must be proactively refreshed

### 3.2 Token Endpoints

| Endpoint | Type | Description |
|----------|------|-------------|
| `/cgi-bin/token` | Standard | Regular AccessToken |
| `/cgi-bin/stable/access_token` | Stable | More stable, recommended |

### 3.3 Token Request Example

```bash
curl -X POST "https://api.weixin.qq.com/cgi-bin/token" \
  -d '{
    "grant_type": "client_credential",
    "appid": "YOUR_APPID",
    "secret": "YOUR_APPSECRET"
  }'
```

**Response:**
```json
{
  "access_token": "ACCESS_TOKEN",
  "expires_in": 7200
}
```

### 3.4 Centralized Token Server Architecture

**Critical:** Use a central token management server.

```
┌─────────────────────────────────────────────────────────┐
│              Centralized Token Architecture             │
├─────────────────────────────────────────────────────────┤
│                                                          │
│  ┌──────────────────┐                                   │
│  │  Token Server    │                                   │
│  │  - Proactive refresh (every ~1 hour)                │
│  │  - Reactive refresh on error                        │
│  │  - Concurrency control (locks)                      │
│  │  - Storage (Redis/MySQL)                            │
│  └────────┬─────────┘                                   │
│           │                                             │
│           │ GetToken()                                  │
│           ▼                                             │
│  ┌─────────────────────────────────────────────┐      │
│  │  Business Servers                           │      │
│  │  - Request token from central server        │      │
│  │  - Cache in memory (1 min)                  │      │
│  │  - Call WeChat APIs                         │      │
│  └─────────────────────────────────────────────┘      │
│                                                          │
└─────────────────────────────────────────────────────────┘
```

### 3.5 Best Practices

| Practice | Description |
|----------|-------------|
| **Proactive Refresh** | Refresh 1 hour before expiration |
| **Dual Token Period** | During refresh, old token still valid for 5 min |
| **Passive Refresh** | API returns error → trigger refresh |
| **Concurrent Control** | Use locks to prevent race conditions |
| **Storage** | Minimum 512 characters storage space |

### 3.6 Token Flow Diagram

```
1. Request Token
   ┌─────────────┐     ┌─────────────┐     ┌─────────────┐
   │ App Server  │────▶│ Token Server│────▶│ WeChat API  │
   └─────────────┘     └─────────────┘     └─────────────┘
                              │
                              ▼ Store
                         ┌─────────────┐
                         │ Redis/MySQL │
                         └─────────────┘

2. API Call with Token
   ┌─────────────┐     ┌─────────────┐     ┌─────────────┐
   │ App Server  │────▶│ WeChat API  │     │ Response    │
   │ (with token)│◀────│             │────▶│             │
   └─────────────┘     └─────────────┘     └─────────────┘

3. Token Expired (Error 42001)
   ┌─────────────┐     ┌─────────────┐     ┌─────────────┐
   │ App Server  │────▶│ WeChat API  │     │ Error       │
   │ (old token) │◀────│             │     │ 42001       │
   └─────────────┘     └─────────────┘           │
                                                  ▼ Trigger Refresh
                                             ┌─────────────┐
                                             │ Token Server│
                                             │ (re-refresh)│
                                             └─────────────┘
```

---

## 4. API Domains & Endpoints

### 4.1 Domain Options

| Domain | Purpose | Use Case |
|--------|---------|----------|
| `api.weixin.qq.com` | General | Default, auto-routed to nearest point |
| `api2.weixin.qq.com` | Disaster Recovery | Fallback when general domain unavailable |
| `sh.api.weixin.qq.com` | Shanghai | Direct access to Shanghai node |
| `sz.api.weixin.qq.com` | Shenzhen | Direct access to Shenzhen node |
| `hk.api.weixin.qq.com` | Hong Kong | Direct access to Hong Kong node |

### 4.2 Usage Recommendations

1. Use `api.weixin.qq.com` by default (auto-routing)
2. Configure `api2.weixin.qq.com` as backup
3. Use regional domains for lower latency if you know your server location
4. **Always use domain names, not IP addresses**

### 4.3 Getting WeChat Server IPs

For firewall whitelisting:

```bash
curl "https://api.weixin.qq.com/cgi-bin/get_api_domain_ip"
```

**Response:**
```json
{
  "ip_list": ["101.226.103.0/25", "140.207.69.0/25", ...]
}
```

### 4.4 Common API Endpoints

#### Base APIs

| Endpoint | Method | Description |
|----------|--------|-------------|
| `/cgi-bin/token` | POST | Get AccessToken |
| `/cgi-bin/stable/access_token` | POST | Get Stable AccessToken |
| `/cgi-bin/get_api_domain_ip` | GET | Get API Server IPs |
| `/cgi-bin/getcallbackip` | GET | Get Callback Server IPs |

#### Message APIs

| Endpoint | Method | Description |
|----------|--------|-------------|
| `/cgi-bin/message/custom/send` | POST | Send客服 message |
| `/cgi-bin/message/mass/sendall` | POST | Mass broadcast |
| `/cgi-bin/message/mass/send` | POST | Mass send by openids |

#### User APIs

| Endpoint | Method | Description |
|----------|--------|-------------|
| `/cgi-bin/user/info` | GET | Get user info |
| `/cgi-bin/user/get` | GET | Get user list |
| `/cgi-bin/tags/create` | POST | Create tag |
| `/cgi-bin/tags/get` | GET | Get tags |
| `/cgi-bin/tags/members/batch tagging` | POST | Tag users |
| `/cgi-bin/user/info/updateremark` | POST | Set user remark |

#### Menu APIs

| Endpoint | Method | Description |
|----------|--------|-------------|
| `/cgi-bin/menu/create` | POST | Create menu |
| `/cgi-bin/menu/get` | GET | Get menu |
| `/cgi-bin/menu/delete` | DELETE | Delete menu |

#### Material APIs

| Endpoint | Method | Description |
|----------|--------|-------------|
| `/cgi-bin/material/add_material` | POST | Upload material |
| `/cgi-bin/material/get_material` | POST | Get material |
| `/cgi-bin/material/del_material` | POST | Delete material |
| `/cgi-bin/material/batchget_material` | POST | List materials |

---

## 5. Rate Limits

### 5.1 Daily Limits (New Accounts)

| API | Daily Limit |
|-----|-------------|
| Get AccessToken | 2,000 |
| Create Menu | 1,000 |
| Query Menu | 10,000 |
| Delete Menu | 1,000 |
| Create Tag | 1,000 |
| Query Tags | 1,000 |
| Update Tag Name | 1,000 |
| Move User to Tag | 100,000 |
| Upload Media | 100,000 |
| Download Media | 200,000 |
| Send Customer Message | 500,000 |
| Mass Broadcast | 100 |
| Upload News | 10 |
| Delete News | 10 |
| Get QR Code | 100,000 |
| Get Follower List | 500 |
| Get User Info | 5,000,000 |
| Set User Remark | 10,000 |

### 5.2 Draft Box APIs

| API | Daily Limit |
|-----|-------------|
| Create Draft | 1,000 |
| Get Draft | 500 |
| Delete Draft | 1,000 |
| Update Draft | 1,000 |
| Get Draft Count | 1,000 |
| Get Draft List | 1,000 |

### 5.3 Publishing APIs

| API | Daily Limit |
|-----|-------------|
| Publish Article | 100 |
| Query Publish Status | 100 |
| Delete Published | 10 |
| Get Published Article | 100 |
| Get Published List | 100 |

### 5.4 Test Account Limits

| API | Daily Limit |
|-----|-------------|
| Get AccessToken | 200 |
| Create Menu | 100 |
| Query Menu | 1,000 |
| Delete Menu | 100 |
| Create Tag | 100 |
| Query Tags | 100 |
| Update Tag Name | 100 |
| Move User to Tag | 1,000 |
| Upload Temp Media | 500 |
| Download Temp Media | 1,000 |
| Send Customer Message | 50,000 |
| Get QR Code | 10,000 |
| Get Follower List | 100 |
| Get User Info | 500,000 |

### 5.5 Limits Increase

When follower count reaches thresholds, limits increase:
- **100,000 followers:** Higher limits
- **1,000,000 followers:** Even higher limits
- Check actual limits at: WeChat Admin → Developer Center → API Quotas

### 5.6 Quota Management

**Reset APIs:**
- `POST /cgi-bin/clear_quota` - Reset all API calls
- `POST /cgi-bin/clear_quota/v2` - Reset specified API
- `POST /cgi-bin/clear_quota/v3` - Reset using AppSecret

**Important:**
- Each account gets 10 reset operations per month
- Includes both admin panel and API resets
- Real-time data may have ~1% error margin

---

## 6. Error Handling

### 6.1 Common Error Codes

| Code | Meaning | Action |
|------|---------|--------|
| 0 | Success | Continue |
| 40001 | Invalid credential (AccessToken) | Refresh token |
| 40002 | Invalid grant_type | Check parameter |
| 40013 | Invalid AppID | Verify AppID |
| 40125 | Invalid AppSecret | Reset AppSecret |
| 40243 | AppSecret frozen | Unfreeze in admin |
| 42001 | AccessToken expired | Refresh token |
| 42002 | RefreshToken expired | Re-authorize |
| 42003 | Code expired | Re-authenticate |
| 43002 | POST method required | Use POST |
| 44002 | Empty POST data | Check request body |
| 44003 | Empty media data | Check media file |
| 45009 | API quota exceeded | Wait or reset |
| 48001 | API not authorized | Check permissions |
| 50005 | User not following | Cannot send message |
| 61004 | IP not in whitelist | Add IP to whitelist |
| 89503 | Risk control trigger | Admin confirmation |
| 40004 | Invalid media type | Check file type |

### 6.2 Error Response Format

```json
{
  "errcode": 40001,
  "errmsg": "invalid credential, access_token is invalid or not latest"
}
```

### 6.3 Error Handling Strategy

```go
func callWeChatAPI(endpoint string, data map[string]interface{}) (*Response, error) {
    token, err := getValidToken()
    if err != nil {
        return nil, err
    }
    
    resp, err := http.Post(endpoint+"?access_token="+token, "application/json", body)
    if err != nil {
        return nil, err
    }
    
    var result Response
    json.Unmarshal(resp.Body, &result)
    
    switch result.ErrCode {
    case 0:
        return &result, nil
    case 40001, 42001:
        // Token invalid/expired, refresh and retry
        token = refreshToken()
        return callWeChatAPI(endpoint, data)  // Retry
    case 45009:
        // Quota exceeded, wait or return error
        return nil, errors.New("quota exceeded")
    default:
        return nil, errors.New(result.ErrMsg)
    }
}
```

### 6.4 Retry Strategy

| Error Code | Retry? | Action |
|------------|--------|--------|
| 40001 | Yes | Refresh token, retry once |
| 42001 | Yes | Refresh token, retry once |
| 45009 | No | Wait 24 hours or reset quota |
| 48001 | No | Check permissions |
| 50005 | No | User unfollowed |
| Others | No | Check error message |

---

## 7. Monitoring & Alerts

### 7.1 Alert Configuration

Configure at: **WeChat Admin → Development → Operations Center → API Alerts**

Alerts are sent to a WeChat group for real-time monitoring.

### 7.2 Alert Types

#### Universal Alerts (All Developers)

| Alert Type | Description | Severity |
|------------|-------------|----------|
| DNS Failure | Cannot resolve domain | High |
| DNS Timeout | DNS resolution > 5s | Medium |
| Connection Timeout | Cannot connect > 3s | High |
| Request Timeout | No response > 5s | High |
| Response Invalid | Response malformed | High |
| MarkFail | Auto-blocked (1 min) | Critical |

#### Third-Party Platform Alerts

| Alert Type | Description |
|------------|-------------|
| Component Verify Ticket Timeout | > 5s response |
| Component Verify Ticket Failed | No "success" response |
| Third-Party Message Timeout | > 5s response |
| Third-Party Message Failed | No "success" response |

### 7.3 Alert Content

Each alert includes:
- `appid` - Service account AppID
- `nickname` - Service account name
- `time` - First occurrence timestamp
- `description` - Error description
- `count` - Failure count
- `example` - Sample error details (IP, message type, response)

### 7.4 Alert Examples

**Request Timeout Alert:**
```
Time: 2014-12-01 20:12:00
IP: 203.205.140.29
Event: Unsubscribe event
Count: 1272 failures in 5 minutes
```

**Response Invalid Alert:**
```
Time: 2014-12-01 20:12:00
IP: 58.248.9.218
Event: Menu click
Response: "Error 500:"
Length: 10 bytes
Count: 1320 failures
```

---

## 8. Troubleshooting

### 8.1 DNS Issues

**Symptoms:** DNS failure alerts

**排查步骤:**
1. Verify URL/domain in admin panel
2. Check domain hasn't expired or changed
3. Test DNS resolution:
   ```bash
   # Using Tencent DNS
   dig @182.254.116.116 your-domain.com
   
   # Windows - change DNS to 182.254.116.116
   ping your-domain.com
   ```
4. If still failing, contact WeChat support

### 8.2 Connection Timeout

**Symptoms:** Connection timeout alerts, MarkFail

**排查步骤:**
1. Verify server IP in alert is correct
2. Check server load: `uptime`, `top`
3. Check network connectivity:
   ```bash
   # Get WeChat callback IPs
   curl "https://api.weixin.qq.com/cgi-bin/getcallbackip"
   
   # Ping test
   ping <wechat_ip>
   ```
4. Check firewall settings
5. Review nginx logs:
   ```
   logs/access.log: Check upstream_status
   logs/error.log: Check "Connection reset", "Timeout"
   ```
6. Check system limits:
   ```bash
   ulimit -n  # File descriptors
   netstat -an | grep ESTABLISHED | wc -l  # Connections
   ```

**Solutions:**
- Optimize performance, scale horizontally
- Implement async processing
- Return "success" immediately, process later

### 8.3 Request Timeout (5s Limit)

**Symptoms:** No response within 5 seconds

**排查步骤:**
1. Identify slow request in logs
2. Check server metrics:
   ```bash
   # CPU load
   uptime
   vmstat 5 5
   top
   
   # Memory
   free -m
   
   # Disk I/O
   iostat -x 1
   
   # Network
   netstat -i
   sar -n DEV
   ```
3. Check nginx configuration
4. Review application logs

**Solutions:**
- Optimize slow queries
- Add caching
- Scale vertically/horizontally
- Async processing pattern:
  ```go
  // Receive message, return "success" immediately
  func handleMessage(w http.ResponseWriter, r *http.Request) {
      // Parse message
      go processMessageAsync(message)  // Process in background
      w.Write([]byte("success"))       // Return immediately
  }
  ```

### 8.4 Response Invalid

**Symptoms:** Response format errors

**排查步骤:**
1. Check alert content for response body
2. Verify response format matches API docs
3. Check for unexpected exceptions

**Common Issues:**
- Missing required XML fields
- Invalid XML/JSON encoding
- Server errors (500) in response
- Empty responses

### 8.5 MarkFail (Auto-Blocking)

**Symptoms:** MarkFail alerts, no messages delivered

**Cause:** Multiple consecutive failures trigger auto-blocking

**Solution:**
1. This is CRITICAL - service is down
2. Check previous alerts (timeout, connection, response)
3. Fix underlying issue
4. Service auto-recovers after 1 minute
5. Implement proper error handling to prevent recurrence

### 8.6 Monitoring Tools

| Tool | Purpose |
|------|---------|
| `uptime` | System load average |
| `vmstat` | Virtual memory statistics |
| `top` | Process and CPU monitoring |
| `free` | Memory usage |
| `netstat` | Network connections |
| `sar` | System activity reporter |
| `iostat` | Disk I/O statistics |
| `nginx` | Check access.log, error.log |

### 8.7 Nginx Configuration Checklist

```nginx
# Worker configuration
worker_processes auto;
worker_rlimit_nofile 65535;

events {
    worker_connections 20480;
    use epoll;
    multi_accept on;
}

http {
    keepalive_timeout 65;
    client_max_body_size 10m;
    
    # Logging
    access_log logs/access.log;
    error_log logs/error.log warn;
}
```

### 8.8 Best Practices Checklist

- [ ] Use centralized AccessToken server
- [ ] Implement proper error handling
- [ ] Add comprehensive logging at each layer
- [ ] Set up monitoring and alerts
- [ ] Configure IP whitelist
- [ ] Implement async processing for slow operations
- [ ] Return "success" immediately for quick responses
- [ ] Scale horizontally for high load
- [ ] Regular performance testing
- [ ] Document common issues and solutions

---

## Quick Reference

### Essential Commands

```bash
# Get AccessToken
curl -X POST "https://api.weixin.qq.com/cgi-bin/token" \
  -d "grant_type=client_credential&appid=APPID&secret=APPSECRET"

# Get API Server IPs
curl "https://api.weixin.qq.com/cgi-bin/get_api_domain_ip"

# Reset Quota
curl -X POST "https://api.weixin.qq.com/cgi-bin/clear_quota" \
  -d '{"appid":"APPID"}'

# Check Server Status
curl -I "https://your-server.com/wechat"
```

### Key Numbers

| Metric | Value |
|--------|-------|
| AccessToken validity | 7200 seconds (2 hours) |
| Request timeout | 5 seconds |
| Connection timeout | 3 seconds |
| DNS timeout | 5 seconds |
| Auto-block duration | 1 minute |
| Monthly quota resets | 10 |
| Token refresh grace period | 5 minutes |

### Emergency Contacts

- **Community Forum:** https://developers.weixin.qq.com/community/
- **API Debug Tool:** https://developers.weixin.qq.com/console/devtools/debug
- **WeChat Support:** Via admin panel

---

## References

- [Official API Documentation](https://developers.weixin.qq.com/doc/service/dev/api/)
- [Rate Limit Details](https://developers.weixin.qq.com/doc/service/guide/dev/api/limit.html)
- [Alert Guide](https://developers.weixin.qq.com/doc/service/guide/dev/api/warn_guide.html)
- [Error Code Reference](https://developers.weixin.qq.com/doc/oplatform/developers/errCode/errCode.html)
- [WeChat Developer Platform](https://developers.weixin.qq.com/platform/)
