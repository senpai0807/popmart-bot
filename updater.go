package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"
	"time"
)

func main() {
	if len(os.Args) != 3 {
		fmt.Println("Usage: updater <target_exe> <new_exe>")
		pause()
		return
	}

	targetExe, err := filepath.Abs(os.Args[1])
	if err != nil {
		fmt.Printf("Failed to get absolute path for target exe: %s\n", err.Error())
		pause()
		return
	}

	newExe, err := filepath.Abs(os.Args[2])
	if err != nil {
		fmt.Printf("Failed to get absolute path for new exe: %s\n", err.Error())
		pause()
		return
	}

	fmt.Printf("Waiting for %s to close...\n", filepath.Base(targetExe))

	for {
		if !isProcessRunning(filepath.Base(targetExe)) {
			break
		}
		time.Sleep(1 * time.Second)
	}

	fmt.Println("Target application closed. Proceeding with update...")

	err = os.Remove(targetExe)
	if err != nil {
		fmt.Printf("Failed to delete old executable: %s\n", err.Error())
		pause()
		return
	}

	err = os.Rename(newExe, targetExe)
	if err != nil {
		fmt.Printf("Failed to move new executable: %s\n", err.Error())
		pause()
		return
	}

	fmt.Println("Update successful! Launching new LunarTools.exe...")

	cmd := exec.Command("cmd", "/C", "start", "", targetExe)
	cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: false}
	cmd.Start()

	pause()
}

func isProcessRunning(processName string) bool {
	tasklistCmd := exec.Command("tasklist")
	output, err := tasklistCmd.Output()
	if err != nil {
		return false
	}
	return strings.Contains(string(output), processName)
}

func pause() {
	fmt.Println("Press any key to exit...")
	exec.Command("cmd", "/C", "pause").Run()
}
