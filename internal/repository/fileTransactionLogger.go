package repository

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"kv-store/internal/event"
	"os"
	"sync"
)

//TODO тесты
//TODO метод Close
//TODO gracefull shutdown события из буфера
//TODO кодировка сообщений в файле
//TODO ограничение в размерах ключей и значений
//TODO сжимать журнал \ хранить бинарные данные
//TODO дедупликация событий об удаленных ключах

type FileTransactionLogger struct {
	events       chan<- event.TransactionEvent
	errors       <-chan error
	lastSequence uint64
	file         *os.File
	wg           sync.WaitGroup
}

func NewTransactionLogger(fileName string) (TransactionLogger, error) {
	file, err := os.OpenFile(fileName, os.O_RDWR|os.O_APPEND|os.O_CREATE, 0755)
	if err != nil {
		return nil, err
	}

	return &FileTransactionLogger{file: file}, nil
}

func (l *FileTransactionLogger) WritePut(key string, value string) {
	l.wg.Add(1)
	l.events <- event.TransactionEvent{
		EventType: event.EventPut,
		Key:       key,
		Value:     value,
	}
}

func (l *FileTransactionLogger) WriteDelete(key string) {
	l.wg.Add(1)
	l.events <- event.TransactionEvent{
		EventType: event.EventDelete,
		Key:       key,
	}
}

func (l *FileTransactionLogger) Err() <-chan error {
	return l.errors
}

func (l *FileTransactionLogger) Run() {
	events := make(chan event.TransactionEvent, 16)
	l.events = events

	errors := make(chan error, 1)
	l.errors = errors

	go func() {
		for e := range events {
			l.lastSequence++

			_, err := fmt.Fprintf(
				l.file,
				"%d\t%d\t%s\t%s\n",
				l.lastSequence, e.EventType, e.Key, e.Value,
			)

			if err != nil {
				errors <- err
				return
			}

			l.wg.Done()
		}
	}()
}

func (l *FileTransactionLogger) ReadEvent() (<-chan event.TransactionEvent, <-chan error) {
	scanner := bufio.NewScanner(l.file)
	outEvent := make(chan event.TransactionEvent)
	outError := make(chan error, 1)

	go func() {
		defer close(outEvent)
		defer close(outError)

		var e event.TransactionEvent

		for scanner.Scan() {
			line := scanner.Text()

			_, err := fmt.Sscanf(line, "%d\t%d\t%s\t%s", &e.Sequence, &e.EventType, &e.Key, &e.Value)
			if err != nil {
				outError <- fmt.Errorf("input parce error: %w", err)
				return
			}

			if l.lastSequence >= e.Sequence {
				outError <- fmt.Errorf("transaction numbers out of sequence")
				return
			}

			l.lastSequence = e.Sequence

			outEvent <- e
		}

		if scanner.Err() != nil {
			outError <- fmt.Errorf("can not read transaction file: %w", scanner.Err())
			return
		}
	}()

	return outEvent, outError
}

func (l *FileTransactionLogger) Close(ctx context.Context) error {
	completing := make(chan struct{})
	errCh := make(chan error)

	go func() {
		l.wg.Wait()

		if l.events != nil {
			close(l.events)
		}

		err := l.file.Close()
		if err != nil {
			errCh <- err
			return
		}

		completing <- struct{}{}
	}()

	select {
	case <-completing:
		return nil
	case err := <-errCh:
		return err
	case <-ctx.Done():
		return errors.New("Graceful shutdown timeout")
	}
}
