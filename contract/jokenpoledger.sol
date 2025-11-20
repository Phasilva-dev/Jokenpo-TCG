// SPDX-License-Identifier: MIT
pragma solidity ^0.8.0;

contract JokenpoLedger {
    
    // O endereço da carteira do seu Servidor (Admin).
    // Só ele pode chamar as funções que alteram o estado.
    address public gameServerAuthority;

    // Construtor: Roda uma vez quando o contrato sobe.
    // Define quem fez o deploy (você) como a autoridade.
    constructor() {
        gameServerAuthority = msg.sender;
    }

    // Modificador de segurança: Garante que só o servidor chame as funções.
    modifier onlyAuthority() {
        require(msg.sender == gameServerAuthority, "Acesso negado: Apenas o Game Server pode registrar logs.");
        _;
    }

    // ============================================================
    // ESTADO (Quem tem o quê)
    // Precisamos disso para garantir a unicidade e validar trocas.
    // ============================================================
    
    // Mapeia UUID do Jogador -> Lista de UUIDs das Cartas que ele possui
    mapping(string => string[]) private ownerAssets;
    
    // ============================================================
    // LOGS DE AUDITORIA (Eventos)
    // É isso que vai gerar seu relatório cronológico.
    // ============================================================

    // Log: Jogador abriu pacote (Entrada de ativos)
    event AuditPackOpened(uint256 timestamp, string playerId, string[] cardIds);

    // Log: Troca realizada (Transferência de ativos)
    event AuditTrade(uint256 timestamp, string fromPlayer, string toPlayer, string cardId);

    // Log: Resultado de Partida (Registro histórico)
    event AuditMatch(uint256 timestamp, string roomId, string winnerId, string loserId);

    // ============================================================
    // TRANSAÇÕES (Escrita no Livro Razão)
    // ============================================================

    // 1. Registrar Abertura de Pacote
    // Ex: "As 1h jogador A comprou um pacote com as cartas XYZ"
    function logPackOpening(string memory _playerId, string[] memory _cardIds) public onlyAuthority {
        // Adiciona as cartas ao "inventário blockchain" do jogador
        for (uint i = 0; i < _cardIds.length; i++) {
            ownerAssets[_playerId].push(_cardIds[i]);
        }
        
        // Emite o log com o timestamp atual do bloco
        emit AuditPackOpened(block.timestamp, _playerId, _cardIds);
    }

    // 2. Registrar Troca
    // Ex: "As 3h o jogador A trocou a carta X pela carta A do jogador B"
    // Nota: Para fazer troca dupla, o servidor deve chamar essa função duas vezes (A->B e B->A)
    function logTrade(string memory _fromPlayer, string memory _toPlayer, string memory _cardId) public onlyAuthority {
        require(hasAsset(_fromPlayer, _cardId), "Erro de Auditoria: O jogador de origem nao possui o ativo.");

        // Transfere a posse no estado interno
        removeAsset(_fromPlayer, _cardId);
        ownerAssets[_toPlayer].push(_cardId);

        // Emite o log
        emit AuditTrade(block.timestamp, _fromPlayer, _toPlayer, _cardId);
    }

    // 3. Registrar Partida
    // Ex: "As 4h o jogador B ganhou uma partida do jogador A"
    function logMatchResult(string memory _roomId, string memory _winnerId, string memory _loserId) public onlyAuthority {
        // Aqui não mudamos posse de cartas, apenas registramos o fato histórico.
        emit AuditMatch(block.timestamp, _roomId, _winnerId, _loserId);
    }

    // ============================================================
    // LEITURA (Para verificar integridade)
    // ============================================================

    // Verifica o que o jogador tem na blockchain
    function getPlayerAssets(string memory _playerId) public view returns (string[] memory) {
        return ownerAssets[_playerId];
    }

    // Função auxiliar interna para verificar posse
    function hasAsset(string memory _ownerId, string memory _assetId) internal view returns (bool) {
        string[] memory assets = ownerAssets[_ownerId];
        for(uint i=0; i < assets.length; i++) {
            if(keccak256(bytes(assets[i])) == keccak256(bytes(_assetId))) {
                return true;
            }
        }
        return false;
    }

    // Função auxiliar interna para remover posse
    function removeAsset(string memory _ownerId, string memory _assetId) internal {
        string[] storage assets = ownerAssets[_ownerId];
        for(uint i=0; i < assets.length; i++) {
            if(keccak256(bytes(assets[i])) == keccak256(bytes(_assetId))) {
                assets[i] = assets[assets.length - 1];
                assets.pop();
                return;
            }
        }
    }
}