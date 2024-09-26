This qproxy provide a safe and encrypted proxy channel with socks5, the transport layer use quic, and the payloads are encrypted by chacha20.

  +---------+        +-----------+         +-----------+
  | Browser |        |  socks5   |         |  github   |
  +---------+        +-----------+         +-----------+
       |             +-----------+              /|\
       |             |  chacha20 |               |
      \|/            +-----------+               |
  +---------+        +-----------+         +-----------+
  | Client  | <----- |  quic     | ------->|  server   |
  +---------+        +-----------+         +-----------+


              
              
                  
