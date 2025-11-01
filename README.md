# Jokenpo Game 🎮✊✋✌️

Projeto desenvolvido em **Go** que implementa um jogo de **Pedra, Papel e Tesoura** sobre uma **arquitetura de microsserviços distribuída**. A solução é projetada para ser escalável e tolerante a falhas, utilizando **Docker Compose** para orquestração e **HashiCorp Consul** para descoberta de serviço e eleição de líder.

---

## 📂 Estrutura do Projeto

*   `cmd/` → Contém o código-fonte dos executáveis:
    *   `client/` → O cliente de terminal para o jogador.
    *   `server/` → Os diferentes microsserviços:
        *   `session/` → API Gateway que gerencia conexões de clientes via WebSocket.
        *   `queue/` → Serviço de matchmaking (fila de pareamento).
        *   `shop/` → Serviço de loja para compra de cartas.
        *   `gameroom/` → Serviço que hospeda as partidas ativas.
        *   `loadbalancer/` → Proxy reverso dinâmico para os serviços de sessão.
*   `internal/` → Pacotes e lógica compartilhada entre os serviços:
    *   `network/` → Camada de comunicação WebSocket base.
    *   `game/` → Lógica e regras do jogo de cartas.
    *   `services/cluster/` → Implementação da integração com Consul (descoberta, eleição, etc.).
*   `docker-compose.yml` → Arquivo principal que orquestra todos os contêineres e a rede.
*   `go.mod` / `go.sum` → Dependências do projeto Go.
*   `LICENSE` → Licença do projeto.

---

## 🚀 Pré-requisitos

Antes de rodar o projeto, certifique-se de ter instalado:

*   [Go 1.22+](https://go.dev/dl/)
*   [Docker](https://docs.docker.com/get-docker/)
*   [Docker Compose](https://docs.docker.com/compose/)

---

## 🐳 Executando com Docker Compose (Recomendado)

A arquitetura foi projetada para ser executada como um conjunto de contêineres. Siga os passos abaixo:

### 1. Iniciar a Infraestrutura (Cluster Consul)

Este comando inicia os 3 nós do Consul que formam o "cérebro" do nosso sistema.

```bash
docker-compose --profile infra up -d
```

> **Aguarde cerca de 30 segundos** para que o cluster Consul se estabilize e eleja um líder antes de prosseguir.

### 2. Iniciar os Serviços da Aplicação

Este comando inicia todas as réplicas dos serviços do jogo, os load balancers e os conecta à rede do Consul.

```bash
docker-compose --profile game up -d --scale jokenpo-session=2 --scale jokenpo-queue=2 --scale jokenpo-shop=2 --scale jokenpo-gameroom=3
```

### 3. (Opcional) Observar o Cluster

Você pode ver todos os serviços registrados e seu status de saúde acessando a interface do Consul no seu navegador:
**[http://localhost:8500](http://localhost:8500)**

### 4. Executar o Cliente

O cliente é executado localmente na sua máquina e se conecta ao sistema através dos load balancers.

Em um novo terminal, na raiz do projeto, execute:

```bash
go run ./cmd/client/main.go
```

> O cliente tentará se conectar a `localhost:9080`, `localhost:9081` ou `localhost:9082`. Ele possui lógica de failover e se conectará a qualquer um dos load balancers que estiver disponível.

### 5. Desligar o Ambiente

Para parar e remover todos os contêineres, execute:

```bash
docker-compose down
```

---

## 🧪 Teste de Resiliência (Chaos Test)

Este cenário de teste valida a capacidade de auto-recuperação do sistema em caso de falha de um componente crítico. Vamos simular a falha do líder do serviço de loja:

1.  **Inicie o ambiente completo** conforme as instruções da seção anterior.

2.  **Identifique o líder atual do `jokenpo-shop`:**
    *   Acesse a UI do Consul: [http://localhost:8500](http://localhost:8500).
    *   Vá para a aba **Key/Value**.
    *   Clique na chave `service/jokenpo-shop/leader`. O valor exibido no campo "Value" é o hostname do contêiner líder (ex: `b91b6cac2597`).

3.  **Encontre o ID do contêiner líder:**
    No seu terminal, liste os contêineres em execução e encontre aquele com o hostname que você anotou.
    ```bash
    docker ps
    ```

4.  **Injete a falha (derrube o líder):**
    Use o ID do contêiner para pará-lo.
    ```bash
    docker stop <ID_DO_CONTAINER_LIDER>
    ```

5.  **Observe a recuperação:**
    *   **Nos logs:** Observe os logs da outra réplica do `jokenpo-shop` (`docker logs -f <ID_DA_REPLICA>`). Você verá mensagens indicando que ela se tornou o novo líder.
    *   **Na UI do Consul:** Atualize a página Key/Value. O valor da chave `service/jokenpo-shop/leader` terá mudado para o hostname do novo contêiner líder.

O serviço de loja continuará funcionando, agora servido pela réplica que foi promovida automaticamente.

---

## 📖 Como Jogar

1.  O cliente conecta ao sistema via **WebSocket** através de um dos **Load Balancers**.
2.  O jogador escolhe uma das opções que o menu exibe, enviando comandos ao servidor.
3.  O **serviço de sessão** recebe os comandos e os orquestra com os serviços de backend (fila, loja, sala de jogo) para processar a lógica e retornar os resultados.

---

## ⚖️ Licença

Este projeto é distribuído sob a licença MIT. Consulte o arquivo [LICENSE](LICENSE) para mais detalhes.