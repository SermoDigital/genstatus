// Copyright 2015 Sermo Digital, LLC. All rights reserved.
// Use of this source code is governed by the MIT License
// that can be found in the LICENSE file.

package main

import (
	"bufio"
	"bytes"
	"flag"
	"fmt"
	"log"
	"os"
	"os/exec"
	"strings"
)

const (
	scrape = `set -eo pipefail && curl -s https://golang.org/src/net/http/status.go?m=text | \
	sed -e '/const /,/\/\//!d' | \
	sed -r 's/= [0-9]{3}//g; s/const \(//g; s/\)//g'`
	fname     = "generated_status_codes.go"
	dne       = "// generated by 'stringer %s'; DO NOT EDIT.%s"
	constDecl = "\n\n%s\n\n// %dXX %s\nconst (\n"
)

var pkg = flag.String("package", "main", "")

func main() {
	flag.Parse()

	args := ""
	if len(os.Args) > 1 {
		args = os.Args[1]
	}
	doNotEdit := fmt.Sprintf(dne, args, "\n\n")

	var buf bytes.Buffer

	tmp, err := os.Create(".genstatustmpfile")
	if err != nil {
		log.Fatalln(err)
	}

	defer func() {
		tmp.Close()
		os.Remove(tmp.Name())
	}()

	cmd := exec.Command("bash", "-c", scrape)

	cmd.Stdout = &buf
	cmd.Stderr = &buf

	if err := cmd.Run(); err != nil {
		log.Fatalln(err)
	}

	f2, err := os.Create(fname)
	if err != nil {
		log.Fatalln(err)
	}
	defer f2.Close()

	f2.WriteString(doNotEdit)
	f2.WriteString(fmt.Sprintf("package %s", *pkg))

	f2.WriteString(`
import "net/http"

// These status codes are here because we sometimes return status codes that
// aren't in the net/http package, and so instead of randomly importing said
// package, we just run this go:generate command which creates the tmps of
// status codes for us. :-)


`)

	f2.WriteString(fmt.Sprintf(constDecl, doNotEdit, 1, level[1]))

	var (
		i    = 2
		init = true
		s    = bufio.NewScanner(&buf)
	)

	for s.Scan() {
		t := strings.TrimSpace(s.Text())

		if !strings.HasPrefix(t, "//") {

			if t != "" {
				f2.WriteString(fmt.Sprintf("\t%s = http.%s\n", t, t))
				init = false
			} else if !init && level[i] != "" {
				f2.WriteString(fmt.Sprintf(")"+constDecl, doNotEdit, i, level[i]))
				i++
			}

			switch t {
			case "StatusTeapot":
				f2.WriteString(`StatusChill = 420
StatusUnprocessableEntity = 422
StatusTooManyRequests = 429
StatusRequestHeaderFieldsTooLarge = 431
`)
			case "StatusHTTPVersionNotSupported":
				f2.WriteString(`StatusNetworkAuthenticationRequired = 511
				`)
			}
		}

		// debug.PrintStack()
	}

	f2.WriteString(")")

	if err := exec.Command("go", "fmt", fname).Run(); err != nil {
		log.Fatalln(err)
	}
}

var level = map[int]string{
	1: "Informational",
	2: "Success",
	3: "Redirection",
	4: "Client Error",
	5: "Server Error",
}
