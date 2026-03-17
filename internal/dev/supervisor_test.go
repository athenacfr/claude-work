package dev

import (
	"net"
	"os"
	"os/exec"
	"testing"
	"time"
)

func TestWaitForPort_AlreadyListening(t *testing.T) {
	ln, err := net.Listen("tcp", "localhost:0")
	if err != nil {
		t.Fatal(err)
	}
	defer ln.Close()
	port := ln.Addr().(*net.TCPAddr).Port

	start := time.Now()
	if err := waitForPort(port, nil, 5*time.Second); err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if time.Since(start) > 2*time.Second {
		t.Fatal("took too long for an already-listening port")
	}
}

func TestWaitForPort_BecomesReady(t *testing.T) {
	ln, err := net.Listen("tcp", "localhost:0")
	if err != nil {
		t.Fatal(err)
	}
	port := ln.Addr().(*net.TCPAddr).Port
	ln.Close()

	// Start listening after 1 second.
	go func() {
		time.Sleep(1 * time.Second)
		ln2, err := net.Listen("tcp", ln.Addr().String())
		if err != nil {
			return
		}
		defer ln2.Close()
		// Keep listening until test ends.
		time.Sleep(10 * time.Second)
	}()

	if err := waitForPort(port, nil, 5*time.Second); err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
}

func TestWaitForPort_Timeout(t *testing.T) {
	// Use a port that's not listening.
	ln, err := net.Listen("tcp", "localhost:0")
	if err != nil {
		t.Fatal(err)
	}
	port := ln.Addr().(*net.TCPAddr).Port
	ln.Close()

	start := time.Now()
	err = waitForPort(port, nil, 1*time.Second)
	if err == nil {
		t.Fatal("expected timeout error, got nil")
	}
	elapsed := time.Since(start)
	if elapsed < 800*time.Millisecond {
		t.Fatalf("returned too quickly: %v", elapsed)
	}
}

func TestWaitForPort_ProcessDied(t *testing.T) {
	// Start a process that exits immediately.
	cmd := exec.Command("true")
	if err := cmd.Start(); err != nil {
		t.Fatal(err)
	}
	proc := cmd.Process
	cmd.Wait() // let it finish

	// Use a port that's not listening.
	ln, err := net.Listen("tcp", "localhost:0")
	if err != nil {
		t.Fatal(err)
	}
	port := ln.Addr().(*net.TCPAddr).Port
	ln.Close()

	start := time.Now()
	err = waitForPort(port, proc, 10*time.Second)
	if err == nil {
		t.Fatal("expected error for dead process, got nil")
	}
	// Should return quickly, not wait for full timeout.
	if time.Since(start) > 3*time.Second {
		t.Fatal("took too long to detect dead process")
	}
}

func TestWaitForPort_ProcessDied_Zombie(t *testing.T) {
	// On Linux, after cmd.Wait(), the process is reaped and Kill(pid, 0)
	// returns an error. This test verifies that waitForPort detects it.
	cmd := exec.Command("sleep", "0")
	if err := cmd.Start(); err != nil {
		t.Fatal(err)
	}
	proc := cmd.Process
	// Wait to reap the zombie.
	cmd.Wait()

	// Verify our assumption: Kill should fail on reaped process.
	if err := proc.Signal(os.Signal(nil)); err == nil {
		// Process somehow still alive, skip.
		t.Skip("process still alive after Wait")
	}

	ln, err := net.Listen("tcp", "localhost:0")
	if err != nil {
		t.Fatal(err)
	}
	port := ln.Addr().(*net.TCPAddr).Port
	ln.Close()

	err = waitForPort(port, proc, 5*time.Second)
	if err == nil {
		t.Fatal("expected error")
	}
}
