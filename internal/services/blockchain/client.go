//START OF FILE jokenpo/internal/services/blockchain/client.go
package blockchain

import (
	"context"
	"fmt"
	"log"
	"math/big"
	"sort"
	"strings"
	"time"

	"jokenpo/internal/ledger"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
)

const (
	BlockchainURL = "http://jokenpo-blockchain:8545"
	DevPrivateKey = "b71c71a67e1177ad4e901695e1b4b9ee17ae16c6668d313eac2f96dbcda3f291"
)

type BlockchainClient struct {
	client   *ethclient.Client
	contract *ledger.Ledger
	auth     *bind.TransactOpts
	address  common.Address
}

type LogEntry struct {
	Timestamp uint64
	Message   string
}

func InitBlockchain(existingAddr string) (*BlockchainClient, string, error) {
	var client *ethclient.Client
	var err error

	log.Println("[Blockchain] Tentando conectar ao nó Geth...")
	for i := 0; i < 10; i++ {
		client, err = ethclient.Dial(BlockchainURL)
		if err == nil {
			if _, err := client.ChainID(context.Background()); err == nil {
				break
			}
		}
		time.Sleep(2 * time.Second)
	}
	if err != nil {
		return nil, "", fmt.Errorf("timeout connecting to Geth: %v", err)
	}

	privateKey, _ := crypto.HexToECDSA(DevPrivateKey)
	chainID, _ := client.ChainID(context.Background())
	auth, _ := bind.NewKeyedTransactorWithChainID(privateKey, chainID)
	auth.GasLimit = 3000000 // Limite alto para evitar erros de estimativa em dev

	var contract *ledger.Ledger
	var addr common.Address
	var finalAddrStr string

	if existingAddr == "" {
		// MODO DEPLOY
		for i := 0; i < 5; i++ {
			addr, _, contract, err = ledger.DeployLedger(auth, client)
			if err == nil { break }
			time.Sleep(1 * time.Second)
		}
		if err != nil { return nil, "", fmt.Errorf("deploy failed: %v", err) }
		finalAddrStr = addr.Hex()
		log.Printf(">>> [Blockchain] CONTRATO CRIADO EM: %s <<<", finalAddrStr)
	} else {
		// MODO CONNECT
		log.Printf(">>> [Blockchain] CONECTANDO EM: %s <<<", existingAddr)
		addr = common.HexToAddress(existingAddr)
		contract, err = ledger.NewLedger(addr, client)
		if err != nil { return nil, "", fmt.Errorf("bind failed: %v", err) }
		finalAddrStr = existingAddr
	}

	return &BlockchainClient{
		client:   client,
		contract: contract,
		auth:     auth,
		address:  addr,
	}, finalAddrStr, nil
}

func (bc *BlockchainClient) GetAuditReport() (string, error) {
	opts := &bind.FilterOpts{Start: 0, Context: context.Background()}
	var allLogs []LogEntry

	iterPacks, err := bc.contract.FilterAuditPackOpened(opts)
	if err == nil {
		for iterPacks.Next() {
			ev := iterPacks.Event
			msg := fmt.Sprintf("PACK: Player %s... recebeu %d cartas", shortID(ev.PlayerId), len(ev.CardIds))
			allLogs = append(allLogs, LogEntry{Timestamp: ev.Timestamp.Uint64(), Message: msg})
		}
	}

	iterTrades, err := bc.contract.FilterAuditTrade(opts)
	if err == nil {
		for iterTrades.Next() {
			ev := iterTrades.Event
			msg := fmt.Sprintf("TRADE: %s -> %s (%s)", shortID(ev.FromPlayer), shortID(ev.ToPlayer), ev.CardId)
			allLogs = append(allLogs, LogEntry{Timestamp: ev.Timestamp.Uint64(), Message: msg})
		}
	}

	iterMatches, err := bc.contract.FilterAuditMatch(opts)
	if err == nil {
		for iterMatches.Next() {
			ev := iterMatches.Event
			msg := fmt.Sprintf("MATCH: Sala %s | Vencedor: %s", shortID(ev.RoomId), shortID(ev.WinnerId))
			allLogs = append(allLogs, LogEntry{Timestamp: ev.Timestamp.Uint64(), Message: msg})
		}
	}

	sort.Slice(allLogs, func(i, j int) bool {
		return allLogs[i].Timestamp < allLogs[j].Timestamp
	})

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("=== AUDITORIA BLOCKCHAIN (Contrato: %s) ===\n", shortID(bc.address.Hex())))
	if len(allLogs) == 0 {
		sb.WriteString("(Nenhum registro encontrado ainda)\n")
	}
	for _, l := range allLogs {
		t := time.Unix(int64(l.Timestamp), 0)
		sb.WriteString(fmt.Sprintf("[%s] %s\n", t.Format("15:04:05"), l.Message))
	}
	sb.WriteString("=======================================")
	return sb.String(), nil
}

