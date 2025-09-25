#!/bin/bash

# --- CONFIGURAÇÃO DO TESTE ---
PACK_OPENER_COUNT=0
PINGER_COUNT=0
MATCHMAKER_COUNT=200
TEST_DURATION=300
COMPOSE_FILE="docker-compose.mixed-test.yml"
LOG_FILE="mixed_test_results.log"

echo "--- Jokenpo TCG Mixed Behavior Stress Test ---"
echo "Limpando ambiente anterior..."
docker-compose -f $COMPOSE_FILE down --volumes

echo ""
echo "Construindo as imagens Docker..."
docker-compose -f $COMPOSE_FILE build
if [ $? -ne 0 ]; then
    echo "Falha no build do Docker. Abortando."
    exit 1
fi

echo ""
echo "Iniciando teste de estresse por $TEST_DURATION segundos com:"
echo "- $PACK_OPENER_COUNT bots comprando pacotes"
echo "- $PINGER_COUNT bots medindo ping"
echo "- $MATCHMAKER_COUNT bots em matchmaking"

# Inicia o docker-compose EM SEGUNDO PLANO, escalando cada serviço de bot
docker-compose -f $COMPOSE_FILE up \
  --scale bot-pack-opener=$PACK_OPENER_COUNT \
  --scale bot-pinger=$PINGER_COUNT \
  --scale bot-matchmaker=$MATCHMAKER_COUNT \
  > $LOG_FILE 2>&1 &

# Barra de progresso
echo "Teste em andamento:"
for i in $(seq $TEST_DURATION); do
    printf "."
    sleep 1
done
echo " Tempo esgotado."

echo ""
echo "Parando o teste e coletando resultados..."
docker-compose -f $COMPOSE_FILE down --volumes >> $LOG_FILE 2>&1

# --- Coleta de Métricas ---
TOTAL_LOGIN_SUCCESS=$(grep -c "Login SUCCESS" $LOG_FILE)
TOTAL_LOGIN_FAIL=$(grep -c "FAIL.*Login" $LOG_FILE)

PACK_BOT_SUCCESS=$(grep -c "SUCCESS (PACK_OPENER)" $LOG_FILE)
PINGER_BOT_SUCCESS=$(grep -c "SUCCESS (PINGER)" $LOG_FILE)
MATCHMAKER_LOGIN_SUCCESS=$(grep -c "SUCCESS (MATCHMAKER): Login complete" $LOG_FILE)

# --- Apresentação dos Resultados ---
{
    echo ""
    echo "--- Métricas do Teste Misto ---"
    echo "Duração: $TEST_DURATION segundos"
    echo ""
    echo "[Métricas Gerais]"
    echo "Logins Bem-sucedidos (Total): $TOTAL_LOGIN_SUCCESS"
    echo "Falhas de Login/Conexão (Total): $TOTAL_LOGIN_FAIL"
    echo ""
    echo "[Métricas por Personalidade]"
    echo "Bots Compradores de Pacotes:"
    echo "  - Compras de pacotes realizadas: $PACK_BOT_SUCCESS"
    echo "Bots Pingers:"
    echo "  - Pings bem-sucedidos: $PINGER_BOT_SUCCESS"
    echo "Bots de Matchmaking:"
    echo "  - Logins bem-sucedidos: $MATCHMAKER_LOGIN_SUCCESS"
    echo "-----------------------------------"
    echo ""
    echo "Resultados completos salvos em: $LOG_FILE"
} | tee -a $LOG_FILE

echo "Teste finalizado."