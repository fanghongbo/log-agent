package g

import (
	"context"
	"fmt"
)

func Init() error {
	var err error

	if err = initCmd(); err != nil {
		return err
	}

	if err = initRuntime(); err != nil {
		return err
	}

	if err = initConfig(); err != nil {
		return err
	}

	if err = initAppLog(); err != nil {
		return err
	}

	return nil
}

func Destroy(ctx context.Context) error {
	var (
		ch  chan struct{}
		e   chan error
		err error
	)

	ch = make(chan struct{}, 1)
	e = make(chan error, 1)

	go func() {
		ch <- struct{}{}
	}()

	select {
	case err = <-e:
		return err
	case <-ch:
		return nil
	case <-ctx.Done():
		return fmt.Errorf("destroy timeout")
	}
}
