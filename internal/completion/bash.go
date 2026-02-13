package completion

import (
	"fmt"
	"os"
	"path"
)

// GenerateBash generates bash completion script
func GenerateBash() {
	scriptName := os.Args[0]
	scriptName = path.Base(scriptName)
	// see https://pkg.go.dev/github.com/jessevdk/go-flags
	fmt.Printf(`
_completion_%s() {
    # All arguments except the first one
    args=("${COMP_WORDS[@]:1:$COMP_CWORD}")

    # Only split on newlines
    local IFS=$'\n'

    # Call completion (note that the first element of COMP_WORDS is
    # the executable itself)
    COMPREPLY=($(GO_FLAGS_COMPLETION=1 ${COMP_WORDS[0]} "${args[@]}"))
    return 0
}

complete -F _completion_%s %s
`, scriptName, scriptName, scriptName)
}
