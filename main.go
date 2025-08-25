package main

import (
	"encoding/hex"
	"errors"
	"fmt"
	"net/url"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"
	"virel-gui/lang"
	"virel-gui/mycontainer"
	"virel-gui/mylayout"
	"virel-gui/mywidget"

	"github.com/virel-project/virel-blockchain/address"
	"github.com/virel-project/virel-blockchain/config"
	"github.com/virel-project/virel-blockchain/logger"
	"github.com/virel-project/virel-blockchain/transaction"
	"github.com/virel-project/virel-blockchain/util"
	"github.com/virel-project/virel-blockchain/util/updatechecker"
	"github.com/virel-project/virel-blockchain/wallet"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	"github.com/Xuanwo/go-locale"
	"golang.org/x/text/language"
)

const VERSION_MAJOR = 1
const VERSION_MINOR = 3
const VERSION_PATCH = 1

const TICKER = "VRL"

var log = logger.New()

var T = lang.Lang[language.English]

var nodeManager *NodeManger

func init() {
	tags, err := locale.DetectAll()
	if err != nil {
		fmt.Println("could not detect language:", err)
		tags = []language.Tag{}
	}

	fmt.Println("tags:", tags)

	for _, v := range tags {
		if lang.Lang[v] != nil {
			fmt.Println("found language:", v)
			T = lang.Lang[v]
		} else if lang.Lang[v.Parent()] != nil {
			fmt.Println("found language:", v.Parent())
			T = lang.Lang[v.Parent()]
		}
	}

	rpcUrls, err := os.ReadFile("rpc-urls.txt")
	if err != nil {
		rpcUrls = []byte(RPC_URLS)
	}
	nodeManager = NewNodeManager(string(rpcUrls))
}

const width_limit = 750

var w fyne.Window
var a fyne.App

func main() {
	detectedLang, err := lang.DetectIETF()
	if detectedLang == "" || err != nil {
		fmt.Println("LANG environment variable is not set.", err)
	}
	fmt.Println("LANG is:", detectedLang)
	if len(detectedLang) > 0 {
		tag, err := language.Parse(detectedLang)
		if err != nil {
			fmt.Println("failed to parse language:", err)
		}
		T = lang.GetTranslation(tag)
	}
	//T = lang.Lang[language.]

	a = app.New()
	w = a.NewWindow(fmt.Sprintf("Virel GUI v%d.%d.%d", VERSION_MAJOR, VERSION_MINOR, VERSION_PATCH))

	w.Resize(fyne.NewSize(800, 600))

	pageHome()

	w.ShowAndRun()
}

func pageHome() {
	title := widget.NewRichTextFromMarkdown("# Virel GUI")

	btnOpen := widget.NewButton(T.OpenWallet, func() {
		fmt.Println("click open!")
		pageOpen()
	})
	btnCreate := widget.NewButton(T.CreateWallet, func() {
		fmt.Println("click create!")
		pageCreate()
	})
	btnRestore := widget.NewButton(T.RestoreFromSeed, func() {
		fmt.Println("click restore!")
		pageRestore()
	})

	fr := container.NewVBox(
		title,
		btnOpen, btnCreate, btnRestore,
	)
	fr.Layout = layout.NewCustomPaddedVBoxLayout(16)

	fr = mycontainer.NewLimiter(width_limit, 800, fr)

	options := make([]string, len(lang.Lang))

	for _, v := range lang.Lang {
		options = append(options, v.Language)
	}

	chooseLanguage := widget.NewSelect(options, func(s string) {
		if s == T.Language {
			return
		}
		fmt.Println("language changed:", s)

		for _, v := range lang.Lang {
			if v.Language == s {
				T = v
				pageHome()
			}
		}

	})
	chooseLanguage.SetSelected(T.Language)

	body := container.NewStack(fr, container.NewVBox(layout.NewSpacer(), container.NewHBox(chooseLanguage, layout.NewSpacer())))

	w.SetContent(body)

	go func() {
		status, remoteVersion, err := updatechecker.CheckForUpdate("https://api.github.com/repos/virel-project/virel-gui/releases/latest",
			VERSION_MAJOR, VERSION_MINOR, VERSION_PATCH)
		if err != nil || status == updatechecker.StatusError {
			log.Err(err)
			return
		}
		if status != updatechecker.StatusUpToDate {
			url, err := url.Parse("https://virel.org/docs/info/links/#wallets")
			if err != nil {
				log.Err(err)
				return
			}
			content := container.NewVBox(
				widget.NewLabel(fmt.Sprintf(T.UpdateRequired, remoteVersion)),
				widget.NewHyperlink(url.String(), url),
			)
			CustomDialog(w, T.UpdateGui, T.Ok, content)
		}
	}()
}

