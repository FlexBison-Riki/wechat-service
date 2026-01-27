# WeChat Message & Menu System Design Guide

## Overview

This document summarizes key points from WeChat Service Account documentation for system design, covering:
1. Custom Menu (自定义菜单)
2. Receiving Messages (接收消息)
3. Receiving Events (接收事件推送)
4. Passive Reply Messages (被动回复消息)

---

## 1. Custom Menu (自定义菜单)

### Overview

Custom menus help Service Accounts present a better interface for users to understand functionality quickly.

### Menu API Endpoints

| Function | English Name | Endpoint |
|----------|-------------|----------|
| Create Menu | `createCustomMenu` | `POST /cgi-bin/menu/create` |
| Query Menu Info | `getCurrentSelfmenuInfo` | `GET /cgi-bin/get_current_selfmenu_info` |
| Get Menu Config | `getMenu` | `GET /cgi-bin/menu/get` |
| Delete Menu | `deleteMenu` | `GET /cgi-bin/menu/delete` |
| Create Conditional Menu | `addConditionalMenu` | `POST /cgi-bin/menu/addconditional` |
| Delete Conditional Menu | `deleteConditionalMenu` | `POST /cgi-bin/menu/delconditional` |
| Test Conditional Menu Match | `tryMatchMenu` | `POST /cgi-bin/menu/trymatch` |

### Menu Types

#### 1.1 Click to Message (CLICK)
User clicks menu → triggers event with `EventKey` → developer receives event → can reply message

#### 1.2 Click to View URL (VIEW)
User clicks menu → redirects to URL in `EventKey` → no event pushed to server

#### 1.3 Scan Code (scancode_push / scancode_waitmsg)
- `scancode_push`: Immediate scan result push
- `scancode_waitmsg`: Shows "receiving message" dialog, then pushes scan result

**Event Data:**
```xml
<Event>scancode_push</Event>
<EventKey>{user_defined_key}</EventKey>
<ScanCodeInfo>
    <ScanType>qrcode</ScanType>
    <ScanResult>{qr_content}</ScanResult>
</ScanCodeInfo>
```

#### 1.4 Photo Actions (pic_*)
- `pic_sysphoto`: System camera
- `pic_photo_or_album`: Camera or album
- `pic_weixin`: WeChat album

**Event Data:**
```xml
<SendPicsInfo>
    <Count>{count}</Count>
    <PicList>
        <item><PicMd5Sum>{md5}</PicMd5Sum></item>
    </PicList>
</SendPicsInfo>
```

#### 1.5 Location Select (location_select)
**Event Data:**
```xml
<SendLocationInfo>
    <Location_X>{lat}</Location_X>
    <Location_Y>{lng}</Location_Y>
    <Scale>{scale}</Scale>
    <Label>{address}</Label>
    <Poiname>{poi_name}</Poiname>
</SendLocationInfo>
```

#### 1.6 View Mini Program (view_miniprogram)
```xml
<Event>view_miniprogram</Event>
<EventKey>{miniprogram_path}</EventKey>
<MenuId>{menu_id}</MenuId>
```

### Important Notes

- Clicking sub-menus does NOT trigger events
- Menu events only supported on WeChat iOS 5.4.1+ and Android 5.4+
- `MenuId` available for conditional menus to identify which rule was matched

### System Design Considerations

```
┌─────────────────────────────────────────────────────────────┐
│                    Menu Management System                    │
├─────────────────────────────────────────────────────────────┤
│                                                              │
│  ┌──────────────┐    ┌──────────────┐    ┌──────────────┐   │
│  │ Menu Builder │───▶│  Validator   │───▶│  Sync to     │   │
│  │   (Config)   │    │  (Rules)     │    │  WeChat API  │   │
│  └──────────────┘    └──────────────┘    └──────────────┘   │
│          │                                        │          │
│          ▼                                        ▼          │
│  ┌──────────────┐                        ┌──────────────┐   │
│  │ Conditional  │                        │  Cache       │   │
│  │  Rules DB    │                        │  (Redis)     │   │
│  └──────────────┘                        └──────────────┘   │
│                                                              │
└─────────────────────────────────────────────────────────────┘
```

