package engine

import (
	"bufio"
	"context"
	"io"
)

// cancellableCopy method copies from source to destination honoring the context.
// If context.Cancel is called, it will return immediately with context cancelled error.
//
// Returns the total number of bytes copied into the buffer.
func cancellableCopy(ctx context.Context, dst io.Writer, src io.ReadCloser) (written int, err error) {
	eCh := make(chan error, 1)
	go func() {
		defer close(eCh)
		// Copy the live log
		r := bufio.NewReader(src)
		for {
			// Read the buffer back
			bytes, rErr := r.ReadBytes('\n')
			// Write the buffer to the destination
			w, wErr := dst.Write(bytes)
			written += w
			// Check if there was a failure to write
			if wErr != nil {
				eCh <- wErr
				return
			}
			// Check if there was a failure to read
			if rErr != nil {
				// If this was _not_ an EOF error, its' remarkable. Return it.
				if rErr != io.EOF {
					eCh <- rErr
					return
				}
				// If this was an EOF error, it can be ignored.
				return
			}
		}
	}()
	select {
	case <-ctx.Done():
		src.Close()
		return written, ctx.Err()
	case err := <-eCh:
		return written, err
	}
}
