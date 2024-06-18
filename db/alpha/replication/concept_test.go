package replication

import (
	"bufio"
	"os"
	"testing"
)

func TestReadUnCommitedAppendLogFile(t *testing.T) {
	// This is an example of write and read the same file at the same time.

	wf, err := os.OpenFile("test_read_uncommited_append_log_file", os.O_CREATE|os.O_TRUNC|os.O_APPEND, 0666)
	if err != nil {
		t.Fatal(err)
	}
	defer func(wf *os.File) {
		_ = wf.Close()
	}(wf)

	rf, err := os.OpenFile("test_read_uncommited_append_log_file", os.O_RDONLY, 0666)
	if err != nil {
		t.Fatal(err)
	}
	defer func(rf *os.File) {
		_ = rf.Close()
	}(rf)

	_, _ = wf.WriteString("hello world\n")
	_, _ = wf.WriteString("hello world\n")
	_, _ = wf.WriteString("hello world\n")
	_ = wf.Sync()

	t.Log("========>>>> read:")
	buf := bufio.NewReader(rf)
	t.Log(buf.ReadString('\n'))
	t.Log(buf.ReadString('\n'))
	t.Log(buf.ReadString('\n'))
	t.Log(buf.ReadString('\n'))

	t.Log("========>>>> write more:")
	_, _ = wf.WriteString("hello world\n")
	_ = wf.Sync()
	t.Log(buf.ReadString('\n'))
	t.Log(buf.ReadString('\n'))

}
