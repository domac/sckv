package sckv

import (
	"bytes"
	"testing"
)

//go test -v -run TestSimpleCommand
func TestSimpleSetCommand(t *testing.T) {
	cmd := "*3\r\n$3\r\nSET\r\n$5\r\nmykey\r\n$7\r\nmyvalue\r\n"
	br := bytes.NewReader([]byte(cmd))
	reqCmd := NewRequestCmd(br)
	cmds, err := reqCmd.ParseCommand()
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("cmds: %s\n", cmds)
}

func TestSimpleGetCommand(t *testing.T) {
	cmd := "*2\r\n$3\r\nGET\r\n$5\r\nmykey\r\n"
	br := bytes.NewReader([]byte(cmd))
	reqCmd := NewRequestCmd(br)
	cmds, err := reqCmd.ParseCommand()
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("cmds: %s\n", cmds)
}

func TestSimpleExpireCommand(t *testing.T) {
	cmd := "*3\r\n$6\r\nEXPIRE\r\n$5\r\nmykey\r\n$2\r\n30\r\n"
	br := bytes.NewReader([]byte(cmd))
	reqCmd := NewRequestCmd(br)
	cmds, err := reqCmd.ParseCommand()
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("cmds: %s\n", cmds)
}

func TestSimpleGetPipelineCommand(t *testing.T) {
	cmd := "*2\r\n$3\r\nGET\r\n$11\r\nmykey123456\r\n*2\r\n$3\r\nGET\r\n$5\r\nmyCmd\r\n*2\r\n$3\r\nGET\r\n$11\r\nmykey123456\r\n*2\r\n$3\r\nGET\r\n$5\r\nmyCmd\r\n*2\r\n$3\r\nGET\r\n$11\r\nmykey123456\r\n*2\r\n$3\r\nGET\r\n$5\r\nmyCmd\r\n"
	br := bytes.NewReader([]byte(cmd))
	reqCmd := NewRequestCmd(br)
	cmds, err := reqCmd.ParseCommand()
	if err != nil {
		t.Fatal(err)
	}
	if len(cmds) != 6 {
		t.Fail()
	}
	t.Logf("cmds: %s\n", cmds)
}

func BenchmarkGetCommand(b *testing.B) {
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			cmd := "*2\r\n$3\r\nGET\r\n$5\r\nmykey\r\n"
			br := bytes.NewReader([]byte(cmd))
			reqCmd := NewRequestCmd(br)
			_, err := reqCmd.ParseCommand()
			if err != nil {
				b.Error(err)
			}
		}
	})
}