**Key Components:**
- Menu Versioning: Track menu changes with timestamps
- Conditional Matching Engine: User tags, location, language matching
- Event Router: Route menu events to appropriate handlers
- A/B Testing: Test different menu configurations

---

## 2. Receiving Standard Messages (接收消息)

### Overview

When users send messages to Service Account, WeChat servers POST XML data to developer's configured URL.

### Protocol Specifications

- **HTTP Method**: POST
- **Content-Type**: application/xml
- **Response Timeout**: 5 seconds (mandatory)
- **Retry Logic**: 3 retries if no response
- **Deduplication**: Use `MsgId` for message deduplication

### Message Types

#### 2.1 Text Message
```xml
<xml>
    <ToUserName><![CDATA[developer_account]]></ToUserName>
    <FromUserName><![CDATA[openid]]></FromUserName>
    <CreateTime>1234567890</CreateTime>
    <MsgType><![CDATA[text]]></MsgType>
    <Content><![CDATA[用户发送的内容]]></Content>
    <MsgId>1234567890123456</MsgId>
    <MsgDataId>xxxx</MsgDataId>       <!-- Optional: from article -->
    <Idx>1</Idx>                       <!-- Optional: article index -->
</xml>
```

#### 2.2 Image Message
```xml
<MsgType><![CDATA[image]]></MsgType>
<PicUrl><![CDATA[http://xxx.jpg]]></PicUrl>
<MediaId><![CDATA[media_id]]></MediaId>
```

#### 2.3 Voice Message
```xml
<MsgType><![CDATA[voice]]></MsgType>
<MediaId><![CDATA[media_id]]></MediaId>
<Format><![CDATA[amr]]></Format>
<MediaId16K><![CDATA[media_id_16k]]></MediaId16K>
```

#### 2.4 Video Message
```xml
<MsgType><![CDATA[video]]></MsgType>
<MediaId><![CDATA[media_id]]></MediaId>
<ThumbMediaId><![CDATA[thumb_media_id]]></ThumbMediaId>
```

#### 2.5 Short Video Message
```xml
<MsgType><![CDATA[shortvideo]]></MsgType>
```

#### 2.6 Location Message
```xml
<MsgType><![CDATA[location]]></MsgType>
<Location_X>23.137466</Location_X>
<Location_Y>113.425425</Location_Y>
<Scale>20</Scale>
<Label><![CDATA[位置信息]]></Label>
```

#### 2.7 Link Message
```xml
<MsgType><![CDATA[link]]></MsgType>
<Title><![CDATA[标题]]></Title>
<Description><![CDATA[描述]]></Description>
<Url><![CDATA[http://xxx]]></Url>
```

### Media Retrieval

For media messages, use `MediaId` with Material API to download:
- **Temporary Media**: 3-day retention
- **Permanent Media**: Long-term storage

### System Design Considerations

```
┌─────────────────────────────────────────────────────────────┐
│                   Message Processing Pipeline                │
├─────────────────────────────────────────────────────────────┤
│                                                              │
│  WeChat Server                                               │
│       │                                                      │
│       ▼                                                      │
│  ┌─────────┐    ┌─────────┐    ┌─────────┐    ┌─────────┐  │
│  │  Verify │───▶│  Parse  │───▶│ Dedupe  │───▶│  Route  │  │
│  │ Signature│    │  XML    │    │(MsgId)  │    │ Handler │  │
│  └─────────┘    └─────────┘    └─────────┘    └─────────┘  │
│                                              │              │
│                                              ▼              │
│  ┌──────────────────────────────────────────────────────┐  │
│  │                  Handler Types                        │  │
│  │  ┌─────────┐ ┌─────────┐ ┌─────────┐ ┌─────────┐    │  │
│  │  │  Text   │ │  Image  │ │  Voice  │ │  Video  │    │  │
│  │  │ Handler │ │ Handler │ │ Handler │ │ Handler │    │  │
│  │  └─────────┘ └─────────┘ └─────────┘ └─────────┘    │  │
│  │  ┌─────────┐ ┌─────────┐ ┌─────────┐               │  │
│  │  │ Location│ │  Link   │ │  Event  │               │  │
│  │  │ Handler │ │ Handler │ │ Handler │               │  │
│  │  └─────────┘ └─────────┘ └─────────┘               │  │
│  └──────────────────────────────────────────────────────┘  │
│                                              │              │
│       ◀──────────────────────────────────────┘              │
│       │   (Passive Reply)                                        │
│       ▼                                                      │
│  ┌──────────────────────────────────────────────────────┐   │
│  │  Response Builder (XML generation)                     │   │
│  └──────────────────────────────────────────────────────┘   │
│                                                              │
└─────────────────────────────────────────────────────────────┘
```

