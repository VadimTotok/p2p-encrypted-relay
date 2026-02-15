# p2p-encrypted-relay

Minimalistic peer-to-peer service for one-time encrypted message delivery.  
Messages are stored only in memory, removed after first read or TTL expiration, and can be retrieved from peer nodes if not found locally.
---

## Features
- One-time read messages (drop deleted after retrieval)
- TTL-based ephemeral storage
- Simple P2P relay to fetch drops from peers
- Written in Go
- Unit tests included
---
