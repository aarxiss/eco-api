package blockchain

import (
	"context"
	"crypto/ecdsa"
	"crypto/sha256"
	"fmt"
	"math/big"
	"strings"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
)

const contractABI = `[{"anonymous":false,"inputs":[{"indexed":true,"internalType":"bytes32","name":"dataHash","type":"bytes32"},{"indexed":true,"internalType":"address","name":"sender","type":"address"},{"indexed":false,"internalType":"uint256","name":"blockNumber","type":"uint256"},{"indexed":false,"internalType":"uint256","name":"timestamp","type":"uint256"}],"name":"HashAnchored","type":"event"},{"inputs":[{"internalType":"bytes32","name":"dataHash","type":"bytes32"}],"name":"anchor","outputs":[],"stateMutability":"nonpayable","type":"function"},{"inputs":[{"internalType":"bytes32","name":"","type":"bytes32"}],"name":"anchoredAtBlock","outputs":[{"internalType":"uint256","name":"","type":"uint256"}],"stateMutability":"view","type":"function"},{"inputs":[{"internalType":"bytes32","name":"dataHash","type":"bytes32"}],"name":"isAnchored","outputs":[{"internalType":"bool","name":"","type":"bool"}],"stateMutability":"view","type":"function"}]`

type AnchorClient struct {
	client          *ethclient.Client
	privateKey      *ecdsa.PrivateKey
	contractAddress common.Address
	chainID         *big.Int
	parsedABI       abi.ABI
}

func Init(rpcURL, privKeyHex, contractAddrHex string) (*AnchorClient, error) {
	client, err := ethclient.Dial(rpcURL)
	if err != nil {
		return nil, fmt.Errorf("помилка підключення до RPC: %v", err)
	}

	privateKey, err := crypto.HexToECDSA(privKeyHex)
	if err != nil {
		return nil, fmt.Errorf("неправильний приватний ключ: %v", err)
	}

	chainID, err := client.ChainID(context.Background())
	if err != nil {
		return nil, err
	}

	parsedABI, err := abi.JSON(strings.NewReader(contractABI))
	if err != nil {
		return nil, err
	}

	return &AnchorClient{
		client:          client,
		privateKey:      privateKey,
		contractAddress: common.HexToAddress(contractAddrHex),
		chainID:         chainID,
		parsedABI:       parsedABI,
	}, nil
}

func (ac *AnchorClient) AnchorData(ctx context.Context, sensorID string, value float64) (string, error) {
	// 1. Створюємо унікальний рядок з показників
	dataString := fmt.Sprintf("%s:%.2f", sensorID, value)

	hash := sha256.Sum256([]byte(dataString))

	data, err := ac.parsedABI.Pack("anchor", hash)
	if err != nil {
		return "", fmt.Errorf("помилка пакування даних: %v", err)
	}

	publicKey := ac.privateKey.Public().(*ecdsa.PublicKey)
	fromAddress := crypto.PubkeyToAddress(*publicKey)

	nonce, err := ac.client.PendingNonceAt(ctx, fromAddress)
	if err != nil {
		return "", err
	}

	gasPrice, err := ac.client.SuggestGasPrice(ctx)
	if err != nil {
		return "", err
	}

	tx := types.NewTransaction(nonce, ac.contractAddress, big.NewInt(0), uint64(150000), gasPrice, data)
	signedTx, err := types.SignTx(tx, types.NewEIP155Signer(ac.chainID), ac.privateKey)
	if err != nil {
		return "", err
	}

	err = ac.client.SendTransaction(ctx, signedTx)
	if err != nil {
		return "", err
	}

	return signedTx.Hash().Hex(), nil
}

// IsAnchored перевіряє, чи захешовані ці дані в блокчейні
func (ac *AnchorClient) IsAnchored(ctx context.Context, dataHash [32]byte) (bool, error) {
	// Викликаємо функцію "isAnchored" нашого смарт-контракту
	var out []interface{}
	err := ac.parsedABI.UnpackIntoInterface(&out, "isAnchored", []byte{}) // Отримуємо результат

	// Але простіше викликати метод через CallOpts
	// (Для простоти ми використовуємо Pack, щоб створити запит)
	data, err := ac.parsedABI.Pack("isAnchored", dataHash)
	if err != nil {
		return false, err
	}

	msg := ethereum.CallMsg{
		To:   &ac.contractAddress,
		Data: data,
	}

	{
		result, err := ac.client.CallContract(ctx, msg, nil)
		if err != nil {
			return false, err
		}

		var isAnchored bool
		err = ac.parsedABI.UnpackIntoInterface(&isAnchored, "isAnchored", result)
		return isAnchored, err
	}

}
