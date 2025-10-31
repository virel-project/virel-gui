package main

import (
	"fmt"
	"strings"

	"github.com/virel-project/virel-blockchain/v3/wallet"
)

type NodeManger struct {
	rpcUrls []string
}

func NewNodeManager(rpcs string) *NodeManger {
	if len(rpcs) == 0 {
		rpcs = RPC_URLS
	}

	rpcUrls := strings.Split(rpcs, ";")
	return &NodeManger{
		rpcUrls: rpcUrls,
	}
}

func (n *NodeManger) Refresh(w *wallet.Wallet) error {
	err := w.Refresh()
	origUrl := w.GetRpcDaemonAddress()
	if err != nil {
		next := false
		for _, v := range n.rpcUrls {
			if v == origUrl {
				next = true
				continue
			}

			if next {
				fmt.Println("trying with node", v)
				w.SetRpcDaemonAddress(v)
				err = w.Refresh()
				if err == nil {
					return nil
				} else {
					log.Warn(err)
				}
				break
			}
		}
		if !next && len(n.rpcUrls) > 1 {
			fmt.Println("trying with first node", n.rpcUrls[0])
			w.SetRpcDaemonAddress(n.rpcUrls[0])
			err = w.Refresh()
			if err == nil {
				return nil
			} else {
				log.Warn(err)
			}
		}
	}
	return err
}

func (n *NodeManger) Urls() []string {
	return n.rpcUrls
}
