/*
 * Created: 2022-09-10 16:26:53
 * Author : Win-Man
 * Email : gang.shen0423@gmail.com
 * -----
 * Last Modified:
 * Modified By:
 * -----
 * Description:
 */

package main

import (
	"os"

	"github.com/Win-Man/ticheck/cmd"
)

func main() {
	// cobra already prints the error message, here we only need a non-zero
	// exit code so that scripts and CI can tell the check failed.
	if err := cmd.Execute(); err != nil {
		os.Exit(1)
	}
}
