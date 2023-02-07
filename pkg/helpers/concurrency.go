package helpers

import (
	"context"
	"fmt"
	"os"
	"os/signal"
)

func CancelOnSignal(ctx context.Context, signal_ os.Signal, cancel func()) error {
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, signal_)
	defer signal.Stop(sigCh)

	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-sigCh:
		fmt.Println("Received Ctrl+C, canceling context")
		cancel()
	}

	return nil
}
