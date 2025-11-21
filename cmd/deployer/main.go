//START OF FILE jokenpo/cmd/deployer/main.go
package main

import (
	"context"
	"log"
	"time"

	"jokenpo/internal/ledger" // Seu pacote gerado
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/hashicorp/consul/api"
)

const (
	BlockchainURL = "http://jokenpo-blockchain:8545"
	// Chave privada padrão do Geth --dev
	DevPrivateKey = "b71c71a67e1177ad4e901695e1b4b9ee17ae16c6668d313eac2f96dbcda3f291"
	ConsulKey     = "jokenpo/config/contract_address"
)

func main() {
	log.Println("[Deployer] Iniciando Job de Deploy do Contrato...")

	// 1. Conectar ao Geth (Com Retry)
	var client *ethclient.Client
	var err error
	for i := 0; i < 30; i++ {
		client, err = ethclient.Dial(BlockchainURL)
		if err == nil {
			_, err = client.ChainID(context.Background())
			if err == nil { break }
		}
		log.Printf("[Deployer] Aguardando Geth... (%v)", err)
		time.Sleep(1 * time.Second)
	}
	if err != nil { log.Fatalf("Fatal: Geth inalcançável: %v", err) }

	// 2. Preparar Transação
	privateKey, _ := crypto.HexToECDSA(DevPrivateKey)
	chainID, _ := client.ChainID(context.Background())
	auth, _ := bind.NewKeyedTransactorWithChainID(privateKey, chainID)
	auth.GasLimit = 3000000

	// 3. Fazer Deploy
	log.Println("[Deployer] Enviando transação de criação de contrato...")
	addr, tx, _, err := ledger.DeployLedger(auth, client)
	if err != nil {
		log.Fatalf("Fatal: Falha no deploy: %v", err)
	}

	log.Printf("[Deployer] Deploy enviado! Tx: %s", tx.Hash().Hex())
	log.Printf("[Deployer] Endereço do Contrato: %s", addr.Hex())

	// 4. Esperar Mineração (Garantia)
	_, err = bind.WaitMined(context.Background(), client, tx)
	if err != nil {
		log.Fatalf("Fatal: Erro na mineração do contrato: %v", err)
	}
	log.Println("[Deployer] Contrato minerado e confirmado!")

	// 5. Salvar no Consul
	log.Println("[Deployer] Conectando ao Consul...")
	consulConfig := api.DefaultConfig()
	consulConfig.Address = "consul-1:8500" // Endereço interno do docker
	consulClient, err := api.NewClient(consulConfig)
	if err != nil {
		log.Fatalf("Fatal: Erro cliente Consul: %v", err)
	}

	// Retry no Consul (caso ele esteja elegendo líder)
	for i := 0; i < 30; i++ {
		kv := &api.KVPair{Key: ConsulKey, Value: []byte(addr.Hex())}
		_, err = consulClient.KV().Put(kv, nil)
		if err == nil {
			log.Printf("✅ [Deployer] SUCESSO! Endereço salvo no Consul: %s", ConsulKey)
			return // SUCESSO TOTAL, O CONTAINER VAI PARAR AQUI
		}
		log.Printf("[Deployer] Aguardando Consul... (%v)", err)
		time.Sleep(1 * time.Second)
	}
	log.Fatal("Fatal: Timeout tentando salvar no Consul")
}
//END OF FILE jokenpo/cmd/deployer/main.go