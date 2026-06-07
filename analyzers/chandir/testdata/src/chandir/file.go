// Eval для GID-189 (Google: channel direction).
package chandir

// --- Позитивные кейсы (двунаправленный параметр-канал) ---

// Функция с параметром-каналом без направления.
func consume(ch chan int) { // want `GID-189: channel parameter ch is bidirectional\. Fix: declare a direction, <-chan to receive or chan<- to send\.`
	<-ch
}

type worker struct{}

// Метод с параметром-каналом без направления.
func (w worker) run(ch chan string) { // want `GID-189: channel parameter ch is bidirectional\. Fix: declare a direction, <-chan to receive or chan<- to send\.`
	ch <- "x"
}

// Функциональный литерал с параметром-каналом без направления.
func withLit() {
	f := func(ch chan int) { // want `GID-189: channel parameter ch is bidirectional\. Fix: declare a direction, <-chan to receive or chan<- to send\.`
		<-ch
	}
	_ = f
}

// Несколько имён в одной группе параметров — диагностика на группу.
func multi(a, b chan int) { // want `GID-189: channel parameter a, b is bidirectional\. Fix: declare a direction, <-chan to receive or chan<- to send\.`
	<-a
	<-b
}

// --- Негативные кейсы (направление указано) ---

// Только чтение.
func recvOnly(ch <-chan int) {
	<-ch
}

// Только запись.
func sendOnly(ch chan<- int) {
	ch <- 1
}

// --- Граничные кейсы (не матчатся) ---

// Возврат двунаправленного канала — владельцу бывает нужен, решает review.
func produce() chan int {
	return make(chan int)
}

// Поле структуры с типом канала — направление задаётся при передаче.
type holder struct {
	ch chan int
}

// Локальная переменная-канал — это создание канала, не сигнатура.
func local() {
	var ch chan int
	_ = ch
}

// Именованный тип-канал в позиции параметра — осознанное решение.
type Pipe chan int

func namedParam(p Pipe) {
	<-p
}

// Срез каналов — не прямой параметр-канал.
func sliceParam(chs []chan int) {
	_ = chs
}
