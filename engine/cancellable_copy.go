package engine

import (
	"context"
	"io"

	"github.com/drone/runner-go/livelog"
)

// cancellableCopy method provides a way to copy from source to destination
// that honors context. If context.Cancel is called, copy method will return immediately.
func cancellableCopy(ctx context.Context, dst io.Writer, src io.ReadCloser) error {
	ch := make(chan error)
	go func() {
		err := livelog.Copy(dst, src)
		ch <- err
	}()

	select {
	case <-ctx.Done():
		src.Close()
		return ctx.Err()
	case err := <-ch:
		return err
	}
}
