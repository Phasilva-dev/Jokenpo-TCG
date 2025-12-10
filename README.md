# Jokenpo Game Multiplayer (Hybrid Web2/Web3) 🎮✊✋✌️⛓️

Projeto desenvolvido em **Go** que implementa um jogo de **Pedra, Papel e Tesoura** sobre uma **arquitetura de microsserviços distribuída e híbrida**.

A solução combina a performance de servidores de jogo em tempo real (Web2) com a transparência e imutabilidade de uma **Blockchain Ethereum** (Web3). Utilizamos **Docker Compose** para orquestração, **HashiCorp Consul** para descoberta de serviço/eleição de líder e **Geth (Go-Ethereum)** para o livro razão distribuído.

---

## 📂 Estrutura do Projeto

*   `cmd/` → Código-fonte dos executáveis:
    *   `client/` → Cliente de terminal (CLI) para o jogador.
    *   `deployer/` → **(Novo)** Serviço utilitário que publica o Smart Contract e configura o endereço no Consul.
    *   `server/` → Microsserviços:
        *   `session/` → API Gateway e BFF (Backend for Frontend) via WebSocket.
        *   `queue/` → Matchmaker e gerenciador de trocas (Atomic Swaps).
        *   `shop/` → Loja de pacotes (Minting de ativos).
        *   `gameroom/` → Lógica da partida e regras do jogo.
        *   `loadbalancer/` → Proxy reverso dinâmico em Go.
*   `contract/` → **(Novo)** Código fonte do Smart Contract (`JokenpoLedger.sol`).
*   `internal/` → Pacotes compartilhados:
    *   `services/blockchain/` → **(Novo)** Cliente Go para interação com Ethereum.
    *   `ledger/` → Bindings Go gerados a partir do contrato Solidity.
    *   `network/`, `game/`, `cluster/` → Core do sistema.
*   `docker-compose.yml` → Orquestração completa do ambiente.

---

## 🚀 Pré-requisitos

*   [Go 1.22+](https://go.dev/dl/)
*   [Docker](https://docs.docker.com/get-docker/) e [Docker Compose](https://docs.docker.com/compose/)

---

## 🐳 Executando o Projeto

A inicialização deve seguir uma ordem estrita para garantir que a infraestrutura e a blockchain estejam prontas antes dos serviços de jogo.

### 1. Iniciar a Infraestrutura (Cluster Consul)
Sobe os 3 nós do Consul para formar o quórum de descoberta e eleição.

```bash
docker-compose --profile infra up --build -d
```
> ⏳ **Aguarde ~10 segundos** para o cluster eleger um líder.

### 2. Iniciar a Blockchain (Geth)
Sobe o nó Ethereum privado em modo de desenvolvimento (mineração instantânea).

```bash
docker-compose --profile bc up --build -d
```
> ⏳ **Aguarde ~5 segundos** para o nó Geth estar pronto para aceitar conexões RPC.

### 3. Iniciar o Jogo e Deployer
Sobe os microsserviços do jogo e o `deployer`. O `deployer` publicará o contrato na blockchain e avisará os outros serviços automaticamente via Consul.

```bash
docker-compose --profile game up --build -d --scale jokenpo-session=2 --scale jokenpo-queue=2 --scale jokenpo-shop=2 --scale jokenpo-gameroom=3
```

---

## 🕹️ Como Jogar

O cliente roda localmente na sua máquina (fora do Docker) e conecta nos Load Balancers.

1.  Abra um terminal na raiz do projeto.
2.  Execute o cliente:
    ```bash
    go run ./cmd/client/main.go
    ```
3.  O cliente tentará conectar em `localhost:9080`, `9081` ou `9082` (Load Balancers).
4.  **No Menu:**
    *   Use as opções **1-8** para jogar, comprar pacotes e trocar cartas.
    *   Use a opção **10. [BLOCKCHAIN] Ver Livro Razão** para auditar suas transações diretamente da rede Ethereum.

---

## 🔍 Monitoramento e Logs (Debug)

Para verificar se a integração Web3 está funcionando, você pode acompanhar os logs específicos:

### Ver a Blockchain trabalhando (Mineração)
Veja os blocos sendo criados e transações sendo aceitas.
```bash
docker logs -f jokenpo-blockchain
```

### Ver o Deploy do Contrato
Confira se o contrato foi publicado e o endereço salvo no Consul.
```bash
docker logs jokenpo-deployer
```

### Ver Transações de Compra (Shop Leader)
Acompanhe o líder da loja "mintando" novas cartas.
```bash
docker logs -f jokenpo-shop-1
# ou jokenpo-shop-2 (dependendo de quem for o líder)
```

---

## 🧪 Arquitetura e Resiliência

### Híbrido Web2 + Web3
*   **Performance:** O jogo roda em memória (Web2) para garantir UX fluida (sem lag de blockchain).
*   **Confiança:** Operações críticas (Compra, Troca, Resultado de Partida) são persistidas assincronamente na Blockchain.
*   **Consistência:** Utilizamos o padrão de **Eleição de Líder** (via Consul) para garantir que apenas uma instância do serviço escreva na Blockchain por vez, evitando conflitos de transação (Nonce) e gasto duplo.

### Teste de Falha (Chaos Test)
Você pode derrubar o líder da loja ou da fila enquanto o sistema roda.
1.  Descubra quem é o líder no Consul ([http://localhost:8500](http://localhost:8500) -> Key/Value -> `service/jokenpo-shop/leader`).
2.  Pare o container: `docker stop <ContainerID>`.
3.  O Consul detectará a falha, elegerá um novo líder, e o novo líder retomará a conexão com a Blockchain automaticamente.

---

## 🧹 Limpeza

Para parar e remover todos os contêineres, redes e volumes:

```bash
docker-compose down
```

---

## ⚖️ Licença

Este projeto é distribuído sob a licença MIT.