func pageOpen() {
	title := widget.NewRichTextFromMarkdown("## " + T.OpenWallet)

	walls, err := getWallets()
	if err != nil {
		ErrorDialog(w, fmt.Errorf("failed to get wallet list: %w", err))
		return
	}

	walletName := widget.NewSelect(walls, func(s string) {})
	walletPass := widget.NewPasswordEntry()

	form := widget.NewForm(widget.NewFormItem(T.WalletName, walletName),
		widget.NewFormItem(T.Password, walletPass))

	form.SubmitText = T.OpenWallet
	form.CancelText = T.Cancel
	form.OnSubmit = func() {
		loadingPage(T.LoadingWallet)
		go func() {
			ci := walletName.Selected
			fileContent, err := os.ReadFile(ci + ".keys")
			if err != nil {
				ErrorDialog(w, fmt.Errorf("failed to open wallet: %w", err))
				pageOpen()
				return
			}

			wall, err := wallet.OpenWallet(nodeManager.Urls()[0], fileContent, []byte(walletPass.Text))
			if err != nil {
				ErrorDialog(w, fmt.Errorf("failed to open wallet: %w", err))
				pageOpen()
				return
			}

			pageWallet(wall)
		}()

	}
	form.OnCancel = func() {
		pageHome()
	}

	fyne.Do(func() {
		w.SetContent(mycontainer.NewWidthLimiter(width_limit, container.NewVBox(
			title, form,
		)))
	})
}

func pageCreate() {
	title := widget.NewRichTextFromMarkdown("## " + T.CreateWallet)

	walletName := widget.NewEntry()
	walletPass := widget.NewPasswordEntry()
	walletPass2 := widget.NewPasswordEntry()

	form := widget.NewForm(widget.NewFormItem(T.WalletName, walletName),
		widget.NewFormItem(T.Password, walletPass),
		widget.NewFormItem(T.RepeatPassword, walletPass2))

	form.CancelText = T.Cancel
	form.OnCancel = func() {
		pageHome()
	}

	form.SubmitText = T.CreateWallet
	form.OnSubmit = func() {
		var err error
		if walletPass.Text != walletPass2.Text {
			err = errors.New(T.PasswordNotMatch)
		} else if len(walletPass.Text) < 5 {
			err = errors.New(T.PasswordTooShort)
		} else if len(walletName.Text) > 50 {
			err = errors.New(T.ErrWalletNameTooLong)
		} else if len(walletName.Text) < 1 {
			err = errors.New(T.ErrWalletNameTooShort)
		} else if strings.Contains(walletName.Text, ".") {
			err = errors.New(T.ErrWalletNameInvalid)
		}
		wallName := url.PathEscape(walletName.Text)

		if err != nil {
			ErrorDialog(w, fmt.Errorf("cannot create wallet: %w", err))
			walletPass.SetText("")
			walletPass2.SetText("")
			return
		}

		loadingPage(T.CreatingWallet)

		go func() {
			wall, db, err := wallet.CreateWallet(nodeManager.Urls()[0], []byte(walletPass.Text), false)
			if err != nil {
				ErrorDialog(w, fmt.Errorf("failed to create wallet: %w", err))
			}

			saveWallet(wallName, db)

			pageWallet(wall)
			displaySeedDialog(wall)
		}()

	}

	w.SetContent(mycontainer.NewWidthLimiter(width_limit, container.NewVBox(
		title, form,
	)))
}

