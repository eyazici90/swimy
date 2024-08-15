# swim
Simple, unambitious implementation of [SWIM: Scalable Weakly-consistent Infection-style Process Group Membership Protocol](https://www.cs.cornell.edu/projects/Quicksilver/public_pdfs/SWIM.pdf)


## Usage

```go

    cfg := swim.DefaultConfig()
    cfg.OnJoin = func(addr net.Addr) {
    // do something with the notification
    }
    cfg.OnLeave = func(addr net.Addr) {
    // do something with the notification
    }
    ms1, err := swim.New(cfg)
    if err != nil {
    //
    }
    defer ms1.Stop()
    
    // some other app
    ms2, err := swim.New(nil) 
    if err != nil {
    //
    }
    defer ms2.Stop()
    err := ms2.Join(ctx, "127.0.0.1:54555") // existing known members to join the membership

```

## TODOs
- Add suspect mechanism to prevent false positives
- Implement lamport clock
- Instrument via otel & slog
- TCP conn. pooling per addr
- Add failure testing 
