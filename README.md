# WeChat Service Project

## Overview

This project provides WeChat platform integration capabilities for connecting with WeChat's Official Account Service (服务号).

## WeChat Platform Integration Guide

### Platforms Overview

**WeChat Official Platform (微信公众平台)**
- URL: https://mp.weixin.qq.com/
- Where operators create service accounts and manage user services

**WeChat Developer Platform (微信开发者平台)**
- URL: https://developers.weixin.qq.com/platform/
- One-stop workbench for WeChat ecosystem development
- Manage accounts: Official Accounts, Service Accounts, Mini Programs, Games, Stores, etc.
- Configure developer settings and API permissions

### Development Scenarios

#### 1. Server-Side API Calls

- Access service account openAPI endpoints
- Check API call rate limits
- Debug and troubleshoot API issues

Reference: [Service Account API Overview](https://developers.weixin.qq.com/doc/service/dev/api/)

#### 2. Web Application Development (H5)

- Build web applications within WeChat
- OAuth2 authentication
- JS-SDK integration

Reference: [Service Account H5 Development](https://developers.weixin.qq.com/doc/service/h5/)

### Required Configuration

1. **AppID** - Unique identifier for your service account
2. **AppSecret** - Secret key for API authentication
3. **Token** - Verification token for message server
4. **EncodingAESKey** - Optional encryption key for secure communication

## SDK Recommendation

This project uses [silenceper/wechat](https://github.com/silenceper/wechat) - a Go SDK for WeChat.

### SDK Features

- **Simple & Easy to Use** - Clean API design
- **Comprehensive** - Supports multiple WeChat products
- **Well Maintained** - Active development with 5.2k+ stars

### Supported Modules

| Module | Description |
|--------|-------------|
| `officialaccount` | Official Account (公众号/服务号) APIs |
| `miniprogram` | Mini Program APIs |
| `minigame` | Mini Game APIs |
| `pay` | WeChat Payment APIs |
| `openplatform` | Open Platform APIs |
| `work` | WeChat Work (企业微信) |
| `aispeech` | AI Conversation APIs |

### Quick Start

```go
import "github.com/silenceper/wechat/v2"

wc := wechat.NewWechat()
memory := cache.NewMemory()
cfg := &config.Config{
    AppID:     "your-appid",
    AppSecret: "your-appsecret",
    Token:     "your-token",
    Cache:     memory,
}
officialAccount := wc.GetOfficialAccount(cfg)
```

## Project Structure

```
wechat-service/
├── docs/              # Documentation
├── sdk/               # SDK integration code
├── README.md          # This file
└── integration.md     # Integration details
```

## Resources

- [WeChat Service Account Documentation](https://developers.weixin.qq.com/doc/service/)
- [WeChat Developer Platform](https://developers.weixin.qq.com/platform/)
- [WeChat SDK (Go) Documentation](https://silenceper.com/wechat)
- [SDK Examples](https://github.com/gowechat/example)
- [Developer Guide](docs/developer-guide.md) - Comprehensive guide for building WeChat integrations
