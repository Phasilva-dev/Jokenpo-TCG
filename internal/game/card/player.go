package card

import (
	"fmt"
	"sort"
	"strings"
)

// CardInstance = carta de um jogador, pode ter várias cópias
type CardInstance struct {
	card  *Card // referência ao catálogo global
	count uint   // número de cópias
}

func (ci *CardInstance) Count() uint { return ci.count }

func (ci *CardInstance) Card() *Card { return ci.card }

func (ci *CardInstance) String() string {
    if ci == nil || ci.card == nil {
        return "<nil card instance>"
    }
    // Retorna algo como: "[x2] Rock:5:Red"
    return fmt.Sprintf("[x%d] %s", ci.count, ci.card)
}

type PlayerCollection struct {
	collection map[string]*CardInstance
}

func NewPlayerCollection() *PlayerCollection {
	return &PlayerCollection{
		collection: make(map[string]*CardInstance),
	}
}

// GetInstance retorna a CardInstance completa (carta + contagem) para uma chave.
func (pc *PlayerCollection) GetInstance(key string) (*CardInstance, error) {
	instance, ok := pc.collection[key]
	if !ok {
		return nil, fmt.Errorf("player dont have this card in collection'%s'", key)
	}
	return instance, nil
}

func (pc *PlayerCollection) AddCard(key string, num uint) error {
	//key == typo string, value uint8, color string

	ci, ok := pc.collection[key]
	if ok {
		ci.count += num
	} else {
		card, err := GetCard(key)
		if err != nil {
			return err
		}
		// se não existe, cria nova CardInstance
		pc.collection[key] = &CardInstance{
			card:  card,
			count: num,
		}
	}
	return nil
}

// GetCard busca por uma carta na coleção do jogador pela sua chave.
// Retorna um ponteiro para a Card se o jogador possuir pelo menos uma instância,
// caso contrário, retorna um erro.
func (pc *PlayerCollection) GetCard(key string) (*Card, error) {
	// 1. Tenta obter a CardInstance do mapa da coleção.
	// A variável 'ok' será 'true' se a chave existir, e 'false' caso contrário.
	instance, ok := pc.collection[key]

	// 2. Verifica se a chave foi encontrada.
	if !ok {
		// Se 'ok' for false, a carta não existe na coleção. Retorna um erro.
		return nil, fmt.Errorf("card with key '%s' not found in player's collection", key)
	}

	// 3. Se a chave foi encontrada, a 'instance' é válida.
	// Como uma entrada só existe no mapa se a contagem for > 0 (assumindo uma boa
	// prática na função de remover cartas), podemos retornar o ponteiro da carta.
	// Usamos o método auxiliar `Card()` da instância para obter o ponteiro.
	return instance.Card(), nil
}

// String implementa a interface fmt.Stringer para PlayerCollection.
func (pc *PlayerCollection) String() string {
	if len(pc.collection) == 0 {
		return "players collection is empty"
	}

	keys := make([]string, 0, len(pc.collection))
	for k := range pc.collection {
		keys = append(keys, k)
	}
	sort.Strings(keys) // Ordena as chaves alfabeticamente

	var sb strings.Builder
	sb.WriteString("--- Player Collection ---\n")

	for _, key := range keys {
		instance := pc.collection[key]
		sb.WriteString(instance.String() + "\n")
	}

	sb.WriteString("--------------------------")
	return sb.String()
}



