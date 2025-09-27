package main

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	"github.com/virel-project/virel-blockchain/v3/config"
	"github.com/virel-project/virel-blockchain/v3/util"
	"github.com/virel-project/virel-blockchain/v3/wallet"
	"github.com/virel-project/virel-gui/v2/mywidget"
)

type StakingTab struct {
	Wallet *wallet.Wallet

	StakedBalance *mywidget.Card
	UnlockTime    *mywidget.Card
	DelegateId    *mywidget.Card

	SetDelegateBtn *widget.Button
	StakeBtn       *widget.Button
	UnstakeBtn     *widget.Button
}

func CreateStakingTab(wall *wallet.Wallet) *StakingTab {
	st := &StakingTab{
		Wallet: wall,
	}

	st.StakedBalance = mywidget.NewCard(a, theme.Color(theme.ColorNamePrimary),
		util.FormatCoin(wall.GetStakedBalance()), T.StakedBalance, T.StakedBalanceCopied)

	delegateStr := "none"
	if wall.GetDelegateId() != 0 {
		delegateStr = strconv.FormatUint(wall.GetDelegateId(), 10)
	}

	st.DelegateId = mywidget.NewCard(a, theme.Color(theme.ColorNamePrimary),
		delegateStr, T.DelegateId, T.Copied)

	st.SetDelegateBtn = widget.NewButton(T.SetDelegate, st.SetDelegateBtnClicked)

	st.StakeBtn = widget.NewButton(T.Stake, st.StakeBtnClicked)
	st.UnstakeBtn = widget.NewButton(T.Unstake, st.UnstakeBtnClicked)

	stakedUnlockHeight := wall.GetStakedUnlock()
	remainingTime := ""
	if stakedUnlockHeight > wall.GetHeight() {
		remainingTime = (time.Duration(stakedUnlockHeight-wall.GetHeight()) * time.Second * config.TARGET_BLOCK_TIME).String()
	}
	st.UnlockTime = mywidget.NewCard(a, theme.Color(theme.ColorNameButton),
		strconv.FormatUint(stakedUnlockHeight, 10), fmt.Sprintf(T.StakedUnlockHeight, remainingTime), T.StakedUnlockCopied)
	if stakedUnlockHeight > st.Wallet.GetHeight() {
		st.UnlockTime.Show()
	} else {
		st.UnlockTime.Hide()
	}

	return st
}

func (s *StakingTab) Update() {
	fyne.Do(func() {
		s.StakedBalance.SetTitle(util.FormatCoin(s.Wallet.GetStakedBalance()))

		stakedUnlockHeight := s.Wallet.GetStakedUnlock()
		remainingTime := ""
		if stakedUnlockHeight > s.Wallet.GetHeight() {
			remainingTime = (time.Duration(stakedUnlockHeight-s.Wallet.GetHeight()) * time.Second * config.TARGET_BLOCK_TIME).String()
		}
		s.UnlockTime.SetTitle(strconv.FormatUint(stakedUnlockHeight, 10))
		s.UnlockTime.SetComment(fmt.Sprintf(T.StakedUnlockHeight, remainingTime))

		if stakedUnlockHeight > s.Wallet.GetHeight() {
			s.UnlockTime.Show()
		} else {
			s.UnlockTime.Hide()
		}

		delegateStr := "none"
		if s.Wallet.GetDelegateId() != 0 {
			delegateStr = strconv.FormatUint(s.Wallet.GetDelegateId(), 10)
		}
		s.DelegateId.SetTitle(delegateStr)
	})
}

func (s *StakingTab) SetDelegateBtnClicked() {
	delegateid := widget.NewEntry()
	formItems := []*widget.FormItem{
		{Text: T.DelegateId, Widget: delegateid, HintText: "Find a delegate here: https://explorer.virel.org/delegates"},
	}

	delegateid.Validator = func(s string) error {
		_, err := strconv.ParseUint(strings.TrimPrefix(s, config.DELEGATE_ADDRESS_PREFIX), 10, 64)

		if err != nil {
			return errors.New(T.ErrInvalidDelegateId)
		}
		return nil
	}

	dialog.NewForm(T.SetDelegate, T.Confirm, T.Cancel, formItems, func(b bool) {
		if !b {
			return
		}

		delid, err := strconv.ParseUint(strings.TrimPrefix(delegateid.Text, config.DELEGATE_ADDRESS_PREFIX), 10, 64)
		if err != nil {
			fmt.Println(err)
			return
		}

		tx, err := s.Wallet.SetDelegate(delid, s.Wallet.GetDelegateId())
		if err != nil {
			dialog.NewError(err, w).Show()
			return
		}

		res, err := s.Wallet.SubmitTx(tx)
		if err != nil {
			dialog.NewError(err, w).Show()
			return
		}

		InfoDialog(w, T.TransferSuccess, T.TXID+": "+res.TXID.String())
	}, w).Show()
}

