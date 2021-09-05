package cmd

import (
	"encoding/json"
	"fmt"
	"github.com/Zilliqa/cross-chain-tooling/util"
	"github.com/Zilliqa/gozilliqa-sdk/account"
	"github.com/Zilliqa/gozilliqa-sdk/crosschain/polynetwork"
	"github.com/Zilliqa/gozilliqa-sdk/provider"
	"github.com/spf13/cobra"
	"log"
	"strconv"
	"strings"
)

var oldCcmAddr string
var ccmProxyAddr string
var ccmAddr string

func init() {
	migrationCmd.Flags().StringVarP(&privateKey, "key", "k", "", "private key used to deploy smart contract, also becomes to the admin of the contracts")
	migrationCmd.Flags().StringVarP(&host, "api", "a", "https://dev-api.zilliqa.com", "zilliqa api endpoint")
	migrationCmd.Flags().StringVarP(&oldCcmAddr, "oldCcmProxy", "o", "", "old cross chain manager proxy contract address")
	migrationCmd.Flags().StringVarP(&ccmProxyAddr, "ccmProxy", "p", "", "cross chain manager proxy contract address")
	migrationCmd.Flags().StringVarP(&ccmAddr, "ccm", "c", "", "cross chain manager contract address")

	RootCmd.AddCommand(migrationCmd)
}

