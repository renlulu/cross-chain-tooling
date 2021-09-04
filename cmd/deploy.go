package cmd

import (
	"fmt"
	"github.com/Zilliqa/cross-chain-tooling/util"
	"github.com/Zilliqa/gozilliqa-sdk/account"
	"github.com/Zilliqa/gozilliqa-sdk/crosschain/polynetwork"
	"github.com/Zilliqa/gozilliqa-sdk/provider"
	"github.com/spf13/cobra"
	"log"
)

var contractLocation string
var privateKey string
var host string
var chainId int

func init() {
	deployCmd.Flags().StringVarP(&contractLocation, "contract", "c", "./contracts", "smart contract folder")
	deployCmd.Flags().StringVarP(&privateKey, "key", "k", "", "private key used to deploy smart contract, also becomes to the admin of the contracts")
	deployCmd.Flags().StringVarP(&host, "api", "a", "https://dev-api.zilliqa.com", "zilliqa api endpoint")
	deployCmd.Flags().IntVarP(&chainId, "chainId", "i", 333, "zilliqa chain id")
	RootCmd.AddCommand(deployCmd)
}

var deployCmd = &cobra.Command{
	Use:   "deploy",
	Short: "deploy cross chain manager contract and its proxy",
	Long:  "deploy cross chain manager contract and its proxy",
	Run: func(cmd *cobra.Command, args []string) {
		if privateKey == "" {
			log.Fatalln("private key cannot be empty")
		}

		deployer := &util.Deployer{
			PrivateKey:    privateKey,
			Host:          host,
			ProxyPath:     fmt.Sprintf("%s/ZilCrossChainManagerProxy.scilla", contractLocation),
			ImplPath:      fmt.Sprintf("%s/ZilCrossChainManager.scilla", contractLocation),
			LockProxyPath: fmt.Sprintf("%s/LockProxySwitcheo.scilla", contractLocation),
		}

		wallet := account.NewWallet()
		wallet.AddByPrivateKey(deployer.PrivateKey)
		client := provider.NewProvider(deployer.Host)
		proxy, impl, lockProxy, err := deployer.Deploy(chainId, wallet, client)
		if err != nil {
			log.Fatalln(err.Error())
		}
		log.Printf("cross chain manager proxy address: %s\n", proxy)
		log.Printf("cross chain manager address: %s\n", impl)
		log.Printf("lock proxy address: %s\n", lockProxy)

		p := &polynetwork.Proxy{
			ProxyAddr:  proxy,
			ImplAddr:   impl,
			Wallet:     wallet,
			Client:     client,
			ChainId:    chainId,
			MsgVersion: util.MsgVersion,
		}

		log.Printf("upgrade ccm proxy to its impl")
		_, err1 := p.UpgradeTo()
		if err1 != nil {
			log.Fatalln(err1.Error())
		}

	},
}
