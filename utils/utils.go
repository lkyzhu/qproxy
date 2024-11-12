package utils

import (
	"io"
	"sync"
)

func BindIO(src io.ReadWriter, dest io.ReadWriter) error {
	wg := &sync.WaitGroup{}

	wg.Add(1)
	go ioCopy(wg, src, dest)

	wg.Add(1)
	go ioCopy(wg, dest, src)

	wg.Wait()
	return nil
}

func ioCopy(wg *sync.WaitGroup, src io.ReadWriter, dest io.ReadWriter) error {
	defer wg.Done()

	_, err := io.Copy(dest, src)

	return err
}
