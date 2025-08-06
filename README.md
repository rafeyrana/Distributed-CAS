# Distributed-CAS

Distributed-CAS is a distributed content addressable storage (CAS) system implemented in Go, designed for peer-to-peer networks. It enables secure file storage and retrieval across connected nodes using TCP communication. The system emphasizes efficiency, security, and scalability, making it suitable for various applications.


## Features

The following features have been implemented from scratch:

1. **TCP Transport Connection**  
   Establishes reliable connections between nodes using TCP, including a handshake protocol for secure communication and data encoding for message integrity.

2. **Key Storage**  
Supports operations for file management in both local and distributed contexts. Key functionalities include reading, writing, deleting, and checking the existence of files, allowing for efficient data retrieval.

3. **File Server**  
Manages file distribution and retrieval from connected peers. The server architecture ensures that files can be accessed seamlessly across the network, promoting redundancy and fault tolerance.

4. **File Encryption**  
Incorporates encryption mechanisms to secure files during transit and storage.

5. **Per-Key TTL Support**
Allows setting a time-to-live (TTL) for each key, after which the file is automatically deleted from the system.

## Future Work:
1. Benchmarking
2. Support for other transport protocols (e.g., WebSockets)
3. Implementing a distributed hash table (DHT) for peer discovery and routing.