func pageRestore() {
	title := widget.NewRichTextFromMarkdown("## " + T.RestoreFromSeed)
	walletSeed := widget.NewEntry()
	walletName := widget.NewEntry()
	walletPass := widget.NewPasswordEntry()
	walletPass2 := widget.NewPasswordEntry()

	form := widget.NewForm(widget.NewFormItem(T.Seed, walletSeed),
		widget.NewFormItem(T.WalletName, walletName),
		widget.NewFormItem(T.Password, walletPass),
		widget.NewFormItem(T.RepeatPassword, walletPass2))
	form.CancelText = T.Cancel
	form.OnCancel = func() {
		pageHome()
	}

	form.SubmitText = T.RestoreFromSeed
	form.OnSubmit = func() {
		loadingPage(T.LoadingWallet)
		go func() {
			filename := url.PathEscape(walletName.Text) + ".keys"

			_, err := os.Stat(filename)
			if err == nil {
				ErrorDialog(w, fmt.Errorf("wallet %v already exists", filename))
				pageOpen()
				return
			}
			wall, err := wallet.CreateWalletFileFromMnemonic(nodeManager.Urls()[0], filename, walletSeed.Text, []byte(walletPass.Text))
			if err != nil {
				ErrorDialog(w, fmt.Errorf("failed to open wallet: %w", err))
				pageOpen()
				return
			}

			pageWallet(wall)
		}()

	}

	w.SetContent(container.NewVBox(title, form))

}

func loadingPage(txt string) {
	fmt.Println("loadingPage", txt)
	w.SetContent(container.NewCenter(widget.NewRichTextFromMarkdown("## " + txt)))
}

var numRegex = regexp.MustCompile(`^([0-9]*[.])?[0-9]+$`)

