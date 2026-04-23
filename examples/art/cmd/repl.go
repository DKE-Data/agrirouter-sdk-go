package cmd

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"os/signal"
	"strings"
	"syscall"

	"github.com/chzyer/readline"
	"github.com/google/shlex"
	"github.com/spf13/cobra"
)

// replAllowed is the set of art subcommands the repl will run. 'serve' and
// 'repl' are excluded to keep the sandbox non-recursive; 'completion' is
// excluded because it prints shell scripts that are useless inside the repl.
var replAllowed = map[string]bool{
	"put-endpoint":                    true,
	"delete-endpoint":                 true,
	"send-messages":                   true,
	"receive-messages":                true,
	"receive-endpoint-deleted-events": true,
	"rede":                            true,
	"confirm-messages":                true,
	"help":                            true,
}

var replCmd = &cobra.Command{
	Use:   "repl",
	Short: "interactive art-only shell (used by 'serve' as a sandboxed terminal)",
	Long: `Starts a minimal interactive shell that accepts only art subcommands.
Each line is shlex-split and executed by re-invoking this same binary with
those arguments, so the available commands and flags are identical to the
regular CLI. Anything that is not a known art subcommand is rejected.

Tab completion is served by cobra's hidden '__complete' command (same source
of truth as bash/zsh completion), filtered to the allowed subcommands.

Intended to be spawned by 'art serve' to back the browser terminal, but
can also be run standalone.`,
	SilenceUsage:  true,
	SilenceErrors: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		bin, err := os.Executable()
		if err != nil {
			return fmt.Errorf("failed to resolve own executable path: %w", err)
		}

		// Don't die on ^C while a child is running; child will receive SIGINT
		// directly from the controlling tty. Between commands, readline handles
		// ^C itself (returns ErrInterrupt).
		sigC := make(chan os.Signal, 1)
		signal.Notify(sigC, syscall.SIGINT)
		defer signal.Stop(sigC)
		go func() {
			for range sigC {
			}
		}()

		rl, err := readline.NewEx(&readline.Config{
			Prompt:          "art> ",
			InterruptPrompt: "^C",
			EOFPrompt:       "exit",
			AutoComplete:    &cobraCompleter{bin: bin},
		})
		if err != nil {
			return fmt.Errorf("failed to start readline: %w", err)
		}
		defer rl.Close()

		allowed := make([]string, 0, len(replAllowed))
		for k := range replAllowed {
			allowed = append(allowed, k)
		}
		fmt.Fprintf(rl.Stderr(), "art repl — commands: %s. Builtins: clear, exit, set-tenant <uuid>.\r\n", strings.Join(allowed, ", "))
		fmt.Fprintf(rl.Stderr(), "Tab completes subcommands, flags and $ART_* env vars (expanded before exec); prefer $ART_APPLICATION_ID / $ART_SOFTWARE_VERSION_ID / $ART_TENANT_ID.\r\n")

		for {
			line, err := rl.Readline()
			if err == readline.ErrInterrupt {
				if line == "" {
					return nil
				}
				continue
			}
			if err == io.EOF {
				return nil
			}
			if err != nil {
				return err
			}
			line = strings.TrimSpace(line)
			if line == "" {
				continue
			}
			if line == "exit" || line == "quit" {
				return nil
			}
			if line == "clear" {
				fmt.Fprint(rl.Stdout(), "\x1b[2J\x1b[H")
				rl.Refresh()
				continue
			}

			argv, err := shlex.Split(line)
			if err != nil {
				fmt.Fprintf(rl.Stderr(), "parse error: %v\r\n", err)
				continue
			}
			if len(argv) == 0 {
				continue
			}
			if argv[0] == "set-tenant" {
				if len(argv) != 2 {
					fmt.Fprintf(rl.Stderr(), "usage: set-tenant <uuid>\r\n")
					continue
				}
				_ = os.Setenv("ART_TENANT_ID", argv[1])
				fmt.Fprintf(rl.Stderr(), "ART_TENANT_ID=%s\r\n", argv[1])
				continue
			}
			if !replAllowed[argv[0]] {
				fmt.Fprintf(rl.Stderr(), "not allowed: %s (allowed: %s)\r\n", argv[0], strings.Join(allowed, ", "))
				continue
			}
			for i := range argv {
				argv[i] = os.ExpandEnv(argv[i])
			}

			// Release the terminal so the child has full control of the TTY
			// (cobra commands write colors / progress, and ^C should hit them).
			rl.Clean()
			c := exec.Command(bin, argv...)
			c.Env = os.Environ()
			c.Stdin = os.Stdin
			c.Stdout = os.Stdout
			c.Stderr = os.Stderr
			if err := c.Run(); err != nil {
				fmt.Fprintf(rl.Stderr(), "%s: %v\r\n", argv[0], err)
			}
			rl.Refresh()
		}
	},
}