**Key Components:**
- **Signature Verifier**: Validate WeChat signature (timestamp + nonce + token)
- **XML Parser**: Convert XML to structured objects
- **Deduplication Cache**: Redis-based, use MsgId, TTL ~10 minutes
- **Message Queue**: For async processing (Kafka/RabbitMQ)
- **Handler Registry**: Map MsgType to handler functions
- **Response Builder**: Generate XML responses

---

## 3. Receiving Event Pushes (接收事件推送)

### Overview

Events are triggered by user actions. Some events allow response, some don't.

### Event Types

#### 3.1 Subscribe/Unsubscribe Events

**Subscribe:**
```xml
<MsgType><![CDATA[event]]></MsgType>
<Event><![CDATA[subscribe]]></Event>
```

**Unsubscribe:**
```xml
<Event><![CDATA[unsubscribe]]></Event>
```

**Key Actions:**
- Send welcome message on subscribe
- **DELETE all user data** on unsubscribe (privacy protection)
- Deduplication: Use `FromUserName + CreateTime`

#### 3.2 Scan QR Code Events

**User not subscribed (new subscribe):**
```xml
<Event><![CDATA[subscribe]]></Event>
<EventKey><![CDATA[qrscene_12345]]></EventKey>
<Ticket><![CDATA[ticket]]></Ticket>
```

**User already subscribed:**
```xml
<Event><![CDATA[SCAN]]></Event>
<EventKey><![CDATA[12345]]></EventKey>
<Ticket><![CDATA[ticket]]></Ticket>
```

**Use Cases:**
- User onboarding with referral tracking
- Campaign attribution
- Location-based services

#### 3.3 Location Report Event

```xml
<Event><![CDATA[LOCATION]]></Event>
<Latitude>23.137466</Latitude>
<Longitude>113.425425</Longitude>
<Precision>119.38541</Precision>
```

**Behavior:**
- Triggered on first entry into chat
- Then every 5 seconds (configurable in WeChat admin)
- Requires user permission

#### 3.4 Custom Menu Events

**Click to Message:**
```xml
<Event><![CDATA[CLICK]]></Event>
<EventKey><![CDATA[USER_DEFINED_KEY]]></EventKey>
```

**Click to View URL:**
```xml
<Event><![CDATA[VIEW]]></Event>
<EventKey><![CDATA[http://example.com]]></EventKey>
<MenuId>12345</MenuId>
```

### Event Comparison Table

| Event | Description | Can Reply? | Dedupe Key |
|-------|-------------|------------|------------|
| subscribe | User subscribed | ✓ | FromUserName + CreateTime |
| unsubscribe | User unsubscribed | ✗ | FromUserName + CreateTime |
| SCAN | QR code scanned | ✓ | FromUserName + CreateTime |
| LOCATION | Location reported | ✗ | - |
| CLICK | Menu clicked | ✓ | FromUserName + CreateTime |
| VIEW | Menu URL opened | ✗ | - |

### System Design Considerations

