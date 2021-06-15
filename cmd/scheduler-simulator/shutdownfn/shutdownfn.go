package shutdownfn

import (
	"log"
	"os"
	"os/signal"
	"sync"
	"syscall"
)

// Shutdownfn represents the function handle to be called, typically in a defer handler, to shutdown a running module.
type Shutdownfn func()

// WaitShutdown waits for signal and exec given shutdownFns.
func WaitShutdown(shutdownFns ...Shutdownfn) {
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGTERM, os.Interrupt)
	<-quit

	wg := &sync.WaitGroup{}
	for i := range shutdownFns {
		wg.Add(1)
		s := shutdownFns[i]
		go func() {
			s()
			wg.Done()
		}()
	}
	wg.Wait()
	log.Println("finish shutdown successfully")
}