var migrationCmd = &cobra.Command{
	Use:   "migrate",
	Short: "migrate contract state",
	Long:  "migrate contract state",
	Run: func(cmd *cobra.Command, args []string) {
		if privateKey == "" {
			log.Fatalln("private key cannot be empty")
		}

		wallet := account.NewWallet()
		wallet.AddByPrivateKey(privateKey)
		client := provider.NewProvider(host)

		p := &polynetwork.Proxy{
			ProxyAddr:  ccmProxyAddr,
			ImplAddr:   ccmAddr,
			Wallet:     wallet,
			Client:     client,
			ChainId:    chainId,
			MsgVersion: util.MsgVersion,
		}
		fmt.Println(p)
		// current we skip follow transitions:
		// PopulateWhiteListFromContract
		// PopulateWhiteListToContract
		// PopulateWhiteListMethod

		// migrate conKeepersPublicKeyList
		conKeepersPublicKeyList, err := client.GetSmartContractSubState(strings.TrimPrefix(oldCcmAddr, "0x"), "conKeepersPublicKeyList", []interface{}{})
		if err != nil {
			log.Fatalf("get conKeepersPublicKeyList error: %s", err.Error())
		}

		type Keepers struct {
			Result map[string][]string `json:"result"`
		}

		var keepers Keepers
		err = json.Unmarshal([]byte(conKeepersPublicKeyList), &keepers)
		if err != nil {
			log.Fatalf("unmarshal conKeepersPublicKeyList error: %s", err.Error())
		}
		log.Println("populate conKeepersPublicKeyList")
		txn, err := p.PopulateConKeepersPublicKeyList(keepers.Result["conKeepersPublicKeyList"])
		if err != nil {
			log.Fatalf("populate conKeepersPublicKeyList error: %s", err.Error())
		}

		txn.Confirm(txn.ID, 1000, 10, client)

		// migrate curEpochStartHeight
		curEpochStartHeightResult, err := client.GetSmartContractSubState(strings.TrimPrefix(oldCcmAddr, "0x"), "curEpochStartHeight", []interface{}{})
		if err != nil {
			log.Fatalf("get curEpochStartHeight error: %s", err.Error())
		}
		type CurEpochStartHeight struct {
			Result map[string]string `json:"result"`
		}
		var curEpochStartHeight CurEpochStartHeight
		err = json.Unmarshal([]byte(curEpochStartHeightResult), &curEpochStartHeight)
		if err != nil {
			log.Fatalf("unmarshal conKeepersPublicKeyList error: %s", err.Error())
		}
		log.Println("populate curEpochStartHeight")
		txn, err = p.PopulateCurEpochStartHeight(curEpochStartHeight.Result["curEpochStartHeight"])
		if err != nil {
			log.Fatalf("populate conKeepersPublicKeyList error: %s", err.Error())
		}
		txn.Confirm(txn.ID, 1000, 10, client)

		// migration zilToPolyTxHashIndex
		zilToPolyTxHashIndexResult, err := client.GetSmartContractSubState(strings.TrimPrefix(oldCcmAddr, "0x"), "zilToPolyTxHashIndex", []interface{}{})
		if err != nil {
			log.Fatalf("get zilToPolyTxHashIndex error: %s", err.Error())
		}
		type ZilToPolyTxHashIndexResult struct {
			Result map[string]string `json:"result"`
		}
		var zilToPolyTxHashIndex ZilToPolyTxHashIndexResult
		fmt.Println(zilToPolyTxHashIndexResult)
		err = json.Unmarshal([]byte(zilToPolyTxHashIndexResult), &zilToPolyTxHashIndex)
		if err != nil {
			log.Fatalf("unmarshal zilToPolyTxHashMap error: %s", err.Error())
		}

		log.Println("populate zilToPolyTxHashIndex")
		txn, err = p.PopulateZilToPolyTxHashIndex(zilToPolyTxHashIndex.Result["zilToPolyTxHashIndex"])
		if err != nil {
			log.Fatalf("populate conKeepersPublicKeyList error: %s", err.Error())
		}
		txn.Confirm(txn.ID, 1000, 10, client)

		// migrate zilToPolyTxHashMap
		zilToPolyTxHashMapResult, err := client.GetSmartContractSubState(strings.TrimPrefix(oldCcmAddr, "0x"), "zilToPolyTxHashMap", []interface{}{})
		if err != nil {
			log.Fatalf("get zilToPolyTxHashMap error: %s", err.Error())
		}
		type ZilToPolyTxHashMapResult struct {
			Result map[string]map[string]string `json:"result"`
		}
		var zilToPolyTxHashMap ZilToPolyTxHashMapResult
		err = json.Unmarshal([]byte(zilToPolyTxHashMapResult), &zilToPolyTxHashMap)
		if err != nil {
			log.Fatalf("unmarshal zilToPolyTxHashMap error: %s", err.Error())
		}
		log.Println("populate zilToPolyTxHashMap")
		zilToPolyTxHashIndexInt, err := strconv.ParseInt(zilToPolyTxHashIndex.Result["zilToPolyTxHashIndex"], 10, 64)
		if err != nil {
			log.Fatalf("parse zilToPolyTxHashIndexInt error: %s", err.Error())
		}
		latestPolyTxHashIndex := strconv.FormatInt(zilToPolyTxHashIndexInt-1, 10)
		txn, err = p.PopulateZilToPolyTxHashMap(zilToPolyTxHashIndex.Result["zilToPolyTxHashIndex"], zilToPolyTxHashMap.Result["zilToPolyTxHashMap"][latestPolyTxHashIndex])
		if err != nil {
			log.Fatalf("populate conKeepersPublicKeyList error: %s", err.Error())
		}
		txn.Confirm(txn.ID, 1000, 10, client)

		// migrate fromChainTxExist
		fromChainTxExistResult, err := client.GetSmartContractSubState(strings.TrimPrefix(oldCcmAddr, "0x"), "fromChainTxExist", []interface{}{})
		if err != nil {
			log.Fatalf("get fromChainTxExist error: %s", err.Error())
		}
		type FromChainTxExistResult struct {
			Result map[string]map[string]map[string]interface{} `json:"result"`
		}
		var fromChainTxExist FromChainTxExistResult
		err = json.Unmarshal([]byte(fromChainTxExistResult), &fromChainTxExist)
		if err != nil {
			log.Fatalf("unmarshal fromChainTxExist error: %s", err.Error())
		}
		fromChainTxExistInnerMap := fromChainTxExist.Result["fromChainTxExist"]

		balAndNonce, err := p.Client.GetBalance(p.Wallet.DefaultAccount.Address)
		if err != nil {
			log.Fatalf("get balance error: %s", err.Error())
		}

		startNonce := balAndNonce.Nonce
		log.Println("populate fromChainTxExist")
		for chainId, innerMap := range fromChainTxExistInnerMap {
			for hash, _ := range innerMap {
				nonce := strconv.FormatInt(startNonce+1, 10)
				txn, err := p.PopulateFromChainTxExistWithNonce(chainId, hash, nonce)
				if err != nil {
					log.Fatalf("PopulateFromChainTxExistWithNonce error: %s", err.Error())
				}
				log.Println(txn.ID)
				startNonce++
			}
		}
	},
}
