package main

import (
	"context"
	"flag"
	"fmt"
	"net/url"
	"os"
	"os/exec"
	"os/user"
	"path/filepath"
	"strings"
)

const (
	name    = "gsproxy"
	version = "0.1.0"
	usage   = `
usage: gsproxy [-key-file=KEY_FILE] [-version] COMMAND gs://SRC_BUCKET/OBJECT gs://DEST_BUCKET [args...]

`
)

// Exit codes are int values that represent an exit code for a particular error.
const (
	ExitCodeOK              = 0
	ExitCodeError           = 1
	ExitCodeMisuseError     = 2
	ExitCodeCommandNotFound = 127
)

var (
	keyFile     string
	showVersion bool
)

func fatalf(format string, args ...interface{}) {
	fmt.Fprintf(os.Stderr, "%s: %s\n", name, fmt.Sprintf(format, args...))
	os.Exit(ExitCodeError)
}

func printUsage() {
	fmt.Fprint(os.Stderr, usage)
	os.Exit(ExitCodeMisuseError)
}

func isFile(path string) bool {
	fi, err := os.Stat(path)
	return err == nil && fi.Mode().IsRegular()
}

func main() {
	var proxyArgs []string
	var cmdArgs []string
	for i, x := range os.Args[1:] {
		if strings.HasPrefix(x, "-") {
			proxyArgs = append(proxyArgs, x)
		} else {
			cmdArgs = os.Args[i+1:]
			break
		}
	}

	flags := flag.NewFlagSet(name, flag.ContinueOnError)
	flags.Usage = printUsage
	flags.SetOutput(os.Stderr)
	// TODO: default
	flags.StringVar(&keyFile, "key-file", "", "Path to the Google Cloud Platform private key file.")
	flags.BoolVar(&showVersion, "version", false, "Print version information and quit.")
	err := flags.Parse(proxyArgs)
	if err != nil {
		fatalf("%s", err)
	}

	if showVersion {
		fmt.Fprintf(os.Stderr, "%s version %s\n", name, version)
		os.Exit(ExitCodeOK)
	}

	if keyFile == "" {
		fatalf("Please specify a key-file by -key-file=...")
	}

	usr, err := user.Current()
	if err != nil {
		fatalf("%s", err)
	}
	keyFile = strings.Replace(keyFile, "~", usr.HomeDir, 1)
	keyFile, err = filepath.Abs(keyFile)
	if err != nil {
		fatalf("%s", err)
	}

	if !isFile(keyFile) {
		fatalf("%s: No such file or directory", keyFile)
	}

	if len(cmdArgs) < 3 {
		printUsage()
	}

	cfg := &Config{KeyFile: keyFile}
	ctx := NewContext(context.Background(), cfg)

	err = run(ctx, cmdArgs)
	if err != nil {
		fatalf("%s", err)
	}
}

func run(ctx context.Context, args []string) error {
	src, err := url.Parse(args[1])
	if err != nil {
		return err
	}

	dest, err := url.Parse(args[2])
	if err != nil {
		return err
	}

	w, err := NewWorkspace(ctx, src, dest)
	if err != nil {
		return err
	}
	defer w.Close()

	err = w.Download(ctx)
	if err != nil {
		return err
	}

	// TODO
	a := append([]string{args[0], w.Src(), w.Dest()}, args[3:]...)
	args = append([]string{"-c", strings.Join(a, " ")})
	cmd := exec.Command("/bin/sh", args...)

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err = cmd.Run()
	if err != nil {
		return err
	}

	err = w.Upload(ctx)
	if err != nil {
		return err
	}

	return nil
}
