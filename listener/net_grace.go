package listener

import (
	"encoding/json"
	"github.com/glory-cd/utils/log"
	"os"
	"os/signal"
	"syscall"
)

// Signal processing function
func gracefulHandle() {
	// Register signal
	signals := make(chan os.Signal)
	signal.Notify(signals, syscall.SIGHUP, syscall.SIGTERM, syscall.SIGINT)
	// Monitoring signal
	for sig := range signals {
		switch sig {
		case syscall.SIGINT, syscall.SIGTERM:
			log.Slogger.Infof("Recieve signal：%s", sig)
			os.Exit(0)
		case syscall.SIGHUP:
			// Upon receipt of SIGHUP, forkexec restarts
			execSpec := &syscall.ProcAttr{
				Env:   os.Environ(),
				Files: []uintptr{os.Stdin.Fd(), os.Stdout.Fd(), os.Stderr.Fd()},
			}
			// Fork exec the new process
			fork, err := syscall.ForkExec(os.Args[0], os.Args, execSpec)
			if err != nil {
				log.Slogger.Fatalf("Fail to fork: %s", err.Error())
			}
			log.Slogger.Infof("SIGHUP received: fork-exec to %d", fork)

			log.Slogger.Infof("Server gracefully shutdown: %d", os.Getpid())

			// Stop the old server, all the connections have been closed and the new one is running
			os.Exit(0)
		}
	}
}

// Handling functions for grace instructions
func dealReceiveGraceCMD(graceJSON string) {

	m := make(map[string]interface{})

	err := json.Unmarshal([]byte(graceJSON), &m)

	if err != nil {
		log.Slogger.Errorf("ConvertGraceJsonTOMapObject Err:[%s]", err.Error())
		return
	}
	// Get current process
	p, err := os.FindProcess(os.Getpid())
	if err != nil {
		log.Slogger.Errorf("FindProcess Err:[%s]", err.Error())
		return
	}
	// Send a signal to the process
	switch m["gracecmd"] {
	case "SIGHUP":
		err = p.Signal(syscall.SIGHUP)
		if err != nil {
			log.Slogger.Errorf("Sent signal Err:[%s]", err.Error())
		}
	case "SIGTERM":
		err = p.Signal(syscall.SIGTERM)
		if err != nil {
			log.Slogger.Errorf("Sent signal Err:[%s]", err.Error())
		}
	case "SIGINT":
		err = p.Signal(syscall.SIGINT)
		if err != nil {
			log.Slogger.Errorf("Sent signal Err:[%s]", err.Error())
		}
	}

}
