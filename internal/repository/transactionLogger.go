package repository

import (
	"context"
	"kv-store/internal/event"
)

type TransactionLogger interface {
	WriteDelete(key string)
	WritePut(key, value string)
	Err() <-chan error
	Run()
	ReadEvent() (<-chan event.TransactionEvent, <-chan error)
	Close(context.Context) error
}