```
┌─────────────────────────────────────────────────────────────┐
│                    Event Processing System                   │
├─────────────────────────────────────────────────────────────┤
│                                                              │
│  ┌─────────────────────────────────────────────────────────┐ │
│  │                  Event Router                            │ │
│  └─────────────────────────────────────────────────────────┘ │
│       │                                                      │
│       ├──▶ Subscribe Handler                                │
│       │      ├─ Create user record                          │
│       │      ├─ Send welcome message                        │
│       │      └─ Trigger onboarding flow                     │
│       │                                                      │
│       ├──▶ Unsubscribe Handler                              │
│       │      ├─ Mark user inactive                          │
│       │      └─ DELETE user data (privacy)                  │
│       │                                                      │
│       ├──▶ Scan Handler                                     │
│       │      ├─ Extract scene ID                            │
│       │      ├─ Update attribution                          │
│       │      └─ Trigger campaign logic                      │
│       │                                                      │
│       ├──▶ Location Handler                                 │
│       │      ├─ Update user location                        │
│       │      └─ Trigger location-based services             │
│       │                                                      │
│       └──▶ Menu Handler                                     │
│              ├─ Route by EventKey                           │
│              ├─ Execute menu action                         │
│              └─ Return response                             │
│                                                              │
│  ┌─────────────────────────────────────────────────────────┐ │
│  │                  User Profile Manager                    │ │
│  │  ┌─────────┐  ┌─────────┐  ┌─────────┐  ┌─────────┐    │ │
│  │  │ Profile │  │  Tags   │  │ Location│  │  Stats  │    │ │
│  │  │  Store  │  │ Manager │  │  Store  │  │ Counter │    │ │
│  │  └─────────┘  └─────────┘  └─────────┘  └─────────┘    │ │
│  └─────────────────────────────────────────────────────────┘ │
│                                                              │
└─────────────────────────────────────────────────────────────┘
```

**Key Components:**
- **Event Dispatcher**: Route events to handlers based on Event type
- **User Profile Service**: CRUD operations for user data
- **Welcome Message Engine**: Template-based welcome messages
- **Data Cleanup Service**: GDPR compliance for unsubscribe
- **Scene Value Parser**: Extract context from QR codes
- **Location Tracker**: Store and query user locations

---

## 4. Passive Reply Messages (被动回复消息)

### Overview

When user sends message or triggers certain events, developer can respond with XML packet within 5 seconds.

### Protocol Specifications

- **Response Time**: Must respond within 5 seconds
- **Retry**: 3 retries if timeout (no response)
- **Empty Response**: Return "success" or empty string to skip retry
- **Encryption**: Optional (configurable in WeChat admin)

### Message Types

#### 4.1 Text Message
```xml
<xml>
    <ToUserName><![CDATA[openid]]></ToUserName>
    <FromUserName><![CDATA[developer_account]]></FromUserName>
    <CreateTime>1234567890</CreateTime>
    <MsgType><![CDATA[text]]></MsgType>
    <Content><![CDATA[回复内容]]></Content>
</xml>
```

#### 4.2 Image Message
```xml
<MsgType><![CDATA[image]]></MsgType>
<Image>
    <MediaId><![CDATA[media_id]]></MediaId>
</Image>
```
**Note**: Media must be uploaded via Material API first

#### 4.3 Voice Message
```xml
<MsgType><![CDATA[voice]]></MsgType>
<Voice>
    <MediaId><![CDATA[media_id]]></MediaId>
</Voice>
```

#### 4.4 Video Message
```xml
<MsgType><![CDATA[video]]></Video>
<Video>
    <MediaId><![CDATA[media_id]]></MediaId>
    <Title><![CDATA[标题]]></Title>
    <Description><![CDATA[描述]]></Description>
</Video>
```

#### 4.5 Music Message
```xml
<MsgType><![CDATA[music]]></MsgType>
<Music>
    <Title><![CDATA[标题]]></Title>
    <Description><![CDATA[描述]]></Description>
    <MusicUrl><![CDATA[http://xxx.mp3]]></MusicUrl>
    <HQMusicUrl><![CDATA[http://xxx.mp3]]></HQMusicUrl>
    <ThumbMediaId><![CDATA[thumb_media_id]]></ThumbMediaId>
</Music>
```

#### 4.6 News (Article) Message
```xml
<MsgType><![CDATA[news]]></MsgType>
<ArticleCount>1</ArticleCount>
<Articles>
    <item>
        <Title><![CDATA[标题]]></Title>
        <Description><![CDATA[描述]]></Description>
        <PicUrl><![CDATA[http://xxx.jpg]]></PicUrl>
        <Url><![CDATA[http://xxx]]></Url>
    </item>
</Articles>
```

**Limits:**
- 1 article for: text, image, voice, video, location, link messages
- Up to 8 articles for other scenarios

#### 4.7 Transfer to AI (AI Reply)
```xml
<MsgType><![CDATA[transfer_biz_ai_ivr]]></MsgType>
```
**Requirements:**
- AI reply enabled in WeChat admin
- AI trained on historical articles

### Error Handling

