/*
 * Created: 2020-04-11 16:53:49
 * Author : Win-Man
 * Email : gang.shen0423@gmail.com
 * -----
 * Last Modified:
 * Modified By:
 * -----
 * Description:
 */

package linux

import (
	"os"
	"strconv"
	"testing"
)

// TestExecCommand is an integration test against a real Linux host.
// It is skipped unless TICHECK_TEST_SSH_HOST is set; credentials come from
// environment variables so no secret is committed to the repository.
//
//	TICHECK_TEST_SSH_HOST=10.0.0.1
//	TICHECK_TEST_SSH_USER=root
//	TICHECK_TEST_SSH_PASSWORD=...
//	TICHECK_TEST_SSH_PORT=22          # optional, defaults to 22
//	TICHECK_TEST_SSH_CMD="uname -a"   # optional, defaults to "uname -a"
func TestExecCommand(t *testing.T) {
	host := os.Getenv("TICHECK_TEST_SSH_HOST")
	if host == "" {
		t.Skip("skipping SSH integration test: set TICHECK_TEST_SSH_HOST to enable")
	}

	port := 22
	if p := os.Getenv("TICHECK_TEST_SSH_PORT"); p != "" {
		parsed, err := strconv.Atoi(p)
		if err != nil {
			t.Fatalf("invalid TICHECK_TEST_SSH_PORT %q: %v", p, err)
		}
		port = parsed
	}

	user := os.Getenv("TICHECK_TEST_SSH_USER")
	if user == "" {
		user = "root"
	}

	command := os.Getenv("TICHECK_TEST_SSH_CMD")
	if command == "" {
		command = "uname -a"
	}

	myssh := Myssh{
		Host:     host,
		User:     user,
		Password: os.Getenv("TICHECK_TEST_SSH_PASSWORD"),
		Port:     port,
		Type:     "password",
	}
	res, err := myssh.ExecCommand(command)
	if err != nil {
		t.Fatalf("exec command %q on %s failed:%v", command, host, err)
	}
	t.Logf("output of %q: %s", command, string(res))
}
