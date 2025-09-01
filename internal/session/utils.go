package session

import (
	"regexp"
)

// parseCardKeysFromPackageString extrai as chaves das cartas (ex: "rock:5:red")
// de uma string formatada retornada por PurchasePackage.
func parseCardKeysFromPackageString(packageResult string) []string {
	// Esta expressão regular procura por um padrão que se parece com uma chave de carta.
	// rock|paper|scissor : 1-10 : red|green|blue
	re := regexp.MustCompile(`(rock|paper|scissor):\d{1,2}:(red|green|blue)`)
	
	// Encontra todas as ocorrências do padrão na string.
	matches := re.FindAllString(packageResult, -1)
	
	return matches
}