**When NOT to reply:**
- Return "success" or empty string "" to prevent retry
- Use async Customer Service API for delayed responses

**Error Scenarios:**
1. No response in 5 seconds → User sees "Service unavailable"
2. Invalid response (e.g., JSON) → User sees "Service unavailable"

### System Design Considerations

```
┌─────────────────────────────────────────────────────────────┐
│                   Response Generation System                 │
├─────────────────────────────────────────────────────────────┤
│                                                              │
│  ┌─────────────────────────────────────────────────────────┐ │
│  │              Response Strategy Selector                  │ │
│  └─────────────────────────────────────────────────────────┘ │
│       │                                                      │
│       ├──▶ Text Response                                    │
│       │      └─ Template Engine                             │
│       │                                                      │
│       ├──▶ Media Response                                   │
│       │      ├─ Check media cache                           │
│       │      └─ Upload if missing (Material API)            │
│       │                                                      │
│       ├──▶ News Response                                    │
│       │      └─ Article Builder                             │
│       │                                                      │
│       └──▶ AI Transfer                                      │
│              └─ Check AI availability                       │
│                                                              │
│  ┌─────────────────────────────────────────────────────────┐ │
│  │              XML Builder                                 │ │
│  │  ┌─────────┐  ┌─────────┐  ┌─────────┐  ┌─────────┐    │ │
│  │  │  Text   │  │  Image  │  │  Video  │  │  News   │    │ │
│  │  │ Builder │  │ Builder │  │ Builder │  │ Builder │    │ │
│  │  └─────────┘  └─────────┘  └─────────┘  └─────────┘    │ │
│  └─────────────────────────────────────────────────────────┘ │
│       │                                                      │
│       ▼                                                      │
│  ┌─────────────────────────────────────────────────────────┐ │
│  │              Response Cache                              │ │
│  │  - Template responses                                    │ │
│  │  - Media metadata                                        │ │
│  └─────────────────────────────────────────────────────────┘ │
│       │                                                      │
│       ▼                                                      │
│  ┌─────────────────────────────────────────────────────────┐ │
│  │              Timeout Monitor                             │ │
│  │  - Track processing time                                 │ │
│  │  - Trigger fallback if > 4s                              │ │
│  └─────────────────────────────────────────────────────────┘ │
│                                                              │
└─────────────────────────────────────────────────────────────┘
```

**Key Components:**
- **Response Factory**: Create response objects based on type
- **Template Engine**: Dynamic text generation with variables
- **Media Manager**: Cache and upload media to WeChat
- **Article Builder**: Construct news articles
- **Timeout Guard**: Return "success" if processing > 4 seconds
- **Encryption Handler**: Optional XML encryption/decryption

---

## 5. Architecture Summary

### Complete Flow Diagram

```
┌────────────────────────────────────────────────────────────────────┐
│                        WeChat User                                  │
└────────────────────────────┬───────────────────────────────────────┘
                             │
                             ▼ (HTTPS POST / GET)
┌────────────────────────────────────────────────────────────────────┐
│                     WeChat Servers                                  │
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐             │
│  │  Message     │  │   Event      │  │   Menu       │             │
│  │  Queue       │  │   Queue      │  │   Click      │             │
│  └──────┬───────┘  └──────┬───────┘  └──────┬───────┘             │
└─────────┼─────────────────┼─────────────────┼───────────────────────┘
          │                 │                 │
          ▼                 ▼                 ▼
┌────────────────────────────────────────────────────────────────────┐
│                    Application Server                               │
│                                                                      │
│  ┌───────────────────────────────────────────────────────────────┐  │
│  │                    API Gateway                                 │  │
│  │  - Signature verification                                      │  │
│  │  - Rate limiting                                               │  │
│  │  - Request parsing                                             │  │
│  └───────────────────────────────────────────────────────────────┘  │
│                              │                                      │
│                              ▼                                      │
│  ┌───────────────────────────────────────────────────────────────┐  │
│  │                  Message/Event Router                          │  │
│  └───────────────────────────────────────────────────────────────┘  │
│                              │                                      │
│       ┌──────────────────────┼──────────────────────┐              │
│       ▼                      ▼                      ▼              │
│  ┌─────────┐           ┌─────────┐           ┌─────────┐          │
│  │ Message │           │  Event  │           │  Menu   │          │
│  │ Handler │           │ Handler │           │ Handler │          │
│  └────┬────┘           └────┬────┘           └────┬────┘          │
│       │                     │                     │                │
│       ▼                     ▼                     ▼                │
│  ┌───────────────────────────────────────────────────────────────┐  │
│  │                  Business Logic Layer                          │  │
│  │  - User Service         - Campaign Service                     │  │
│  │  - Content Service      - Location Service                     │  │
│  │  - Media Service        - AI Service                           │  │
│  └───────────────────────────────────────────────────────────────┘  │
│                              │                                      │
│                              ▼                                      │
│  ┌───────────────────────────────────────────────────────────────┐  │
│  │                  Response Builder                              │  │
│  │  - Text/XML generation                                         │  │
│  │  - Media lookup                                                │  │
│  │  - Encryption (optional)                                       │  │
│  └───────────────────────────────────────────────────────────────┘  │
│                              │                                      │
│                              ▼                                      │
│  ┌───────────────────────────────────────────────────────────────┐  │
│  │                  Response Cache (Redis)                        │  │
│  └───────────────────────────────────────────────────────────────┘  │
│                                                                      │
└──────────────────────────────────────────────────────────────────────┘
                              │
                              ▼ (XML Response)
                       ┌──────────────────┐
                       │   WeChat Server  │
                       │  (delivers to user)
                       └──────────────────┘
```

