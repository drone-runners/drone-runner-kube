package engine

import (
	"context"
	"io"

	"github.com/drone/runner-go/livelog"
)

// cancellableCopy method copies from source to destination honoring the context.
// If context.Cancel is called, it will return immediately with context cancelled error.
func cancellableCopy(ctx context.Context, dst io.Writer, src io.ReadCloser) error {
	ch := make(chan error, 1)
	go func() {
		defer close(ch)
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