// Helper para enviar e aguardar mineração
func (bc *BlockchainClient) sendAndWait(txFunc func() error) error {
	return nil 
}

func (bc *BlockchainClient) LogPack(playerId string, uniqueCardIds []string) error {
	nonce, _ := bc.client.PendingNonceAt(context.Background(), bc.auth.From)
	bc.auth.Nonce = big.NewInt(int64(nonce))
	
	tx, err := bc.contract.LogPackOpening(bc.auth, playerId, uniqueCardIds)
	if err != nil { return err }

    // Aguarda mineração
    receipt, err := bind.WaitMined(context.Background(), bc.client, tx)
    if err != nil { return err }
    if receipt.Status == 0 { return fmt.Errorf("transação falhou (REVERT)") }
    
	log.Printf("[Blockchain] LogPack Confirmado! Bloco: %d", receipt.BlockNumber)
	return nil
}

func (bc *BlockchainClient) LogTrade(from, to, cardId string) error {
	nonce, _ := bc.client.PendingNonceAt(context.Background(), bc.auth.From)
	bc.auth.Nonce = big.NewInt(int64(nonce))

	tx, err := bc.contract.LogTrade(bc.auth, from, to, cardId)
	if err != nil { return err }

    receipt, err := bind.WaitMined(context.Background(), bc.client, tx)
    if err != nil { return err }
    if receipt.Status == 0 { return fmt.Errorf("transação falhou (REVERT)") }
	return nil
}

func (bc *BlockchainClient) LogMatch(roomId, winnerId, loserId string) error {
	nonce, _ := bc.client.PendingNonceAt(context.Background(), bc.auth.From)
	bc.auth.Nonce = big.NewInt(int64(nonce))

	tx, err := bc.contract.LogMatchResult(bc.auth, roomId, winnerId, loserId)
	if err != nil { return err }
    
    receipt, err := bind.WaitMined(context.Background(), bc.client, tx)
    if err != nil { return err }
    if receipt.Status == 0 { return fmt.Errorf("transação falhou (REVERT)") }
	return nil
}

func shortID(id string) string {
	if len(id) > 8 { return id[:8] + "..." }
	return id
}

// FindTokenForCard consulta a blockchain para encontrar um Token UUID que corresponda
// à carta genérica (cardKey) que o jogador possui.
// Ex: Entrada: "rock:1:red" -> Saída: "rock:1:red#uuid-1234..."
func (bc *BlockchainClient) FindTokenForCard(playerID, cardKey string) (string, error) {
    // Chamada de leitura (Call), não gasta gás e é rápida.
    opts := &bind.CallOpts{Context: context.Background()}
    
    // Pega todos os ativos do jogador na blockchain
    assets, err := bc.contract.GetPlayerAssets(opts, playerID)
    if err != nil {
        return "", fmt.Errorf("erro ao ler ativos da blockchain: %v", err)
    }

    // Procura o primeiro ativo que corresponda ao tipo da carta
    // O formato na blockchain é "cardKey#UUID"
    prefix := cardKey + "#"
    for _, token := range assets {
        if strings.HasPrefix(token, prefix) {
            return token, nil
        }
    }

    return "", fmt.Errorf("token não encontrado na blockchain para a carta %s do jogador %s", cardKey, playerID)
}
//END OF FILE jokenpo/internal/services/blockchain/client.go