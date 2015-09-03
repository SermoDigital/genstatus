// Copyright 2015 Sermo Digital, LLC. All rights reserved.
// Use of this source code is governed by the MIT License
// that can be found in the LICENSE file.

package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"os/exec"
	"strings"
)

const (
	cmd = `set -eo pipefail && curl -s https://golang.org/src/net/http/status.go?m=text | \
	sed -e '/const /,/)/!d' | \
	sed -r 's/= [0-9]{3}//g; s/const \(//g; s/\)//g'`
	fname     = "generated_status_codes.go"
	doNotEdit = "// AUTOMATICALLY GENERATED. DO NOT EDIT."
	constDecl = "\n\n%s\n\n// %dXX %s\nconst (\n"
)

func main() {
	if err := exec.Command("bash", "-c", cmd).Run(); err != nil {
		log.Fatalln(err)
	}

	file, err := os.Open("st.txt")
	if err != nil {
		log.Fatalln(err)
	}
	defer file.Close()

	f2, err := os.Create(fname)
	if err != nil {
		log.Fatalln(err)
	}
	defer f2.Close()

	f2.WriteString(doNotEdit + "\n")
	f2.WriteString(`
package api

import "net/http"

// These status codes are here because we sometimes return status codes that
// aren't in the net/http package, and so instead of randomly importing said
// package, we just run this go:generate command which creates the files of
// status codes for us. :-)


`)

	i := 0
	f2.WriteString(fmt.Sprintf(constDecl, doNotEdit, i+1, level[i+1]))

	s := bufio.NewScanner(file)
	for s.Scan() {
		t := strings.TrimSpace(s.Text())

		if t != "" {
			f2.WriteString(fmt.Sprintf("\t%s = http.%s\n", t, t))
		} else {
			i++
			f2.WriteString(fmt.Sprintf(")"+constDecl, doNotEdit, i+1, level[i]))
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

	f2.WriteString(")")

	if err := exec.Command("go", "fmt", fname).Run(); err != nil {
		log.Fatalln(err)
	}
}

var level = [...]string{
	"Informational",
	"Success",
	"Redirection",
	"Client Error",
	"Server Error",
}
