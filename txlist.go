package main

import (
	"cmp"
	"encoding/hex"
	"fmt"
	"slices"
	"time"

	"github.com/virel-project/virel-blockchain/config"
	"github.com/virel-project/virel-blockchain/wallet"
)

type HistoryObject struct {
	TXID   string
	Time   time.Time
	Amount float64
	Fee    float64
	Height uint64
}

type TxList struct {
	List      []HistoryObject
	refreshed bool
}

func (t *TxList) Refresh(wall *wallet.Wallet) error {
	height := wall.GetHeight()
	err := t.refreshTxs(wall, height, false, !t.refreshed, 0)
	if err != nil {
		return err
	}
	err = t.refreshTxs(wall, height, true, !t.refreshed, 0)
	if err != nil {
		return err
	}

	t.refreshed = true

	return nil
}

func (t *TxList) refreshTxs(wall *wallet.Wallet, height uint64, inc, fullscan bool, page int) error {
	txns, err := wall.GetTransactions(inc, 0)
	if err != nil {
		return err
	}

	if fullscan && page < int(txns.MaxPage) {
		defer func() {
			t.refreshTxs(wall, height, inc, fullscan, page+1)
		}()
	}

	updated := false
	for _, v := range txns.Transactions {
		strhash := hex.EncodeToString(v[:])
		notfound := true
		for _, tx := range t.List {
			if strhash == tx.TXID {
				notfound = false
				break
			}
		}
		if notfound {
			tx, err := wall.GetTransaction(v)

			if err != nil {
				log.Warn(err)
				return err
			}

			txTime := time.Now().Unix()

			if tx.Height != 0 && tx.Height < height {
				deltaHeight := int64(height) - int64(tx.Height)
				txTime -= config.TARGET_BLOCK_TIME * deltaHeight
			}

			amt := float64(tx.TotalAmount-tx.Fee) / config.COIN
			if tx.Sender != nil && tx.Sender.Addr == wall.GetAddress().Addr {
				amt = -amt - (float64(tx.Fee) / config.COIN)
			}

			t.List = append(t.List, HistoryObject{
				TXID:   strhash,
				Time:   time.Unix(txTime, 0),
				Amount: amt,
				Fee:    float64(tx.Fee) / config.COIN,
				Height: tx.Height,
			})
			updated = true
		}
	}
	if updated {
		fmt.Println("found new transaction, inc:", inc)
		slices.SortStableFunc(t.List, func(i, j HistoryObject) int {
			return cmp.Compare(j.Height, i.Height)
		})
	}

	return nil
}