func pageWallet(wall *wallet.Wallet) {
	yourBalance := mywidget.NewCard(a, theme.Color(theme.ColorNamePrimary),
		util.FormatCoin(wall.GetBalance()), T.Balance, T.BalanceCopied)
	yourAddress := mywidget.NewCard(a, theme.Color(theme.ColorNameButton), wall.GetAddress().String(),
		T.Address, T.AddressCopied)

	cardsGrid := container.New(mylayout.NewWrapLayout(800),
		yourBalance, yourAddress)

	myWallet := container.NewPadded(container.NewVBox(
		cardsGrid,
		layout.NewSpacer(),
		// widget.NewLabel(T.RecentTransactions),
		//list,
	))

	recipient := widget.NewEntry()
	recipient.Validator = func(s string) error {
		if len(s) < 5 {
			return errors.New(T.InvalidWallet)
		}
		_, err := address.FromString(s)
		if err != nil {
			return errors.New(T.InvalidWallet)
		}
		return nil
	}

	amount := widget.NewEntry()
	amount.Validator = func(s string) error {
		if !numRegex.MatchString(s) {
			return errors.New(T.InvalidAmount)
		}
		return nil
	}

	recipient.Enable()
	amount.Enable()

	sendForm := &widget.Form{
		Items: []*widget.FormItem{ // we can specify items in the constructor
			{Text: T.Recipient, Widget: recipient},
			{Text: T.TransferAmount, Widget: amount},
		},

		SubmitText: T.Transfer,
	}
	sendForm.OnSubmit = func() { // optional, handle form submission

		log.Info("Form submitted:", recipient.Text)
		log.Info("amount:", amount.Text)

		walletPass := widget.NewEntry()
		walletPass.Password = true
		walletPass.Validator = func(s string) error {
			if len(s) == 0 {
				return errors.New(T.FieldRequired)
			}
			return nil
		}

		dlg := dialog.NewCustomConfirm(T.TransferConfirm, T.Confirm, T.Cancel, container.NewVBox(
			widget.NewLabel(T.ReviewTransferDetails),
			widget.NewLabel(
				fmt.Sprintf(T.TransferFields, amount.Text, recipient.Text),
			),
			widget.NewForm(widget.NewFormItem(T.Password, walletPass)),
		), func(ok bool) {
			if !ok {
				return
			}

			if walletPass.Text != string(wall.GetPassword()) {
				ErrorDialog(w, errors.New(T.PasswordNotMatch))
				return
			}

			amt, err := strconv.ParseFloat(amount.Text, 64)
			if err != nil {
				ErrorDialog(w, errors.New(T.InvalidAmount))
				return
			}
			if amt <= 0 {
				ErrorDialog(w, errors.New(T.InvalidAmount))
				return
			}

			recv, err := address.FromString(recipient.Text)
			if err != nil {
				ErrorDialog(w, errors.New(T.InvalidWallet))
				return
			}

			amtInt := uint64(amt * config.COIN)

			txn, err := wall.Transfer([]transaction.Output{
				{
					Amount:    amtInt,
					Recipient: recv.Addr,
					PaymentId: recv.PaymentId,
				},
			})
			if err != nil {
				ErrorDialog(w, fmt.Errorf(T.FailedToCreateTx, err))
				return
			}

			dialog.NewCustomConfirm(T.ConfirmTransfer, T.Confirm, T.Cancel, container.NewVBox(
				widget.NewLabel(T.Recipient+": "+recv.String()),
				widget.NewLabel(T.TransferAmount+": "+util.FormatCoin(amtInt)+" "+TICKER),
				widget.NewLabel(T.TxFee+": "+util.FormatCoin(txn.Fee)+" "+TICKER),
				widget.NewLabel(T.TXID+": "+txn.Hash().String()),
			), func(ok bool) {
				if !ok {
					return
				}

				res, err := wall.SubmitTx(txn)
				if err != nil {
					ErrorDialog(w, fmt.Errorf(T.FailedToSubmitTx, err))
					return
				}

				InfoDialog(w, T.TransferSuccess, T.TXID+": "+res.TXID.String())

			}, w).Show()

		}, w)

		dlg.Show()

	}

	sendTitle := NewTitle(T.Transfer)
	mySend := container.NewVBox(sendTitle, sendForm)

	/*historyTitle := canvas.NewText(T.TabHistory, theme.Color(theme.ColorNameForeground))
	historyTitle.TextStyle.Bold = true
	historyTitle.TextSize = theme.TextSubHeadingSize()
	historyTitle.Alignment = fyne.TextAlignCenter*/

	txlist := TxList{
		List: make([]HistoryObject, 0),
	}

	const time_format = "2006-01-02 15:04"

	historyList := widget.NewTableWithHeaders(
		func() (int, int) {
			return len(txlist.List), 4
		},
		func() fyne.CanvasObject {
			return widget.NewLabel(hex.EncodeToString(make([]byte, 32)))
		},
		func(i widget.TableCellID, co fyne.CanvasObject) {
			//lbl := co.(*fyne.Container).Objects[0].(*widget.Label)
			lbl := co.(*widget.Label)

			x := txlist.List[i.Row]

			switch i.Col {
			case 0: // time
				lbl.SetText(x.Time.Format(time_format))
			case 1: // txid
				lbl.SetText(x.TXID)
			case 2: // amount
				amtStr := strconv.FormatFloat(x.Amount, 'f', int(config.ATOMIC), 64)
				if amtStr[0] != '-' {
					amtStr = "+" + amtStr
				}
				lbl.SetText(amtStr)
			case 3: // confs
				lbl.SetText(strconv.FormatUint(wall.GetHeight()-x.Height, 10))
			}
		},
	)
	historyList.ShowHeaderColumn = false
	historyList.ShowHeaderRow = true

	timeWidth := widget.NewLabel(time_format).MinSize().Width
	amtWidth := widget.NewLabel("+100.000000000").MinSize().Width

	historyList.SetColumnWidth(0, timeWidth)
	historyList.SetColumnWidth(2, amtWidth)
	historyList.SetColumnWidth(3, timeWidth)

	historyList.CreateHeader = func() fyne.CanvasObject {
		lbl := widget.NewLabel("")
		lbl.Truncation = fyne.TextTruncateEllipsis
		return lbl
	}
	historyList.UpdateHeader = func(id widget.TableCellID, template fyne.CanvasObject) {
		lbl := template.(*widget.Label)

		switch id.Col {
		case 0:
			lbl.SetText(T.Time)
		case 1:
			lbl.SetText(T.TXID)
		case 2:
			lbl.SetText(T.TransferAmount)
		case 3:
			lbl.SetText(T.Confirmations)
		}
	}

	seedBtn := widget.NewButton(T.DisplaySeed, func() {
		passEntry := widget.NewEntry()
		passEntry.Password = true

		content := container.NewVBox(passEntry)

		d := dialog.NewCustomConfirm(T.InputPassword, T.ViewSeed, T.Cancel, content, func(confirm bool) {
			if !confirm {
				return
			}

			if passEntry.Text != string(wall.GetPassword()) {
				ErrorDialog(w, errors.New(T.PasswordNotMatch))
				return
			}

			displaySeedDialog(wall)
		}, w)
		d.Show()
	})

	nodeLbl := widget.NewLabel(T.NodeAddress + ": " + wall.GetRpcDaemonAddress())

	changeNodeBtn := widget.NewButton(T.ChangeNode, func() {
		nodeUrlInput := widget.NewEntry()
		nodeUrlInput.SetText(strings.Join(nodeManager.rpcUrls, ";"))

		Dialog(w, T.ChangeNode, T.Ok, T.Cancel, nodeUrlInput, func(b bool) {
			if !b {
				return
			}
			nodeManager.rpcUrls = strings.Split(nodeUrlInput.Text, ";")
			os.WriteFile("rpc-urls.txt", []byte(nodeUrlInput.Text), 0o644)
			nodeLbl.SetText(T.NodeAddress + ": " + wall.GetRpcDaemonAddress())
		})
	})

	settingsCont := container.NewVBox(NewTitle(T.Settings), seedBtn, nodeLbl, changeNodeBtn)

	tabs := container.NewAppTabs(
		container.NewTabItemWithIcon(T.TabHome, theme.HomeIcon(), myWallet),
		container.NewTabItemWithIcon(T.TabTransfer, theme.MailSendIcon(), mySend),
		container.NewTabItemWithIcon(T.TabHistory, theme.HistoryIcon(), historyList),
		container.NewTabItemWithIcon(T.Settings, theme.SettingsIcon(), settingsCont),
	)

	statusLabel := widget.NewLabel(T.StatusConnected)
	statusBar := mywidget.NewBar(theme.Color(theme.ColorNameHeaderBackground), statusLabel)

	go func() {
		for {
			err := nodeManager.Refresh(wall)
			fyne.Do(func() {
				if err != nil {
					statusLabel.SetText(T.StatusError)
				} else {
					statusLabel.SetText(T.StatusConnected + " " + wall.GetRpcDaemonAddress())
				}

				yourBalance.SetTitle(util.FormatCoin(wall.GetBalance()))
			})
			if err != nil {
				fmt.Println("failed to refresh:", err)
				time.Sleep(10 * time.Second)
				continue
			}
			err = txlist.Refresh(wall)
			if err != nil {
				fmt.Println("error fetching tx list:", err)
			}

			time.Sleep(10 * time.Second)
		}
	}()

	fyne.DoAndWait(func() {
		w.SetContent(container.NewBorder(nil, statusBar, nil, nil, tabs))
	})
}

