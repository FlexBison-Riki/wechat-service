# WeChat Service Account Developer Guide

## Overview

This guide provides a practical walkthrough for building WeChat Service Account (服务号) integrations, from server setup to production deployment. Based on the official WeChat documentation, this document covers essential concepts, common patterns, and implementation examples.

## Table of Contents

1. [Quick Start](#1-quick-start)
2. [Architecture Overview](#2-architecture-overview)
3. [Message Handling](#3-message-handling)
4. [AccessToken Management](#4-accesstoken-management)
5. [Media Management](#5-media-management)
6. [Custom Menus](#6-custom-menus)
7. [Mass Messaging](#7-mass-messaging)

---

## 1. Quick Start

### 1.1 Server Setup

**Recommended: Cloud Server (e.g., Tencent Cloud)**
- Purchase a cloud server with public IP
- Configure security groups (allow ports 80/443)
- Install required software:
  - Python 2.7+ or Go 1.22+
  - Web framework (Gin for Go, web.py for Python)
  - SSL certificates for HTTPS

**Alternative: WeChat Cloud Development**
- If you have a Mini Program with Cloud Development enabled
- Use [Service Account Environment Sharing](https://developers.weixin.qq.com/miniprogram/dev/wxcloud/basis/web.html)
- No server setup required

### 1.2 Account Registration

1. Register Service Account at [WeChat Official Platform](https://mp.weixin.qq.com/)
2. Apply for Developer permission at [WeChat Developer Platform](https://developers.weixin.qq.com/platform/)
3. Configure development settings:
   - Server URL (must be accessible from WeChat servers)
   - Token (for verification)
   - EncodingAESKey (optional, for encryption)

### 1.3 Basic Configuration

Navigate to: **WeChat Developer Platform → My Business → Service Account → Development Information**

Configure:
- URL: `https://your-domain.com/wechat`
- Token: Your verification token
- EncodingAESKey: Optional message encryption key

**Important:** Complete code logic before submitting for verification.

---

## 2. Architecture Overview

### Recommended Production Architecture

```
┌─────────────────────────────────────────────────────────┐
│                  Production Architecture                 │
├─────────────────────────────────────────────────────────┤
│                                                          │
│  ┌─────────────────────┐     ┌─────────────────────┐    │
│  │   Business Logic   │     │   API-Proxy Server  │    │
│  │     Server         │────▶│                     │    │
│  └─────────────────────┘     └─────────────────────┘    │
│           │                            │                  │
│           │                            │                  │
│           ▼                            ▼                  │
│  ┌───────────────────────────────────────────────┐    │
│  │          AccessToken中控服务器                 │    │
│  │  - Centralized token management               │    │
│  │  - Automatic refresh (proactive + reactive)  │    │
│  │  - Concurrency control (prevent race conditions)│ │
│  │  - Provides valid tokens to business logic   │    │
│  └───────────────────────────────────────────────┘    │
│                                                          │
│  Benefits:                                               │
│  ✓ Prevents token race conditions                       │
│  ✓ Improves system stability                          │
│  ✓ Provides single source of truth                   │
│  ✓ Enables API rate limiting and access control      │
│  ✓ Enhances security (hides internal APIs)          │
│  ✓ Ensures high availability                        │
│                                                          │
└─────────────────────────────────────────────────────────┘
```

### Key Components

#### AccessToken中控服务器 (Central Token Server)

**Responsibilities:**
- Proactively refresh tokens before expiration
- Reactively refresh when API returns invalid token errors
- Store tokens with proper concurrency control (locks)
- Serve valid tokens to business logic

**Why Essential:**
- Multiple concurrent requests could overwrite each other's tokens
- Without centralized control, tokens get corrupted
- Ensures API call stability

#### API-Proxy服务器 (API Proxy Server)

**Responsibilities:**
- Single point of contact with WeChat APIs
- Rate limiting per business domain
- Access control and permissions
- API versioning and routing

**Benefits:**
- High availability (other proxies can take over if one fails)
- Security (hides internal service interfaces)
- Attack prevention (not directly exposing internal APIs)

---

## 3. Message Handling

### 3.1 Message Types

| Type | Description | XML Element |
|------|-------------|-------------|
| Text | Plain text messages | `<xml><MsgType>text</MsgType><Content>...</Content></xml>` |
| Image | Images with MediaId | `<xml><MsgType>image</MsgType><MediaId>...</MediaId></xml>` |
| Voice | Voice messages | `<xml><MsgType>voice</MsgType><MediaId>...</MediaId></xml>` |
| Video | Video messages | `<xml><MsgType>video</MsgType><MediaId>...</MediaId></xml>` |
| Location | Geographic location | `<xml><MsgType>location</MsgType><Location_X>...</Location_X></xml>` |
| Link | Article links | `<xml><MsgType>link</MsgType><Title>...</Title></xml>` |
| Event | User actions | `<xml><MsgType>event</MsgType><Event>...</Event></xml>` |

### 3.2 Message Flow

```
┌─────────────────────────────────────────────────────────┐
│                  Message Flow                           │
├─────────────────────────────────────────────────────────┤
│                                                          │
│  User ──▶ WeChat Server ──▶ Developer Server            │
│     ◀── Response Message ◀──                          │
│                                                          │
│  1. User sends message to Service Account               │
│  2. WeChat server validates and forwards to your URL   │
│  3. Server processes message (parse XML)               │
│  4. Generate response XML                             │
│  5. Return XML to WeChat (within 5 seconds)          │
│  6. WeChat delivers message to user                   │
│                                                          │
└─────────────────────────────────────────────────────────┘
```

### 3.3 Passive Reply Messages

**Definition:** Responses triggered by user messages, not initiated by the service.

**Important Rules:**

1. **5-Second Timeout:**
   - Must respond within 5 seconds
   - If unable, reply with "success" or empty string ""
   - Otherwise, WeChat retries 3 times

2. **Retry Behavior:**
   - Without response: WeChat retries 3 times
   - After retries: User sees "Service unavailable" error
   - Reply "success": WeChat confirms receipt, no retry

3. **XML Structure:**
```xml
<xml>
    <ToUserName><![CDATA[openid]]></ToUserName>
    <FromUserName><![CDATA[gh_abc123]]></FromUserName>
    <CreateTime>1234567890</CreateTime>
    <MsgType><![CDATA[text]]></MsgType>
    <Content><![CDATA[response content]]></Content>
</xml>
```

### 3.4 Example: "You Ask, I Answer" (Echo Bot)

**Purpose:** Understand message receiving and sending

**Flow:**
1. User sends text message
2. Server receives and logs message
3. Server sends back the same text
4. User receives echo response

**Code Structure:**
```
main.py          # Entry point, HTTP server
handle.py        # Request routing
receive.py       # XML parsing
reply.py         # Response generation
```

### 3.5 Example: "Image Exchange"

**Purpose:** Introduce media handling

**Flow:**
1. User sends image
2. Server receives and extracts MediaId
3. Server replies with the same image
4. User receives their image back

**Key Points:**
- `PicUrl`: Public URL for image access
- `MediaId`: Unique identifier for the media
- To send different image: Upload via Material API first

---

## 4. AccessToken Management

### 4.1 Overview

**AccessToken** is required for most WeChat API calls.

**Properties:**
- Expires in 7200 seconds (2 hours)
- Must be refreshed proactively
- Cached locally with proper concurrency control

### 4.2 Why Centralized Management?

**Without Centralization:**
```
Request 1 → Get Token → Set Cache
Request 2 → Get Token → Overwrite Cache
Request 3 → Get Token → Overwrite Cache
...
```

**Problems:**
- Tokens get overwritten
- Later requests use invalid tokens
- API calls fail

**With Centralized Management:**
```
Token Server → Get Token (with lock)
               ↓
               Refresh if expired
               ↓
               Store and return
               ↓
All Requests → Get valid token from server
```

### 4.3 Implementation Pattern

```go
// Pseudocode for token management
type TokenServer struct {
    token   string
    expires time.Time
    mutex   sync.Mutex
}

func (s *TokenServer) GetToken() string {
    s.mutex.Lock()
    defer s.mutex.Unlock()
    
    if time.Now().After(s.expires) {
        s.refresh()
    }
    return s.token
}

func (s *TokenServer) refresh() {
    // API call to get new token
    // Update s.token and s.expires
}
```

---

## 5. Media Management

### 5.1 Media Types

#### Temporary Media (临时素材)
- **Retention:** 3 days
- **Limits:** Not stored in backend
- **Use Case:** One-time use images, quick interactions
- **APIs:** Upload, Download

#### Permanent Media (永久素材)
- **Retention:** Until explicitly deleted
- **Limits:** Total count per type (news: limited, images: limited, etc.)
- **Use Case:** Long-term storage, reuse
- **APIs:** Upload, Download, List, Delete

### 5.2 Media Operations

#### Upload Temporary Media
```go
// Using the SDK
media, err := officialAccount.GetMaterial().Upload(
    "image",
    filepath,
)
mediaID := media.MediaID
```

#### Download Media
```go
// Get media content
data, err := officialAccount.GetMaterial().Get(mediaID)
```

#### List Permanent Media
```go
// Batch get material list
list, err := officialAccount.GetMaterial().List("image", 0, 20)
```

### 5.3 Media ID Handling

**Important:**
- Media IDs are unique identifiers for media files
- Not stored in WeChat backend UI
- Must be obtained via:
  - Upload response
  - User message XML
  - List API response

---

## 6. Custom Menus

### 6.1 Menu Types

| Type | Description | Behavior |
|------|-------------|----------|
| **click** | Key click | Sends event to server |
| **view** | URL link | Opens URL in browser |
| **media_id** | Media | Sends stored media |
| **miniprogram** | Mini Program | Opens mini program |
| **location_select** | Location | Requests user location |
| **pic_*** | Photo | Requests photos from user |

### 6.2 Menu Creation Flow

```
1. Design menu structure (JSON)
2. Call Create Menu API
3. Users refresh to see new menu
4. Handle menu clicks (for click types)
```

### 6.3 Menu Click Events

When user clicks a click-type menu:
1. WeChat sends event XML to your server
2. Server processes the event
3. Server responds with appropriate message

**Example Event XML:**
```xml
<xml>
    <ToUserName><![CDATA[gh_abc]]></ToUserName>
    <FromUserName><![CDATA[openid]]></FromUserName>
    <CreateTime>1234567890</CreateTime>
    <MsgType><![CDATA[event]]></MsgType>
    <Event><![CDATA[CLICK]]></Event>
    <EventKey><![CDATA[MENU_KEY]]></EventKey>
</xml>
```

### 6.4 Implementation

```go
// Create menu
menu := []byte(`{
    "button": [
        {"type": "click", "name": "Today's Song", "key": "V1001_TODAY_MUSIC"},
        {"type": "view", "name": "Search", "url": "http://www.soso.com/"},
        {"name": "Menu", "sub_button": [
            {"type": "click", "name": "Yestoday", "key": "V1001_YESTERDAY"},
            {"type": "click", "name": "Song", "key": "V1001_TODAY_SONG"}
        ]}
    ]
}`)

// Handle click event
server.SetMessageHandler(func(msg *message.MixMessage) *message.Reply {
    switch msg.Event {
    case message.EventClick:
        return handleMenuClick(msg.EventKey)
    }
    return nil
})
```

---

## 7. Mass Messaging

### 7.1 Overview

Mass messaging allows sending messages to all followers or filtered groups.

**Limitations:**
- 1 broadcast per hour for mass send
- 4 messages/day/user for主动推送 (active push)
- Content restrictions apply

### 7.2 Message Types for Mass Send

| Type | Description |
|------|-------------|
| text | Text content |
| mpnews |图文消息 |
| voice | Voice messages |
| images | Image messages |
| mpvideo | Video messages |

### 7.3 Filtering Options

- **Tag:** By user tag
- **Region:** By location
- **Gender:** By gender
- **Language:** By language

---

## Best Practices

### 1. Error Handling

```go
// Always handle errors gracefully
func handleMessage(msg message.MixMessage) *message.Reply {
    defer func() {
        if r := recover(); r != nil {
            log.Error("Panic:", r)
        }
    }()
    
    // Process message
    // ...
    
    // If can't respond in 5s, return "success"
    return &message.Reply{MsgType: message.MsgTypeText, MsgData: text}
}
```

### 2. Security

1. **Verify Server:** Always verify requests from WeChat
2. **Encrypt Messages:** Use EncodingAESKey for sensitive data
3. **IP Whitelist:** Configure in WeChat admin
4. **Rate Limiting:** Implement on your server

### 3. Performance

1. **Cache AccessToken:** Don't fetch on every request
2. **Use Locks:** Prevent concurrent token refresh
3. **Async Processing:** For non-critical operations
4. **Connection Pooling:** For database/media storage

### 4. Monitoring

1. **Log All Requests:** For debugging
2. **Monitor API Calls:** Track success/failure rates
3. **Alert on Errors:** High priority issues
4. **Metrics Collection:** Prometheus integration

---

## Quick Reference

### Common API Endpoints

| Endpoint | Method | Description |
|----------|--------|-------------|
| `/wechat` | GET | Server verification |
| `/wechat` | POST | Message callback |
| `/cgi-bin/token` | GET | Get AccessToken |
| `/cgi-bin/menu/create` | POST | Create menu |
| `/cgi-bin/message/custom/send` | POST | Send客服 message |
| `/cgi-bin/material/upload` | POST | Upload media |

### Timeout Rules

| Operation | Timeout | Action |
|-----------|---------|--------|
| Message response | 5s | Return "success" if delayed |
| Token refresh | None | Auto-refresh before expiry |
| API calls | Varies | Handle rate limits |

### Error Codes

| Code | Meaning | Action |
|------|---------|--------|
| 0 | Success | Continue |
| 40001 | Invalid token | Refresh token |
| 40014 | Invalid access_token | Refresh token |
| 42001 | Token expired | Refresh token |
| 48001 | API not authorized | Check permissions |

---

## References

- [Service Account Documentation](https://developers.weixin.qq.com/doc/service/)
- [API Reference](https://developers.weixin.qq.com/doc/offiaccount/)
- [Community Forum](https://developers.weixin.qq.com/community/)
- [WeChat Developer Platform](https://developers.weixin.qq.com/platform/)
- [Official WeChat SDK (Go)](https://github.com/silenceper/wechat)

---

## Conclusion

Building a WeChat Service Account requires understanding:

1. **Server Setup:** Properly configured server with public access
2. **Token Management:** Centralized AccessToken handling
3. **Message Flow:** Request/response cycle within 5 seconds
4. **Media Handling:** Temporary and permanent素材
5. **Menu Design:** Interactive menu with event handling
6. **Production Architecture:** Scalable, maintainable design

Start with simple examples (echo bot, image exchange), then gradually add complexity. Always follow security best practices and implement proper error handling.

For production systems, use the recommended architecture with dedicated token server and API proxy for stability and security.
