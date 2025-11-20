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

	"jokenpo/internal/ledger" // O pacote gerado pelo abigen

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
)

// Configurações do Geth --dev
const (
	// Endereço do container docker (visto de dentro da rede docker)
	BlockchainURL = "http://jokenpo-blockchain:8545"
	// Chave privada padrão do modo --dev
	DevPrivateKey = "b71c71a67e1177ad4e901695e1b4b9ee17ae16c6668d313eac2f96dbcda3f291"
)

type BlockchainClient struct {
	client   *ethclient.Client
	contract *ledger.Ledger
	auth     *bind.TransactOpts
	address  common.Address
}

// LogEntry é uma struct auxiliar para ordenar logs misturados por data
type LogEntry struct {
	Timestamp uint64
	Message   string
}

func NewBlockchainClient() (*BlockchainClient, error) {
	// 1. Conecta no Geth
	client, err := ethclient.Dial(BlockchainURL)
	if err != nil {
		return nil, fmt.Errorf("falha ao conectar no Geth: %v", err)
	}

	// 2. Configura Autenticação (Admin)
	privateKey, err := crypto.HexToECDSA(DevPrivateKey)
	if err != nil {
		return nil, fmt.Errorf("erro na chave privada: %v", err)
	}
	chainID, _ := client.ChainID(context.Background())
	auth, _ := bind.NewKeyedTransactorWithChainID(privateKey, chainID)

	// 3. Deploy ou Load do Contrato
	// Para simplificar no PBL: Toda vez que o serviço sobe, ele tenta fazer deploy.
	// O endereço retornado será usado para interações.
	addr, _, instance, err := ledger.DeployLedger(auth, client)
	if err != nil {
		return nil, fmt.Errorf("falha ao fazer deploy/load do contrato: %v", err)
	}
	
	log.Printf("[Blockchain] Contrato ativo em: %s", addr.Hex())

	return &BlockchainClient{
		client:   client,
		contract: instance,
		auth:     auth,
		address:  addr,
	}, nil
}

// GetAuditReport busca TODOS os eventos e formata como texto cronológico
func (bc *BlockchainClient) GetAuditReport() (string, error) {
	// Filtro para pegar desde o bloco 0
	opts := &bind.FilterOpts{Start: 0}

	var allLogs []LogEntry

	// 1. Buscar Logs de Pacotes
	iterPacks, err := bc.contract.FilterAuditPackOpened(opts)
	if err == nil {
		for iterPacks.Next() {
			ev := iterPacks.Event
			msg := fmt.Sprintf("Player %s abriu um pacote e recebeu %d cartas: %v", 
				shortID(ev.PlayerId), len(ev.CardIds), ev.CardIds)
			allLogs = append(allLogs, LogEntry{Timestamp: ev.Timestamp.Uint64(), Message: msg})
		}
	}

	// 2. Buscar Logs de Trocas
	iterTrades, err := bc.contract.FilterAuditTrade(opts)
	if err == nil {
		for iterTrades.Next() {
			ev := iterTrades.Event
			msg := fmt.Sprintf("Player %s TROCOU a carta %s com Player %s", 
				shortID(ev.FromPlayer), ev.CardId, shortID(ev.ToPlayer))
			allLogs = append(allLogs, LogEntry{Timestamp: ev.Timestamp.Uint64(), Message: msg})
		}
	}

	// 3. Buscar Logs de Partidas
	iterMatches, err := bc.contract.FilterAuditMatch(opts)
	if err == nil {
		for iterMatches.Next() {
			ev := iterMatches.Event
			msg := fmt.Sprintf("PARTIDA %s: Vencedor %s vs Perdedor %s", 
				shortID(ev.RoomId), shortID(ev.WinnerId), shortID(ev.LoserId))
			allLogs = append(allLogs, LogEntry{Timestamp: ev.Timestamp.Uint64(), Message: msg})
		}
	}

	// 4. Ordenar por Data (Cronológico)
	sort.Slice(allLogs, func(i, j int) bool {
		return allLogs[i].Timestamp < allLogs[j].Timestamp
	})

	// 5. Formatar String Final
	var sb strings.Builder
	sb.WriteString("=== RELATÓRIO DE AUDITORIA BLOCKCHAIN (IMUTÁVEL) ===\n\n")
	if len(allLogs) == 0 {
		sb.WriteString("(Nenhum registro encontrado na blockchain ainda)\n")
	}
	for _, l := range allLogs {
		t := time.Unix(int64(l.Timestamp), 0)
		sb.WriteString(fmt.Sprintf("[%s] %s\n", t.Format("15:04:05"), l.Message))
	}
	sb.WriteString("\n====================================================")

	return sb.String(), nil
}

// Funções de Escrita (Helpers para seus outros serviços usarem)

func (bc *BlockchainClient) LogPack(playerId string, uniqueCardIds []string) error {
	// Atualiza o Nonce para evitar erro de transação substituta
	nonce, err := bc.client.PendingNonceAt(context.Background(), bc.auth.From)
	if err != nil {
		return fmt.Errorf("erro ao obter nonce: %v", err)
	}
	bc.auth.Nonce = big.NewInt(int64(nonce))
	
	// Chama o contrato inteligente
	tx, err := bc.contract.LogPackOpening(bc.auth, playerId, uniqueCardIds)
	if err != nil {
		return err
	}

	log.Printf("[Blockchain] Transação LogPack enviada! Hash: %s", tx.Hash().Hex())
	return nil
}

func (bc *BlockchainClient) LogMatch(roomId, winnerId, loserId string) error {
	nonce, _ := bc.client.PendingNonceAt(context.Background(), bc.auth.From)
	bc.auth.Nonce = big.NewInt(int64(nonce))

	_, err := bc.contract.LogMatchResult(bc.auth, roomId, winnerId, loserId)
	return err
}

// Helper visual
func shortID(id string) string {
	if len(id) > 8 {
		return id[:8] + "..."
	}
	return id
}

//END OF FILE jokenpo/internal/services/blockchain/client.go