func displaySeedDialog(wall *wallet.Wallet) {
	confirmBtn := widget.NewButton(T.Confirm, nil)
	checkbox := widget.NewCheck(T.UnderstandSeed, func(b bool) {
		if b {
			confirmBtn.Enable()
		} else {
			confirmBtn.Disable()
		}
	})

	mnemonic := widget.NewEntry()
	mnemonic.SetText(wall.GetMnemonic())
	mnemonic.OnChanged = func(_ string) {
		mnemonic.SetText(wall.GetMnemonic())
	}

	content := container.NewVBox(
		widget.NewRichTextFromMarkdown(T.YourSeedIs),
		mnemonic,
		widget.NewLabel(T.StoreSeedSafely),
		checkbox,
		confirmBtn,
	)

	d := dialog.NewCustomWithoutButtons(T.ViewSeed, content, w)

	confirmBtn.OnTapped = func() {
		d.Hide()
	}

	d.Show()
}

func saveWallet(walletName string, db []byte) {
	os.WriteFile(walletName+".keys", db, 0o660)
}
func getWallets() ([]string, error) {
	entries, err := os.ReadDir(".")
	if err != nil {
		return nil, err
	}
	var walls = []string{}
	for _, v := range entries {
		if v.IsDir() || len(v.Name()) < 6 {
			continue
		}
		if v.Name()[len(v.Name())-5:] != ".keys" {
			continue
		}
		n := v.Name()[:len(v.Name())-5]
		fmt.Println("correct name!", n)
		walls = append(walls, n)
	}
	return walls, nil
}
