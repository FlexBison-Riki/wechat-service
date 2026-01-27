# SDK Analysis: silenceper/wechat (Go)

## Overview

**Repository**: https://github.com/silenceper/wechat
**Stars**: 5.2k+
**License**: Apache License 2.0
**Language**: Go (Golang)
**Version**: v2 (recommended)

A simple, easy-to-use WeChat SDK for Go with comprehensive coverage of WeChat APIs.

## Architecture

### Core Structure

```
wechat.NewWechat()  →  WeChat instance
    ↓
GetOfficialAccount(cfg)  →  OfficialAccount instance
    ↓
GetServer(req, rw)  →  Server instance (message handling)
```

### Configuration

```go
type Config struct {
    AppID           string
    AppSecret       string
    Token           string
    EncodingAESKey  string  // Optional
    Cache           Cache   // Required for token storage
    UseStableAK     bool    // Use stable access token
}
```

## Module Breakdown

### 1. OfficialAccount (公众号/服务号)

Main class with methods for all OA features:

| Method | Returns | Purpose |
|--------|---------|---------|
| `GetBasic()` | Basic | QR codes, short URLs |
| `GetMenu()` | Menu | Custom menu management |
| `GetServer()` | Server | Message handling |
| `GetOauth()` | Oauth | Web OAuth2 |
| `GetMaterial()` | Material | Media files |
| `GetDraft()` | Draft | Draft management |
| `GetJs()` | Js | JS-SDK config |
| `GetUser()` | User | User management |
| `GetTemplate()` | Template | Template messages |
| `GetCustomerMessageManager()` | Manager | Customer service |
| `GetBroadcast()` | Broadcast | Mass messaging |
| `GetDataCube()` | DataCube | Analytics |
| `GetOCR()` | OCR | Text recognition |
| `GetSubscribe()` | Subscribe | Subscription messages |
| `GetCustomerServiceManager()` | Manager | Customer service admin |
| `GetOpenAPI()` | OpenAPI | Open platform APIs |

### 2. Message Handling

**Server** - Handles incoming messages:

```go
server.SetMessageHandler(func(msg *message.MixMessage) *message.Reply {
    // Process message
    // Return response
})
server.Serve()
server.Send()
```

**Message Types Supported**:
- Text
- Image
- Voice
- Video
- Short Video
- Location
- Link
- Event (subscribe, click, etc.)

### 3. Credential Management

**Access Token Handling**:
- `DefaultAccessToken` - Standard token refresh
- `StableAccessToken` - More stable token (avoids some limits)
- Custom token handlers supported

**Cache Interface**:
```go
type Cache interface {
    Get(key string) (interface{}, error)
    Set(key string, value interface{}) error
}
```

**Built-in Caches**:
- Memory cache (for dev)
- Redis cache (production)
- Memcached cache

### 4. User Management

```go
user := officialAccount.GetUser()
// List users: user.List()
// Get user info: user.Info(openID)
// Update remark: user.UpdateRemark(openID, remark)
// Get user tags: user.GetTags()
// Batch tag operations
```

### 5. Material/Media Management

```go
material := officialAccount.GetMaterial()
// Upload: material.Upload(mediaType, file)
// Download: material.Get(mediaID)
// Delete: material.Delete(mediaID)
// Lists and statistics
```

### 6. OAuth2

```go
oauth := officialAccount.GetOauth()
// Get redirect URL: oauth.GetRedirectURL(callback, scope, state)
// Get user info: oauth.GetUserInfo(code)
// Refresh token: oauth.RefreshToken(refreshToken)
```

### 7. JS-SDK

```go
js := officialAccount.GetJs()
// Get signature: js.GetSignature(url)
// Config params ready for frontend
```

## Dependencies

```go
require (
    github.com/alicebob/miniredis/v2 v2.30.0
    github.com/bradfitz/gomemcache v0.0.0-20220106215444-fb4bf637b56d
    github.com/fatih/structs v1.1.0
    github.com/go-redis/redis/v8 v8.11.5
    github.com/sirupsen/logrus v1.9.0
    github.com/spf13/cast v1.4.1
    github.com/stretchr/testify v1.7.1
    github.com/tidwall/gjson v1.14.1
    golang.org/x/crypto v0.0.0-20220622213112-05595931fe9d
    gopkg.in/h2non/gock.v1 v1.1.2
)
```

## Usage Example

```go
package main

import (
    "fmt"
    "net/http"
    
    "github.com/silenceper/wechat/v2"
    "github.com/silenceper/wechat/v2/cache"
    "github.com/silenceper/wechat/v2/officialaccount/config"
    "github.com/silenceper/wechat/v2/officialaccount/message"
)

func main() {
    wc := wechat.NewWechat()
    memory := cache.NewMemory()
    cfg := &config.Config{
        AppID:     "your-appid",
        AppSecret: "your-appsecret",
        Token:     "your-token",
        Cache:     memory,
    }
    officialAccount := wc.GetOfficialAccount(cfg)
    
    // Message handler
    http.HandleFunc("/wechat", func(rw http.ResponseWriter, req *http.Request) {
        server := officialAccount.GetServer(req, rw)
        server.SetMessageHandler(func(msg *message.MixMessage) *message.Reply {
            // Echo back text
            text := message.NewText("Received: " + msg.Content)
            return &message.Reply{MsgType: message.MsgTypeText, MsgData: text}
        })
        server.Serve()
        server.Send()
    })
    
    fmt.Println("Server starting on :8080")
    http.ListenAndServe(":8080", nil)
}
```

## Pros & Cons

### Advantages

✓ Simple, clean API design
✓ Comprehensive WeChat API coverage
✓ Good documentation
✓ Active maintenance (611 commits)
✓ Supports multiple WeChat products
✓ Flexible token handling
✓ Production-ready

### Limitations

- Some APIs may be incomplete (check doc/api)
- Documentation mainly in Chinese
- v2 requires Go 1.16+

## Alternative SDKs

| SDK | Language | Notes |
|-----|----------|-------|
| wechat-php | PHP | Official-ish |
| wechat4j | Java | Popular for Java |
| wechat-dev-tools | JS/Node | Node.js |

## Resources

- [Full Documentation](https://silenceper.com/wechat)
- [GoDoc Reference](https://pkg.go.dev/github.com/silenceper/wechat/v2)
- [Examples Repository](https://github.com/gowechat/example)
- [API Coverage List](https://github.com/silenceper/wechat/tree/v2/doc/api)
