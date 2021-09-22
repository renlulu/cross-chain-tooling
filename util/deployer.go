package util

import (
	"errors"
	"github.com/Zilliqa/gozilliqa-sdk/account"
	contract2 "github.com/Zilliqa/gozilliqa-sdk/contract"
	"github.com/Zilliqa/gozilliqa-sdk/core"
	"github.com/Zilliqa/gozilliqa-sdk/keytools"
	"github.com/Zilliqa/gozilliqa-sdk/provider"
	"github.com/Zilliqa/gozilliqa-sdk/util"
	"io/ioutil"
	"log"
	"strconv"
)

const MsgVersion = 1

type Deployer struct {
	PrivateKey    string
	Host          string
	ProxyPath     string
	ImplPath      string
	LockProxyPath string
}

func (d *Deployer) deploy(chainID int, contractCode []byte, init []core.ContractValue, wallet *account.Wallet, client *provider.Provider, senderPubKey []byte, sendAddress string) (string, error) {

	gasPrice, err := client.GetMinimumGasPrice()
	if err != nil {
		return "", err
	}
	contract := contract2.Contract{
		Code:     string(contractCode),
		Init:     init,
		Signer:   wallet,
		Provider: client,
	}
	balAndNonce, _ := client.GetBalance(sendAddress)
	deployParams := contract2.DeployParams{
		Version:      strconv.FormatInt(int64(util.Pack(chainID, MsgVersion)), 10),
		Nonce:        strconv.FormatInt(balAndNonce.Nonce+1, 10),
		GasPrice:     gasPrice,
		GasLimit:     "30000",
		SenderPubKey: util.EncodeHex(senderPubKey),
	}

	tx, err1 := contract.Deploy(deployParams)
	if err1 != nil {
		return "", err1
	}

	tx.Confirm(tx.ID, 1000, 10, client)
	if tx.Status == core.Confirmed {
		return tx.ContractAddress, nil
	} else {
		return "", errors.New("deploy failed")
	}
}

func (d *Deployer) Deploy(chainId int, wallet *account.Wallet, client *provider.Provider) (string, string, string, error) {
	pubKey := keytools.GetPublicKeyFromPrivateKey(util.DecodeHex(d.PrivateKey), true)
	address := keytools.GetAddressFromPublic(pubKey)

	// deploy proxy
	log.Println("deploy cross chain manager proxy contract")
	code, _ := ioutil.ReadFile(d.ProxyPath)
	init := []core.ContractValue{
		{
			"_scilla_version",
			"Uint32",
			"0",
		},
		{
			"init_crosschain_manager",
			"ByStr20",
			"0x0000000000000000000000000000000000000000",
		},
		{
			"init_admin",
			"ByStr20",
			"0x" + address,
		},
	}

	proxy, err1 := d.deploy(chainId, code, init, wallet, client, pubKey, address)
	if err1 != nil {
		return "", "", "", err1
	}

	// deploy cross chain manager
	log.Println("deploy cross chain manager contract")
	code, _ = ioutil.ReadFile(d.ImplPath)
	init = []core.ContractValue{
		{
			"_scilla_version",
			"Uint32",
			"0",
		},
		{
			"this_chain_id",
			"Uint64",
			"111",
		},
		{
			"init_proxy_address",
			"ByStr20",
			"0x" + proxy,
		},
		{
			"init_admin",
			"ByStr20",
			"0x" + address,
		},
	}

	impl, err := d.deploy(chainId, code, init, wallet, client, pubKey, address)
	if err != nil {
		return "", "", "", err
	}

	// deploy lock proxy
	log.Println("deploy lock proxy contract")
	code, _ = ioutil.ReadFile(d.LockProxyPath)
	init = []core.ContractValue{
		{
			"_scilla_version",
			"Uint32",
			"0",
		},
		{
			"init_admin",
			"ByStr20",
			"0x" + address,
		},
		{
			"init_manager_proxy",
			"ByStr20",
			"0x" + proxy,
		},
		{
			"init_manager",
			"ByStr20",
			"0x" + impl,
		},
		{
			"init_counterpart_chainId",
			"Uint64",
			"1",
		},
	}

	lockProxy, err2 := d.deploy(chainId, code, init, wallet, client, pubKey, address)
	if err2 != nil {
		return "", "", "", err2
	}

	return proxy, impl, lockProxy, nil
}
