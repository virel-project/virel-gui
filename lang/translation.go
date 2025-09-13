package lang

import (
	"embed"
	"reflect"
	"runtime"
	"strings"

	"github.com/BurntSushi/toml"
	"golang.org/x/text/language"
)

type Translation struct {
	Language string

	Ok, Cancel, Back      string
	Confirm               string
	CreateWallet          string
	OpenWallet            string
	RestoreFromSeed       string
	WalletName            string
	ErrWalletNameTooLong  string
	ErrWalletNameTooShort string
	ErrWalletNameInvalid  string
	Password              string
	PasswordTooShort      string
	RepeatPassword        string
	PasswordNotMatch      string
	Seed                  string
	DisplaySeed           string
	TransferAmount        string
	Recipient             string
	Transfer              string
	TabHome               string
	TabTransfer           string
	TabHistory            string
	Settings              string
	Balance               string
	BalanceCopied         string
	StakedBalance         string
	StakedBalanceCopied   string
	Address               string
	AddressCopied         string
	ConfirmTransfer       string
	LoadingWallet         string
	CreatingWallet        string
	ViewSeed              string
	InputPassword         string
	YourSeedIs            string
	StoreSeedSafely       string
	UnderstandSeed        string

	TransferConfirm       string
	ReviewTransferDetails string
	TransferFields        string
	RecentTransactions    string

	InvalidAmount    string
	InvalidWallet    string
	FailedToCreateTx string
	FailedToSubmitTx string
	TransferSuccess  string
	FieldRequired    string

	TXID  string
	TxFee string

	NodeAddress string
	ChangeNode  string

	StatusConnected string
	StatusError     string

	Time          string
	Confirmations string

	UpdateGui      string
	UpdateRequired string
}

var Lang = map[language.Tag]*Translation{}

//go:embed translations/*.toml
var translationFiles embed.FS

func init() {
	allFiles, err := translationFiles.ReadDir("translations")
	if err != nil {
		panic(err)
	}
	for _, v := range allFiles {
		file, err := translationFiles.Open("translations/" + v.Name())
		if err != nil {
			panic(err)
		}

		if runtime.GOOS == "js" && v.Name() == "zh-cn.toml" {
			continue
		}

		languageName, err := language.Parse(strings.Split(v.Name(), ".")[0])
		if err != nil {
			panic(err)
		}

		decoder := toml.NewDecoder(file)
		Lang[languageName] = &Translation{}
		_, err = decoder.Decode(Lang[languageName])
		if err != nil {
			panic(err)
		}
	}

	en := Lang[language.English]
	for i, v := range Lang {
		if i == language.English {
			continue
		}

		x := reflect.ValueOf(v)
		xEn := reflect.ValueOf(en)
		xnum := x.Elem().NumField()

		for i := 0; i < xnum; i++ {
			va := x.Elem().Field(i).String()
			if va == "" {
				x.Elem().Field(i).SetString(xEn.Elem().Field(i).String())
			}
		}
	}
}

func GetTranslation(l language.Tag) *Translation {
	for i, v := range Lang {
		if i == l {
			return v
		}
	}
	for i, v := range Lang {
		ibase, _ := i.Base()
		base, _ := l.Base()
		if ibase == base {
			return v
		}
	}

	return Lang[language.English]
}
