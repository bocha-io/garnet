package txbuilder

import (
	"context"
	"fmt"
	"math/big"

	"github.com/bocha-io/garnet/internal/logger"
	"github.com/ethereum/go-ethereum/accounts"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
	hdwallet "github.com/miguelmota/go-ethereum-hdwallet"
)

var (
	mnemonic     = "eternal envelope hat fame output noble roast screen bulk mind beyond sun brass dolphin wealth solid tone age diagram wall nothing often use delay"
	worldAddress = "0x5FbDB2315678afecb367f032d93F642f64180aa3"
	endpoint     = "http://localhost:8545"
)

func GetWallet(accountID int) (*hdwallet.Wallet, accounts.Account, error) {
	wallet, err := hdwallet.NewFromMnemonic(mnemonic)
	if err != nil {
		return nil, accounts.Account{}, err
	}

	path := hdwallet.MustParseDerivationPath(fmt.Sprintf("m/44'/60'/0'/0/%d", accountID))
	account, err := wallet.Derive(path, false)
	if err != nil {
		return nil, accounts.Account{}, err
	}
	return wallet, account, nil
}

func SendTransaction(accountID int, message string, args ...interface{}) (common.Hash, error) {
	// Generate the wallet
	wallet, account, err := GetWallet(accountID)
	if err != nil {
		return common.Hash{}, err
	}

	// Get coins
	_, err = Faucet(account.Address.Hex())
	if err != nil {
		return common.Hash{}, err
	}

	// Send transaction
	client, err := ethclient.Dial(endpoint)
	if err != nil {
		return common.Hash{}, err
	}

	nonce, err := client.PendingNonceAt(context.Background(), account.Address)
	if err != nil {
		return common.Hash{}, err
	}

	value := big.NewInt(0)
	gasLimit := uint64(20000000)
	gasPrice, err := client.SuggestGasPrice(context.Background())
	if err != nil {
		return common.Hash{}, err
	}

	var data []byte
	if len(args) > 0 {
		data, err = IWorldABI.Pack(message, args...)
		if err != nil {
			return common.Hash{}, err
		}
	} else {
		data, err = IWorldABI.Pack(message)
		if err != nil {
			return common.Hash{}, err
		}
	}

	toAddress := common.HexToAddress(worldAddress)
	tx := types.NewTransaction(nonce, toAddress, value, gasLimit, gasPrice, data)

	chainID, err := client.NetworkID(context.Background())
	if err != nil {
		return common.Hash{}, err
	}

	privateKey, err := wallet.PrivateKey(account)
	if err != nil {
		return common.Hash{}, err
	}

	signedTx, err := types.SignTx(tx, types.NewEIP155Signer(chainID), privateKey)
	if err != nil {
		return common.Hash{}, err
	}

	err = client.SendTransaction(context.Background(), signedTx)
	if err != nil {
		return common.Hash{}, err
	}

	logger.LogDebug(fmt.Sprintf("[backend] tx sent (%s) with hash: %s", message, signedTx.Hash().Hex()))

	return signedTx.Hash(), nil
}