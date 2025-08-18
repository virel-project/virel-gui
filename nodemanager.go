package main

import (
	"strings"

	"github.com/virel-project/virel-blockchain/wallet"
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
		for _, v := range n.rpcUrls {
			if v == origUrl {
				break
			}

			w.SetRpcDaemonAddress(v)
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