func (s *StakingTab) StakeBtnClicked() {
	amount := widget.NewEntry()
	formItems := []*widget.FormItem{
		{Text: T.TransferAmount, Widget: amount},
	}
	amount.Validator = func(s string) error {
		n, err := strconv.ParseFloat(s, 64)
		if err != nil {
			return errors.New(T.InvalidAmount)
		}
		if n < config.MIN_STAKE_AMOUNT/config.COIN {
			return fmt.Errorf("minimum stake amount is %v", config.MIN_STAKE_AMOUNT/config.COIN)
		}
		return nil
	}
	dialog.NewForm(T.Stake, T.Confirm, T.Cancel, formItems, func(b bool) {
		if !b {
			return
		}

		fmt.Println("staking...")

		floatAmt, err := strconv.ParseFloat(amount.Text, 64)
		if err != nil {
			fmt.Println(err)
			dialog.NewError(err, w).Show()
			return
		}

		amt := uint64(floatAmt * config.COIN)

		if amt < config.MIN_STAKE_AMOUNT {
			fmt.Println("stake amount too small")
			dialog.NewError(fmt.Errorf("amount < min stake amount"), w).Show()
			return
		}

		txn, err := s.Wallet.Stake(s.Wallet.GetDelegateId(), amt, s.Wallet.GetStakedUnlock())
		if err != nil {
			fmt.Println(err)
			dialog.NewError(fmt.Errorf("failed to create stake transaction: %w", err), w).Show()
			return
		}

		totalAmt, err := txn.TotalAmount()
		if err != nil {
			fmt.Println(err)
			dialog.NewError(err, w).Show()
			return
		}

		dialog.NewCustomConfirm(T.ConfirmTransfer, T.Confirm, T.Cancel, widget.NewLabel(fmt.Sprintf(T.ConfirmStake, util.FormatCoin(totalAmt))), func(b bool) {
			if !b {
				return
			}
			res, err := s.Wallet.SubmitTx(txn)
			if err != nil {
				dialog.NewError(err, w).Show()
				return
			}

			InfoDialog(w, T.TransferSuccess, T.TXID+": "+res.TXID.String())
		}, w).Show()

	}, w).Show()

}

func (s *StakingTab) UnstakeBtnClicked() {
	amount := widget.NewEntry()
	formItems := []*widget.FormItem{
		{Text: T.TransferAmount, Widget: amount},
	}
	amount.Validator = func(s string) error {
		n, err := strconv.ParseFloat(s, 64)
		if err != nil {
			return errors.New(T.InvalidAmount)
		}
		if n <= 0 {
			return fmt.Errorf("invalid unstake amount")
		}
		return nil
	}
	dialog.NewForm(T.Unstake, T.Confirm, T.Cancel, formItems, func(b bool) {
		if !b {
			return
		}
		floatAmt, err := strconv.ParseFloat(amount.Text, 64)
		if err != nil {
			fmt.Println(err)
			return
		}

		amt := uint64(floatAmt * config.COIN)

		if amt == 0 {
			dialog.NewError(fmt.Errorf("amount can't be zero"), w).Show()
			return
		}

		txn, err := s.Wallet.Unstake(s.Wallet.GetDelegateId(), amt)
		if err != nil {
			dialog.NewError(fmt.Errorf("failed to create stake transaction: %w", err), w).Show()
			return
		}

		totalAmt, err := txn.TotalAmount()
		if err != nil {
			fmt.Println(err)
			dialog.NewError(err, w).Show()
			return
		}

		dialog.NewCustomConfirm(T.ConfirmTransfer, T.Confirm, T.Cancel, widget.NewLabel(fmt.Sprintf(T.ConfirmUnstake, util.FormatCoin(totalAmt))), func(b bool) {
			if !b {
				return
			}
			res, err := s.Wallet.SubmitTx(txn)
			if err != nil {
				dialog.NewError(err, w).Show()
				return
			}

			InfoDialog(w, T.TransferSuccess, T.TXID+": "+res.TXID.String())
		}, w).Show()
	}, w).Show()
}

func (s *StakingTab) Container() *fyne.Container {
	return container.NewPadded(container.NewVBox(
		s.StakedBalance, s.DelegateId, s.UnlockTime,
		layout.NewSpacer(),
		s.SetDelegateBtn, s.StakeBtn, s.UnstakeBtn,
	))
}
