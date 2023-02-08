package main

import (
	"context"
	"log"
	"os"
	"os/exec"
	"os/signal"
)

// ControlCContext returns a context that is canceled when the application receives an Interrupt
// signal (Control-C).  The signal receiving is one-shot; after receiving interrupt we stop
// listening for the signal, and behavior reverts to its default.  That means that if the program is
// unresponsive to the context being cancelled, you can press Control-C again for a less graceful
// termination.
func ControlCContext() (context.Context, func()) {
	ctx, c := context.WithCancel(context.Background())
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt)
	go func() {
		_, ok := <-sigCh
		if ok {
			log.Println("interrupt")
			signal.Stop(sigCh)
			c()
		}
		// If not ok, this is the close(sigCh) from the cancel function we return.
	}()

	return ctx, func() {
		signal.Stop(sigCh)
		close(sigCh)
		c()
	}
}

func main() {
	ctx, cancel := ControlCContext()
	defer cancel()
	cmd := exec.CommandContext(ctx, "postgres", "-D", "/usr/local/var/postgres")
	if err := cmd.Start(); err != nil {
		panic(err)
	}
	cmd = exec.CommandContext(ctx, "etcd")
	if err := cmd.Start(); err != nil {
		panic(err)
	}
	if err := cmd.Wait(); err != nil {
		panic(err)
	}
}
