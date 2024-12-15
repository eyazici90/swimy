# swimy
Simple, unambitious implementation of [swimy: Scalable Weakly-consistent Infection-style Process Group Membership Protocol](https://www.cs.cornell.edu/projects/Quicksilver/public_pdfs/swimy.pdf)
`plus` some improvements.



## Usage

```go

    cfg := swimy.DefaultConfig()
    cfg.OnJoin = func(addr net.Addr) {
    // do something with the notification
    }
    cfg.OnLeave = func(addr net.Addr) {
    // do something with the notification
    }
    ms1, err := swimy.New(cfg)
    if err != nil {
    //
    }
    defer ms1.Stop()
    
    // some other app
    ms2, err := swimy.New(nil) 
    if err != nil {
    //
    }
    defer ms2.Stop()
    err := ms2.Join(ctx, "127.0.0.1:54555") // existing known members to join the membership

```

## Decisions
- Failure is immediately disseminated by multicast
  - using gossip style dissemination could have reduced msgs going around, yet reducing convergence speed. 
- Gossip uses randomness rather than round-robin
  - round-robin may be added for probing to reduce time to detect failure 


## TODOs
- Implement lamport clock
- Instrument via otel & slog
- TCP conn. pooling per addr
- Add failure testing 
