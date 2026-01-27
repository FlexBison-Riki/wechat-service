# WeChat Platform Integration Details

## 1. Service Account (服务号) Overview

Service Accounts provide enterprises and organizations with:
- **Business Services** - Advanced customer service capabilities
- **User Management** - Comprehensive user relationship management
- **Quick Service Platform** - Fast implementation of service platforms within WeChat

### Key Capabilities

| Capability | Description |
|------------|-------------|
| Message API | Send/receive messages, custom menus |
| OAuth2 | Web authorization for user identity |
| Payment | WeChat Pay integration |
| Media Management | Image, video, voice素材 management |
| User Management | User tags, attributes, grouping |
| Data Analytics | User behavior and statistics |
| Customer Service | Multi-agent customer service |

## 2. Integration Steps

### Step 1: Account Setup

1. Register Service Account at [WeChat Official Platform](https://mp.weixin.qq.com/)
2. Apply for Developer permission at [WeChat Developer Platform](https://developers.weixin.qq.com/platform/)
3. Complete verification and configuration

### Step 2: Server Configuration

Configure your server to:
- **Verify Server** - Validate token with WeChat
- **Handle Messages** - Process incoming messages/events
- **Respond** - Send appropriate responses

**Required Endpoints:**
```
GET  /wechat?signature=...&timestamp=...&nonce=...&echostr=...
POST /wechat (for message handling)
```

### Step 3: API Integration

#### Access Token Management

The `access_token` is required for most API calls:

```go
// Using the SDK
token, err := officialAccount.GetAccessToken()
```

**Important:**
- Token expires every 2 hours
- Must be cached and refreshed proactively
- Rate limits apply per account

#### Message Handling

```go
server := officialAccount.GetServer(req, writer)
server.SetMessageHandler(func(msg *message.MixMessage) *message.Reply {
    text := message.NewText(msg.Content)
    return &message.Reply{MsgType: message.MsgTypeText, MsgData: text}
})
server.Serve()
server.Send()
```

### Step 4: OAuth2 Web Authorization

For web-based applications:

```go
oauth := officialAccount.GetOauth()
// Get user info via OAuth2
userInfo, err := oauth.GetUserInfo(code)
```

## 3. API Categories

### Message APIs

- **Send Messages** - Text, image, video, articles
- **Receive Messages** - Text, event, media messages
- **Template Messages** - Automated notifications
- **Customer Service** - Multi-agent conversations

### User Management

- **User Info** - Get user details, tags
- **User Tags** - Create/manage tags
- **Blacklist** - Manage blocked users

### Menu Management

- **Custom Menus** - Button configurations
- **Conditional Menus** - User-specific menus

### Material Management

- **Temporary Media** - Upload/download (3 days retention)
- **Permanent Media** - Long-term storage
- **Article Management** - News articles

### Analysis

- **User Analysis** - User growth, activity
- **Article Analysis** - Read statistics
- **Interface Analysis** - API call stats

## 4. Rate Limits

Common limits:
- **API Calls**: ~2000 calls per day (varies by account type)
- **Messages**: 4 messages/day/user (主动推送)
- **Mass Messages**: 1 broadcast per hour

See: [API Rate Limits](https://developers.weixin.qq.com/doc/service/dev/api/limit.html)

## 5. Security Considerations

### Message Encryption

WeChat supports symmetric encryption (AES):

```go
cfg := &config.Config{
    AppID:           "xxx",
    AppSecret:       "xxx",
    Token:           "xxx",
    EncodingAESKey:  "your-43-char-aes-key", // Optional
    Cache:           memory,
}
```

### IP Whitelist

Configure IP whitelist in WeChat Admin console for API access.

### Signature Verification

Verify requests using SHA1 signature:

```go
// Check signature from query params
signature := req.URL.Query().Get("signature")
timestamp := req.URL.Query().Get("timestamp")
nonce := req.URL.Query().Get("nonce")
```

## 6. Development Workflow

```
┌─────────────────┐
│  WeChat Server  │
└────────┬────────┘
         │ HTTPS
         ▼
┌─────────────────┐
│  Your Server    │  ◄── Verify, Process, Respond
└────────┬────────┘
         │
         ▼
┌─────────────────┐
│  Business Logic │
│  - Database     │
│  - APIs         │
│  - Cache        │
└─────────────────┘
```

## 7. Testing

### Test Accounts

- Use WeChat Test Account for development
- Register personal account for limited testing

### Local Development

1. Expose local server via ngrok or similar
2. Configure callback URL in WeChat admin
3. Test message flow

## 8. Deployment Checklist

- [ ] Server verification completed
- [ ] Access token caching implemented
- [ ] Message handler configured
- [ ] OAuth2 flow working
- [ ] Error handling in place
- [ ] Logging configured
- [ ] Rate limit handling implemented
- [ ] Security measures (IP whitelist, encryption)
- [ ] Monitoring and alerts set up

## References

- [Service Account Docs](https://developers.weixin.qq.com/doc/service/)
- [API Reference](https://developers.weixin.qq.com/doc/subscription/dev/api/)
- [SDK GoDocs](https://pkg.go.dev/github.com/silenceper/wechat/v2)