// cobraCompleter is a readline AutoCompleter that asks cobra's hidden
// '__complete' subcommand for suggestions. The first token is filtered to
// the repl's allowlist so hidden commands like 'serve' don't leak in.
type cobraCompleter struct {
	bin string
}

func (c *cobraCompleter) Do(line []rune, pos int) ([][]rune, int) {
	text := string(line[:pos])

	tokens, err := shlex.Split(text)
	if err != nil {
		// Fall back to whitespace split if the line has an open quote etc.
		tokens = strings.Fields(text)
	}

	endsWithSpace := pos > 0 && (line[pos-1] == ' ' || line[pos-1] == '\t')
	var partial string
	if endsWithSpace || len(tokens) == 0 {
		tokens = append(tokens, "")
		partial = ""
	} else {
		partial = tokens[len(tokens)-1]
	}

	// If the user is in the middle of typing a $VAR reference, bypass cobra
	// and offer env var names matching (ART_* / AGRIROUTER_* only, to keep
	// the menu signal-heavy).
	if cands, ok := envVarCompletions(partial); ok {
		return cands, len([]rune(partial))
	}

	args := append([]string{"__complete"}, tokens...)
	out, err := exec.Command(c.bin, args...).Output()
	if err != nil {
		return nil, 0
	}

	firstToken := len(tokens) == 1

	var candidates [][]rune
	if firstToken {
		for _, builtin := range []string{"clear", "exit", "set-tenant"} {
			if strings.HasPrefix(builtin, partial) {
				candidates = append(candidates, []rune(builtin[len(partial):]))
			}
		}
	}
	for _, raw := range strings.Split(string(out), "\n") {
		raw = strings.TrimRight(raw, "\r")
		if raw == "" || strings.HasPrefix(raw, ":") || strings.HasPrefix(raw, "Completion ended") {
			continue
		}
		val, _, _ := strings.Cut(raw, "\t")
		if firstToken && !replAllowed[val] {
			continue
		}
		if !strings.HasPrefix(val, partial) {
			continue
		}
		tail := val[len(partial):]
		candidates = append(candidates, []rune(tail))
	}
	return candidates, len([]rune(partial))
}

// envVarCompletions returns completion tails for the rightmost $VAR / ${VAR
// reference in partial, if partial ends mid-reference. Second return is false
// if the partial isn't completing a variable reference.
func envVarCompletions(partial string) ([][]rune, bool) {
	idx := strings.LastIndex(partial, "$")
	if idx < 0 {
		return nil, false
	}
	varPart := partial[idx+1:]
	hasBrace := strings.HasPrefix(varPart, "{")
	if hasBrace {
		varPart = varPart[1:]
	}
	for _, r := range varPart {
		if !(r == '_' || (r >= '0' && r <= '9') || (r >= 'A' && r <= 'Z') || (r >= 'a' && r <= 'z')) {
			return nil, false
		}
	}
	var candidates [][]rune
	for _, e := range os.Environ() {
		name, _, _ := strings.Cut(e, "=")
		if !strings.HasPrefix(name, "ART_") {
			continue
		}
		if !strings.HasPrefix(name, varPart) {
			continue
		}
		tail := name[len(varPart):]
		if hasBrace {
			tail += "}"
		}
		candidates = append(candidates, []rune(tail))
	}
	return candidates, true
}

func init() {
	rootCmd.AddCommand(replCmd)
}
