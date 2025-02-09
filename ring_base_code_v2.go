package main

import (
    "fmt"
    "sync"
)

type mensagem struct {
    tipo  int    // tipo da mensagem (eleição, confirmação, falha, novo coordenador)
    corpo [4]int // IDs dos processos envolvidos
    electionInitiator int
}

var (
    chans    = []chan mensagem{ // vetor de canais para formar o anel de eleição
        make(chan mensagem),
        make(chan mensagem),
        make(chan mensagem),
        make(chan mensagem),
    }
    controle = make(chan int)
    wg       sync.WaitGroup // wg para esperar o término do programa
)

func ElectionControler(in chan int) {
    defer wg.Done()

    var temp mensagem

    // Simula a falha do processo 0
    temp.tipo = 2
    chans[3] <- temp
    fmt.Println("Controle: mudar o processo 0 para falho")
    if confirmation := <-in; confirmation != 0 {
        fmt.Printf("Controle: confirmação %d\n", confirmation)
    }

    // Simula a falha do processo 1
    temp.tipo = 2
    chans[0] <- temp
    fmt.Println("Controle: mudar o processo 1 para falho")
    if confirmation := <-in; confirmation != 0 {
        fmt.Printf("Controle: confirmação %d\n", confirmation)
    }

        // Simula a falha do processo 1
        temp.tipo = 2
        chans[1] <- temp
        fmt.Println("Controle: mudar o processo 1 para falho")
        if confirmation := <-in; confirmation != 0 {
            fmt.Printf("Controle: confirmação %d\n", confirmation)
        }

    fmt.Println("\n   Processo controlador concluído\n")
}

func ElectionStage(TaskId int, in chan mensagem, out chan mensagem, leader int) {
    defer wg.Done()

    var actualLeader int = leader
    var bFailed bool = false      // todos iniciam sem falha
    var electionInitiator int = -1 // Identificador do iniciador da eleição
    finished := false              // controla o término da eleição

    for !finished {
        temp := <-in // ler mensagem
        fmt.Printf("%2d: recebi mensagem %d, [ %d, %d, %d, %d ]\n", TaskId, temp.tipo, temp.corpo[0], temp.corpo[1], temp.corpo[2], temp.corpo[3])

        switch temp.tipo {
        case 1: // Iniciar uma eleição
            if !bFailed { // Apenas processos ativos iniciam a eleição
                electionInitiator = TaskId
                fmt.Printf("%2d: Iniciando eleição, meu id: %d\n", TaskId, TaskId)
                temp.tipo = 5
                temp.corpo[TaskId] = 1
                temp.electionInitiator = electionInitiator
                out <- temp // Passa a mensagem para o próximo
            } else {
                out <- temp
            }

        case 2: // Marca como falho
            bFailed = true
            fmt.Printf("%2d: falho %v\n", TaskId, bFailed)
            fmt.Printf("%2d: líder atual %d\n", TaskId, actualLeader)
            controle <- TaskId // Enviar confirmação ao controlador

            // Inicia uma nova eleição ao detectar uma falha
            electionMessage := mensagem{tipo: 1} // tipo 1 para iniciar a eleição
            out <- electionMessage // Passa a mensagem de eleição para o próximo

        case 3: // Volta o falho como ativo
            bFailed = false
            fmt.Printf("%2d: falho %v\n", TaskId, bFailed)
            fmt.Printf("%2d: líder atual %d\n", TaskId, actualLeader)
            controle <- TaskId // Enviar confirmação ao controlador

        case 4: // Marca novo líder
            actualLeader = temp.corpo[0]
            bFailed = false
            fmt.Printf("%2d: novo coordenador é %d\n", TaskId, actualLeader)
            finished = true // Finaliza o processo após estabelecer o líder

        case 5: // Mensagem de eleição
            if !bFailed { // Apenas processos ativos participam da votação
                fmt.Printf("%2d: recebendo eleição, meu id: %d\n", TaskId, TaskId)
                temp.corpo[TaskId] = 1 // Registra o próprio ID
                fmt.Printf("%2d: votei\n", TaskId)
                electionInitiator = temp.electionInitiator

                // Verifica se o processo inicial já recebeu de volta a eleição
                if TaskId == electionInitiator {
                    // Define o novo coordenador com o maior ID
                    newLeader := minID(temp.corpo)
                    fmt.Printf("%2d: Eleição finalizada, o novo coordenador é %d\n", TaskId, newLeader)

                    // Envia mensagem de novo coordenador para todos
                    temp.tipo = 4
                    temp.corpo[0] = newLeader
                    temp.corpo[1] = -1
                    temp.corpo[2] = -1
                    temp.corpo[3] = -1

                    actualLeader = newLeader
                    out <- temp
                    finished = true // Finaliza o loop
                } else {
                    out <- temp // Passa a mensagem para o próximo processo
                }
            } else {
                out <- temp // Passa a mensagem para o próximo processo mesmo se falho
            }

        default:
            fmt.Printf("%2d: não conheço este tipo de mensagem\n", TaskId)
        }

        fmt.Printf("%2d: terminei\n", TaskId)
    }
}

func minID(ids [4]int) int {
    for i, id := range ids {
        if id == 1 {
            return i 
        }
    }
    return -1 // Retorna -1 caso não encontre o número "1" no array
}

func main() {
    wg.Add(5) // Adiciona contagem para cada goroutine

    // Criar os processos do anel de eleição
    go ElectionStage(0, chans[3], chans[0], 0) // este é o líder
    go ElectionStage(1, chans[0], chans[1], 0) // não é líder
    go ElectionStage(2, chans[1], chans[2], 0) // não é líder
    go ElectionStage(3, chans[2], chans[3], 0) // não é líder

    fmt.Println("\n   Anel de processos criado")

    // Criar o processo controlador
    go ElectionControler(controle)

    fmt.Println("\n   Processo controlador criado\n")

    wg.Wait() // Espera todas as goroutines terminarem

    // Fecha todos os canais após o término
    for _, ch := range chans {
        close(ch)
    }
}
