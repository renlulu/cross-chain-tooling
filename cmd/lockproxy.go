package cmd

import (
	"encoding/json"
	"github.com/Zilliqa/cross-chain-tooling/util"
	"github.com/Zilliqa/gozilliqa-sdk/account"
	"github.com/Zilliqa/gozilliqa-sdk/crosschain/polynetwork"
	"github.com/Zilliqa/gozilliqa-sdk/provider"
	"github.com/spf13/cobra"
	"log"
	"strings"
)

var lockProxy string
var oldLockProxy string

func init() {
	lockproxyCmd.Flags().StringVarP(&privateKey, "key", "k", "", "private key used to deploy smart contract, also becomes to the admin of the contracts")
	lockproxyCmd.Flags().StringVarP(&lockProxy, "lp", "l", "", "lock proxy address")
	lockproxyCmd.Flags().StringVarP(&oldLockProxy, "olp", "o", "", "old lock proxy address")
	lockproxyCmd.Flags().IntVarP(&chainId, "chainId", "i", 333, "zilliqa chain id")

	RootCmd.AddCommand(lockproxyCmd)

}

var lockproxyCmd = &cobra.Command{
	Use:   "lockproxy",
	Short: "migrate lockproxy and withdraw funds",
	Long:  "migrate lockproxy and withdraw funds",
	Run: func(cmd *cobra.Command, args []string) {
		if privateKey == "" {
			log.Fatalln("private key cannot be empty")
		}

		wallet := account.NewWallet()
		wallet.AddByPrivateKey(privateKey)
		client := provider.NewProvider(host)

		lp := &polynetwork.LockProxy{
			Addr:       lockProxy,
			Wallet:     wallet,
			Client:     client,
			ChainId:    chainId,
			MsgVersion: util.MsgVersion,
		}

		// migrate register
		register, err := client.GetSmartContractSubState(strings.TrimPrefix(oldLockProxy, "0x"), "register", []interface{}{})
		if err != nil {
			log.Fatalf("get register error: %s", err.Error())
		}

		type RegisterResult struct {
			Result map[string]map[string]string `json:"result"`
		}

		var registerResult RegisterResult
		err = json.Unmarshal([]byte(register), &registerResult)
		if err != nil {
			log.Fatalf("unmarshal registerResult error: %s", err.Error())
		}

		registerInnerMap := registerResult.Result["register"]

		log.Println("populate register")
		for asset, hash := range registerInnerMap {
			txn, err := lp.PopulateRegister(asset, hash)
			if err != nil {
				log.Fatalf("PopulateRegister error: %s", err.Error())
			}
			txn.Confirm(txn.ID, 1000, 10, client)
		}

		// migrate nonce
		nonce, err := client.GetSmartContractSubState(strings.TrimPrefix(oldLockProxy, "0x"), "nonce", []interface{}{})
		if err != nil {
			log.Fatalf("get nonce error: %s", err.Error())
		}

		type Nonce struct {
			Result map[string]string `json:"result"`
		}
		var n Nonce
		err = json.Unmarshal([]byte(nonce), &n)
		if err != nil {
			log.Fatalf("unmarshal nonce error: %s", err.Error())
		}

		log.Println("populate nonce")
		txn, err := lp.PopulateNonce(n.Result["nonce"])
		if err != nil {
			log.Fatalf("populate nonce error: %s", err.Error())
		}

		txn.Confirm(txn.ID, 1000, 10, client)

	},
}
