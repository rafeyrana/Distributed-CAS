# Distributed-CAS
Distributed-CAS is a distributed content addressable storage (CAS) system implemented in Go, designed for peer-to-peer networks. It enables secure file storage and retrieval across connected nodes using TCP communication.

## 
Following features have been implemented from scratch:
1. TCP Transport Connection (handshake and encoding)
2. Key Storage (Read, Write, Delete, Exists)
3. File Server: Manages the file distribution and retrieval from connected peers.
4. File Encryption