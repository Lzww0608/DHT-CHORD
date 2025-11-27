# CHORD DHT

## Overview

This repository contains an implementation of the **Chord Distributed Hash Table (DHT)**, one of the most popular protocols for decentralized peer-to-peer systems. Chord provides a scalable, robust, and efficient method for distributed key-value storage and retrieval.

## What is Chord?

Chord is a protocol and algorithm for a peer-to-peer distributed hash table. It enables nodes (peers) to locate the node responsible for storing a particular piece of data in a decentralized network, using consistent hashing. Each node and key is assigned an identifier (hash), and keys are stored on their successor node.

**Main features:**
- **Scalability**: Handles thousands of nodes with logarithmic lookup overhead.
- **Robustness**: Automatically adapts to nodes joining and leaving (fault tolerance).
- **Decentralization**: No central coordinator.

## Key Concepts

- **Identifier Circle**: Nodes and keys are arranged in a logical ring (modulo 2^m, where m is the hash size).
- **Finger Table**: Each node maintains a routing table (finger table) to accelerate lookups.
- **Successor/Predecessor**: Every node knows its immediate successor and predecessor for quick join/leave stabilization.

## How it Works

### Node Join

When a new node joins:
1. It calculates its identifier.
2. It finds its position in the identifier circle.
3. Updates the relevant finger tables.

### Lookup

To find the node for key `k`:
1. Start the search at any node.
2. Use the finger table to route queries quickly (in O(log N) hops).
3. The responsible node returns the value or stores the value.

### Node Leave/Failure

- The protocol self-heals by having neighboring nodes update their successor/predecessor pointers and finger tables.
- Periodic stabilization ensures robustness.

## API Overview

Typical Chord DHT exposes:

- `Join(nodeAddress)`: Join the network.
- `Put(key, value)`: Store a key-value pair.
- `Get(key)`: Retrieve the value for a key.
- `Leave()`: Cleanly leave the network.

## Example Usage

```go
// create or join a network
chordNode := chord.NewNode("localhost:8000")
chordNode.Join("seed.node:8000")

// put and get
chordNode.Put("myKey", "myValue")
value, err := chordNode.Get("myKey")
```

## References

- [Chord: A Scalable Peer-to-peer Lookup Protocol for Internet Applications (Stanford paper)](https://pdos.csail.mit.edu/papers/chord:sigcomm01/chord_sigcomm.pdf)
- [Wikipedia: Chord (peer-to-peer)](https://en.wikipedia.org/wiki/Chord_(peer-to-peer))

## Contributing

Pull requests and issues are welcome! Please see [CONTRIBUTING.md](CONTRIBUTING.md) for details.

## License

This project is open source and available under the [MIT License](LICENSE).

