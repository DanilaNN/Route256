package workerpool

import (
	"context"
)

// Результат работы воркера для пользователя
type Either[T any] struct {
	Value *T
	Err   error
}

// Возвращаем пользователю канал, из которого можно будет прочитать результат, когда он будет готов
type Promise[T any] struct {
	Ch chan Either[T]
}

// Абстракция задачи, которая будет использоваться вокруг функции step (step будет выполняться асинхронно)
type Task[I, O any] struct {
	in     I // что на входе функции step
	ctx    context.Context
	result Promise[O] // выход функции step записывается в этот канал
}

type Future[I, O any] struct {
	tasks chan *Task[I, O] // Канал с входными задачами, которые будут выполняться асинхронно
}

func NewFuture[I, O any](ctx context.Context, limit int,
	step func(ctx context.Context, i I) (*O, error),
) *Future[I, O] {

	// создаем  workerpool, по типу future - т.е. результат из него сможем забрать, когда он будет готов
	wp := &Future[I, O]{tasks: make(chan *Task[I, O])}

	// закрываем канал с входными задачами при закрытии контекста
	go func() {
		<-ctx.Done()
		close(wp.tasks)
	}()

	// запускаем горутины-обработчики, которые считывают задачи из входного канала
	for i := 0; i < limit; i++ {
		go func() {
			for in := range wp.tasks { // ожидаем задач в канале
				result := Either[O]{}
				result.Value, result.Err = step(in.ctx, in.in) // считали задачу и выполняем её в горутине-обработчике
				in.result.Ch <- result                         // кладем результат в выходной канал, блокируемся, пока из него не прочитают
			}
		}()
	}
	return wp
}

func (wp *Future[I, O]) Exec(ctx context.Context, i I) Promise[O] {
	result := Promise[O]{Ch: make(chan Either[O], 1)} // создаем канал с результатом выполнения задачи
	select {
	case <-ctx.Done(): // возращаем ошибку, если отменили контекст
		result.Ch <- Either[O]{Err: ctx.Err()}
	case wp.tasks <- &Task[I, O]{ // отправляем задачу на выполнение
		in:     i,
		ctx:    ctx,
		result: result,
	}:
	}
	return result
}
