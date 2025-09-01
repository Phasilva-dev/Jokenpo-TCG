// card/rules.go
package card

// Constantes para representar o resultado da comparação de cartas.
// Usar constantes torna o código que utiliza esta função muito mais legível.
const (
	Card1Wins = 1
	Card2Wins = -1
	Tie       = 0
)

// winConditions define a regra primária do Jokenpo.
// A chave vence o valor. Ex: "rock" vence "scissor".
var winConditions = map[string]string{
	"rock":    "scissor",
	"scissor": "paper",
	"paper":   "rock",
}

// Compare executa a lógica de batalha completa entre duas cartas.
// Ela primeiro compara os tipos. Se os tipos forem iguais, usa o valor como desempate.
// Retorna uma das constantes: Card1Wins, Card2Wins, or Tie.
func Compare(card1, card2 *Card) int {
	type1 := card1.Typo()
	type2 := card2.Typo()

	// --- Etapa 1: Comparação por Tipo (Regra Principal do Jokenpo) ---

	// Verifica se o tipo da carta 1 vence o tipo da carta 2.
	if winConditions[type1] == type2 {
		return Card1Wins
	}

	// Verifica se o tipo da carta 2 vence o tipo da carta 1.
	if winConditions[type2] == type1 {
		return Card2Wins
	}

	// Se chegamos aqui, significa que os tipos são os mesmos (ou não há relação de vitória,
	// o que no Jokenpo significa que são iguais). Agora, vamos para o desempate.

	// --- Etapa 2: Comparação por Valor (Desempate) ---

	value1 := card1.Value()
	value2 := card2.Value()
	
	if value1 > value2 {
		return Card1Wins
	}
	
	if value2 > value1 {
		return Card2Wins
	}

	// --- Etapa 3: Empate Verdadeiro ---
	
	// Se os tipos são iguais e os valores também, é um empate.
	return Tie
}