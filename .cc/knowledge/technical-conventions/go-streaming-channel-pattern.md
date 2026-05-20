---
name: Go Streaming Channel Pattern
description: <-chan T + goroutine pattern for bridging third-party streaming APIs in Go
type: technical-convention
status: confirmed
confidence: high
source: voice-interaction change (Tasks 1-3)
---

**模式**: 用 `<-chan T` 接口 + 内部 goroutine 桥接第三方流式 API。

**典型代码**:
```go
// 接口层 — 返回只读 channel
type Client interface {
    ChatStream(ctx context.Context, ...) (<-chan StreamChunk, error)
}

// 实现层 — 启动 goroutine 转发
func (c *Client) ChatStream(ctx context.Context, ...) (<-chan StreamChunk, error) {
    stream, err := thirdParty.CreateStream(ctx, req)
    if err != nil {
        return nil, err
    }
    ch := make(chan StreamChunk, 8) // buffer 避免阻塞
    go func() {
        defer close(ch)
        defer stream.Close()
        for {
            // 检查 context 是否已取消
            if ctx.Err() != nil {
                return
            }
            item, err := stream.Recv()
            if err == io.EOF {
                ch <- StreamChunk{Done: true}
                return
            }
            if err != nil {
                ch <- StreamChunk{Error: err}
                return
            }
            ch <- StreamChunk{Content: item.Data}
        }
    }()
    return ch, nil
}
```

**关键要点**:
- Channel buffer (8-16) 给 goroutine 一点余量，避免 consumer 慢时立即阻塞
- 必须检查 `ctx.Err()` 防止 goroutine 泄漏
- goroutine 负责 `close(ch)` — channel 的 owner 负责关闭
- Done/Error 通过 channel 传播，不单独返回

**适用场景**: 将任何基于 callback/iterator 的流式 API 适配为 Go channel 接口。
