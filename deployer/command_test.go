package deployer

import (
	"fmt"
	"os/exec"
	"testing"
)

func TestCommandA(t *testing.T) {
	cmd := exec.Command("sh", "-c", "cd /home/ilisin/workplace/dockers/nsq && docker compose down -v && docker compose up -d")
	data, err := cmd.CombinedOutput()
	// // cmd := exec.Command("docker", "info")
	// stdout, err := cmd.StdoutPipe()
	// if err != nil {
	// 	log.Fatal(err)
	// }
	// if err = cmd.Start(); err != nil {
	// 	log.Fatal(err)
	// }
	// data, err := io.ReadAll(stdout)
	// if err != nil {
	// 	log.Fatal(err)
	// }
	// if err = cmd.Wait(); err != nil {
	// 	log.Fatal(err)
	// }
	// err := cmd.Run()
	if err != nil {
		t.Fatal(err)
	}
	fmt.Printf("%s\n", string(data))
	// stderr, err := cmd.StderrPipe()
	// if err != nil {
	// 	log.Fatal(err)
	// }
	// errout, err := io.ReadAll(stderr)
	// if err != nil {
	// 	log.Fatal(err)
	// }
	// fmt.Printf("%v\n", string(errout))
}