### Key Technical Decisions

| Aspect | Decision | Reason |
|--------|----------|--------|
| Sync/Async | Hybrid | Quick responses sync, heavy processing async via queue |
| Message Storage | Time-series DB | Efficient for message history and analytics |
| Media Storage | Object Storage (OSS/S3) | Cost-effective for media files |
| Token Cache | Redis | High-performance, automatic expiration |
| Event Dedupe | Redis with TTL | Atomic operations, fast lookup |
| Response Time | < 4s processing | 5s total timeout (leave 1s buffer) |
| Encryption | Optional | Start without, enable when needed |

### Data Models

```go
// User
type User struct {
    OpenID       string    `json:"openid"`
    SubscribeAt  time.Time `json:"subscribe_at"`
    UnsubscribeAt time.Time `json:"unsubscribe_at,omitempty"`
    Tags         []string  `json:"tags"`
    Location     *Location `json:"location,omitempty"`
    Profile      *Profile  `json:"profile,omitempty"`
    CreatedAt    time.Time `json:"created_at"`
    UpdatedAt    time.Time `json:"updated_at"`
}

// Message
type Message struct {
    ID          int64     `json:"id"`
    MsgID       int64     `json:"msg_id"`
    FromUser    string    `json:"from_user"`
    ToUser      string    `json:"to_user"`
    MsgType     string    `json:"msg_type"`
    Content     string    `json:"content,omitempty"`
    MediaID     string    `json:"media_id,omitempty"`
    Location    *Location `json:"location,omitempty"`
    CreatedAt   time.Time `json:"created_at"`
}

// Event
type Event struct {
    ID        int64     `json:"id"`
    OpenID    string    `json:"openid"`
    EventType string    `json:"event_type"`
    EventKey  string    `json:"event_key"`
    Ticket    string    `json:"ticket,omitempty"`
    Location  *Location `json:"location,omitempty"`
    MenuID    string    `json:"menu_id,omitempty"`
    CreatedAt time.Time `json:"created_at"`
}
```

---

## 6. References

### Official Documentation

- [Custom Menu Guide](https://developers.weixin.qq.com/doc/service/guide/product/menu/intro.html)
- [Receiving Messages](https://developers.weixin.qq.com/doc/service/guide/product/message/Receiving_standard_messages.html)
- [Receiving Events](https://developers.weixin.qq.com/doc/service/guide/product/message/Receiving_event_pushes.html)
- [Passive Reply](https://developers.weixin.qq.com/doc/service/guide/product/message/Passive_user_reply_message.html)

### Related Documentation

- [Message Encryption](https://developers.weixin.qq.com/doc/service/dev/push/encryption.html)
- [API Rate Limits](https://developers.weixin.qq.com/doc/service/dev/api/limit.html)
- [Material API](https://developers.weixin.qq.com/doc/subscription/dev/api/material.html)
