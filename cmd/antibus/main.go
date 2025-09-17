package main

import (
	"bufio"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"

	"regexp"
	// regexp "github.com/wasilibs/go-re2"
)

// scan for ANSI codes
var re = regexp.MustCompile(`\x1b\[[0-9;]*[a-zA-Z]`)

func main() {
	var verbose bool
	flag.BoolVar(&verbose, "v", false, "verbose")
	flag.Parse()

	for _, fn := range flag.Args() {
		// fmt.Println(fn)
		stripFile(fn)
	}
}

func stripFile(fn string) {
	in, err := os.Open(fn)
	if err != nil {
		panic(err)
	}
	defer in.Close()

	// see if the pattern is actually in othe file
	if !doesHaveAnsiCodes(in) {
		fmt.Println("not found")
		return
	}
	fmt.Println("found")

	// ensure we are at the start of the input file
	if offset, err := in.Seek(0, 0); err != nil && offset != 0 {
		panic(err)
	}

	// open the file and ensure we are at the start of it
	out, err := os.OpenFile(fn, os.O_RDWR, 0666)
	if err != nil {
		panic(err)
	}
	defer out.Close()
	if offset, err := out.Seek(0, 0); err != nil && offset != 0 {
		panic(err)
	}

	// open the pipe
	pr, pw := io.Pipe()
	defer pw.Close()
	defer pr.Close()

	// do the sanitization
	go sanitize(pw, in, out)

	n, err := io.Copy(out, pr)
	if err != nil {
		panic(err)
	}

	stat, err := in.Stat()
	if err != nil {
		panic(err)
	}

	// if the file size is different, then pad the rest of
	// the file out with NULs
	if size := stat.Size() - n; size > 0 {
		pad := make([]byte, size)
		pad[len(pad)-1] = '\n'

		if _, err := out.Write(pad); err != nil {
			panic(err)
		}
	}

	if err := out.Sync(); err != nil {
		panic(err)
	}
}

func doesHaveAnsiCodes(r io.Reader) bool {
	s := bufio.NewReader(r)
	var err error
	var line string
	for err == nil {
		line, err = s.ReadString('\n')
		if ansiRegex.MatchString(line) {
			return true
		}
	}
	return false
}

func sanitize(pw io.WriteCloser, ro io.Reader, out io.Writer) {
	s := bufio.NewReader(ro)
	defer pw.Close()
	for {
		data, err := s.ReadBytes('\n')
		if err != nil {
			if err == io.EOF {
				break
			}
			panic(err)
		}
		if ansiRegex.Match(data) {
			data = ansiRegex.ReplaceAll(data, []byte{0x00})
		}
		if _, err := pw.Write(data); err != nil {
			panic(err)
		}
	}
}

func NewANSIStripper(in interface{}) *ANSIStripper {
	if in == nil {
		return nil
	}
	a := &ANSIStripper{}

	if r, ok := in.(io.Reader); ok {
		a.r = r
	}

	if w, ok := in.(io.Writer); ok {
		a.w = w
	}

	return a
}

type ANSIStripper struct {
	w io.Writer
	r io.Reader
}

func (a *ANSIStripper) Write(b []byte) (int, error) {
	if a.w == nil {
		return 0, errors.New("no underlying writer, ensure the passed writer implements io.Writer")
	}

	b = ansiRegex.ReplaceAll(b, []byte{0x00})
	return a.w.Write(b)
}

func (a *ANSIStripper) Read(b []byte) (int, error) {
	if a.r == nil {
		return 0, errors.New("no underlying reader, ensure the passed reader implements io.Reader")
	}

	_b := make([]byte, len(b))
	_, err := a.r.Read(_b)
	if err != nil {
		return 0, err
	}
	b = ansiRegex.ReplaceAll(_b, []byte{0x00})
	return len(b), nil
}
