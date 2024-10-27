// Código exemplo para o trabaho de sistemas distribuidos (eleicao em anel)
// By Cesar De Rose - 2022

package main

import (
	"fmt"
	"sync"
)

type mensagem struct {
	tipo  int    // tipo da mensagem para fazer o controle do que fazer (eleição, confirmacao da eleicao)
	corpo [3]int // conteudo da mensagem para colocar os ids (usar um tamanho ocmpativel com o numero de processos no anel)
}

var (
	chans = []chan mensagem{ // vetor de canias para formar o anel de eleicao - chan[0], chan[1] and chan[2] ...
		make(chan mensagem),
		make(chan mensagem),
		make(chan mensagem),
		make(chan mensagem),
	}
	controle = make(chan int)
	wg       sync.WaitGroup // wg is used to wait for the program to finish
)

func ElectionControler(in chan int) {
	defer wg.Done()
	var temp mensagem

	// Mudar o processo 0 para falho
	temp.tipo = 2
	chans[3] <- temp
	fmt.Printf("Controle: mudar o processo 0 para falho\n")
	fmt.Printf("Controle: confirmação %d\n", <-in) // receber confirmação

	// Mudar o processo 1 para falho
	temp.tipo = 2
	chans[0] <- temp
	fmt.Printf("Controle: mudar o processo 1 para falho\n")
	fmt.Printf("Controle: confirmação %d\n", <-in) // receber confirmação

	// Enviar uma mensagem de eleição para iniciar o processo de eleição
	temp.tipo = 1 // Tipo de mensagem de eleição
	chans[2] <- temp
	fmt.Printf("Controle: iniciar eleição pelo processo 2\n")

	// Recuperar processos (opcional, pode ser removido)
	temp.tipo = 4
	chans[1] <- temp
	chans[2] <- temp
}


func ElectionStage(TaskId int, in chan mensagem, out chan mensagem, leader int) {
	defer wg.Done()

	// Variáveis locais para controle
	var actualLeader int = leader // Coordenador atual
	var bFailed bool = false      // Estado de falha
	electionInProgress := false   // Indica se a eleição está em andamento
	temp := mensagem{tipo: 1}     // Tipo 1 indica mensagem de eleição

	// Função auxiliar para iniciar uma eleição
	startElection := func() {
		electionInProgress = true
		temp.tipo = 1                      // Inicia eleição
		temp.corpo = [3]int{TaskId, -1, -1} // Define o próprio ID
		fmt.Printf("%2d: Iniciando eleição...\n", TaskId)
		out <- temp
	}

	for {
		select {
		case temp = <-in:
			fmt.Printf("%2d: Recebi mensagem %d, [ %d, %d, %d ]\n", TaskId, temp.tipo, temp.corpo[0], temp.corpo[1], temp.corpo[2])

			switch temp.tipo {
			case 1: // Mensagem de eleição
				if bFailed {
					// Se está falho, ignora a mensagem
					fmt.Printf("%2d: Estou falho, ignorando mensagem.\n", TaskId)
					continue
				}

				// Inclui o ID do processo na mensagem se houver espaço
				added := false
				for i := 0; i < len(temp.corpo); i++ {
					if temp.corpo[i] == -1 {
						temp.corpo[i] = TaskId
						added = true
						break
					}
				}

				// Verifica se completou uma volta
				if temp.corpo[0] == TaskId && added {
					// Define o novo coordenador com o maior ID
					newLeader := maxID(temp.corpo)
					fmt.Printf("%2d: Eleição concluída. Novo coordenador é %d\n", TaskId, newLeader)
					actualLeader = newLeader

					// Envia mensagem de coordenador para todos os processos
					temp.tipo = 3 // Tipo 3 para mensagem de novo coordenador
					out <- temp
					electionInProgress = false
				} else {
					// Propaga a mensagem de eleição para o próximo processo ativo
					out <- temp
				}

			case 3: // Mensagem de coordenador
				// Atualiza o coordenador atual
				actualLeader = maxID(temp.corpo)
				fmt.Printf("%2d: Coordenador atualizado para %d\n", TaskId, actualLeader)

			case 2: // Mensagem de falha
				bFailed = true
				fmt.Printf("%2d: Falho %v \n", TaskId, bFailed)
				controle <- -5 // Confirma a falha para o controle externo

			case 4: // Recuperação do processo
				bFailed = false
				fmt.Printf("%2d: Falha recuperada, ativo novamente\n", TaskId)

			default:
				fmt.Printf("%2d: Mensagem desconhecida\n", TaskId)
			}

		default:
			// Se o coordenador falhar, inicia nova eleição
			if actualLeader == -1 && !electionInProgress && !bFailed {
				startElection()
			}
		}
	}
}

// Função auxiliar para obter o maior ID na mensagem
func maxID(ids [3]int) int {
	max := ids[0]
	for _, id := range ids {
		if id > max && id != -1 {
			max = id
		}
	}
	return max
}


func main() {

	wg.Add(5) // Add a count of four, one for each goroutine

	// criar os processo do anel de eleicao

	go ElectionStage(0, chans[3], chans[0], 0) // este é o lider
	go ElectionStage(1, chans[0], chans[1], 0) // não é lider, é o processo 0
	go ElectionStage(2, chans[1], chans[2], 0) // não é lider, é o processo 0
	go ElectionStage(3, chans[2], chans[3], 0) // não é lider, é o processo 0

	fmt.Println("\n   Anel de processos criado")

	// criar o processo controlador

	go ElectionControler(controle)

	fmt.Println("\n   Processo controlador criado\n")

	wg.Wait() // Wait for the goroutines to finish\
}
