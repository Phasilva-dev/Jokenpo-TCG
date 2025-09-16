#!/bin/bash

# --- CONFIGURAÇÃO DO TESTE ---
CONCURRENT_USERS=200
TEST_DURATION=180
COMPOSE_FILE="docker-compose.simple-test.yml"
# Define o nome do arquivo de log em uma variável para fácil reutilização
LOG_FILE="simple_test_results.log"

echo "--- Jokenpo TCG Stress Test ---"
echo "Limpando ambiente anterior (se houver)..."
docker-compose -f $COMPOSE_FILE down --volumes

echo ""
echo "Construindo as imagens Docker..."
docker-compose -f $COMPOSE_FILE build
if [ $? -ne 0 ]; then
    echo "Falha no build do Docker. Abortando."
    exit 1
fi

echo ""
echo "Iniciando o teste de estresse com $CONCURRENT_USERS usuários por $TEST_DURATION segundos..."

# Inicia o docker-compose EM SEGUNDO PLANO (&) e redireciona os logs para o arquivo
docker-compose -f $COMPOSE_FILE up --scale bot=$CONCURRENT_USERS > $LOG_FILE 2>&1 &

# Salva o Process ID (PID) do docker-compose
COMPOSE_PID=$!

# Barra de progresso simples
echo "Teste em andamento:"
for i in $(seq $TEST_DURATION); do
    printf "."
    sleep 1
done
echo " Tempo esgotado."

echo ""
echo "Parando o teste e coletando resultados..."
# Força o desligamento de todo o ambiente de teste para parar os bots
docker-compose -f $COMPOSE_FILE down --volumes >> $LOG_FILE 2>&1

# --- Coleta de Métricas ---
SUCCESSFUL_CONNECTIONS=$(grep -c "Login SUCCESS" $LOG_FILE)
FAILED_CONNECTIONS=$(grep -c "Connection FAIL" $LOG_FILE)
TOTAL_ATTEMPTS=$((SUCCESSFUL_CONNECTIONS + FAILED_CONNECTIONS))
FINAL_SESSIONS=$(grep "Session created for" $LOG_FILE | tail -n 1 | grep -oP 'Total sessions: \K\d+')

# --- A CORREÇÃO ESTÁ AQUI ---
# Este bloco usa 'tee -a' para imprimir as métricas na tela E anexá-las
# ao final do arquivo de log.
{
    echo ""
    echo "--- Métricas do Teste de Estresse ---"
    echo "Duração do Teste:            $TEST_DURATION segundos"
    echo "Usuários Concorrentes:       $CONCURRENT_USERS"
    echo "Conexões Bem-sucedidas:      $SUCCESSFUL_CONNECTIONS"
    echo "Conexões Falhas:             $FAILED_CONNECTIONS"
    if [ ! -z "$FINAL_SESSIONS" ]; then
        echo "Sessões Ativas no Servidor (pico): $FINAL_SESSIONS"
    fi
    echo "-----------------------------------"
    echo ""
    echo "Resultados completos salvos em: $LOG_FILE"
} | tee -a $LOG_FILE

echo "Teste finalizado."