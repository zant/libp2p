package main

import (
  "context"
  "fmt"
  "os"
  "os/signal"
  "syscall"

  "github.com/libp2p/go-libp2p"
  peerstore "github.com/libp2p/go-libp2p-core/peer"
  "github.com/libp2p/go-libp2p/p2p/protocol/ping"
  multiaddr "github.com/multiformats/go-multiaddr"
)

func main() {
  ctx := context.Background()

  node, err := libp2p.New(
          ctx,
          libp2p.ListenAddrStrings("/ip4/127.0.0.1/tcp/0"),
          libp2p.Ping(false),
  )
  if err != nil {
    panic(err)
  }

  pingService := &ping.PingService{Host : node}
  node.SetStreamHandler(ping.ID, pingService.PingHandler)

  peerInfo := peerstore.AddrInfo{
    ID: node.ID(),
    Addrs: node.Addrs(),
  }
  addrs, err := peerstore.AddrInfoToP2pAddrs(&peerInfo)
  fmt.Println("libp2p node address: ", addrs[0])

  if len(os.Args) > 1 {
    addr, err := multiaddr.NewMultiaddr(os.Args[1])
    if err != nil {
      panic(err)
    }
    peer, err := peerstore.AddrInfoFromP2pAddr(addr)
    if err != nil {
      panic(err)
    }
    if err := node.Connect(ctx, *peer); err != nil {
      panic(err)
    }
    fmt.Println("sending 5 ping messages to", addr)
    ch := pingService.Ping(ctx, peer.ID)
    for i := 0; i < 5; i++ {
      res := <-ch
      fmt.Println("pinged", addr, "in", res.RTT)
    }
  } else {
    ch := make(chan os.Signal, 1)
    signal.Notify(ch, syscall.SIGINT, syscall.SIGTERM)
    <-ch
    fmt.Println("Received signal, shutting down...")
  }


  if err := node.Close(); err != nil {
    panic(err)
  }
}
