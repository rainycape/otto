package otto

import (
	"bytes"
	"encoding/gob"
	"errors"
	"strings"

	"github.com/rainycape/otto/parser"
)

var ErrVersion = errors.New("version mismatch")

var scriptVersion = "2014-04-13/1"

// Script is a handle for some (reusable) JavaScript.
// Passing a Script value to a run method will evaluate the JavaScript.
//
type Script struct {
	version string
	program *_nodeProgram
	src     string
}

// Compile will parse the given source and return a Script value or nil and
// an error if there was a problem during compilation.
//
//      script, err := vm.Compile("", `var abc; if (!abc) abc = 0; abc += 2; abc;`)
//      vm.Run(script)
//
func (self *Otto) Compile(filename string, src interface{}) (*Script, error) {
	{
		src, err := parser.ReadSource(filename, src)
		if err != nil {
			return nil, err
		}

		program, err := self.runtime.parse(filename, src)
		if err != nil {
			return nil, err
		}

		cmpl_program := cmpl_parse(program)

		script := &Script{
			version: scriptVersion,
			program: cmpl_program,
			src:     string(src),
		}

		return script, nil
	}
}

// Filename returns the Script filename, or the empty string
// if no filename was specified when it was parsed.
func (self *Script) Filename() string {
	if self.program != nil && self.program.file != nil {
		return self.program.file.Name()
	}
	return ""

}

func (self *Script) String() string {
	filename := "<unknown>"
	if n := self.Filename(); n != "" {
		filename = n
	}
	prefix := "// "
	if strings.Contains(filename, "://") {
		prefix = ""
	}
	return prefix + filename + "\n" + self.src
}

// MarshalBinary will marshal a script into a binary form. A marshalled script
// that is later unmarshalled can be executed on the same version of the otto runtime.
//
// The binary format can change at any time and should be considered unspecified and opaque.
//
func (self *Script) marshalBinary() ([]byte, error) {
	var bfr bytes.Buffer
	encoder := gob.NewEncoder(&bfr)
	err := encoder.Encode(self.version)
	if err != nil {
		return nil, err
	}
	err = encoder.Encode(self.program)
	if err != nil {
		return nil, err
	}
	err = encoder.Encode(self.src)
	if err != nil {
		return nil, err
	}
	return bfr.Bytes(), nil
}

// UnmarshalBinary will vivify a marshalled script into something usable. If the script was
// originally marshalled on a different version of the otto runtime, then this method
// will return an error.
//
// The binary format can change at any time and should be considered unspecified and opaque.
//
func (self *Script) unmarshalBinary(data []byte) error {
	decoder := gob.NewDecoder(bytes.NewReader(data))
	err := decoder.Decode(&self.version)
	if err != nil {
		goto error
	}
	if self.version != scriptVersion {
		err = ErrVersion
		goto error
	}
	err = decoder.Decode(&self.program)
	if err != nil {
		goto error
	}
	err = decoder.Decode(&self.src)
	if err != nil {
		goto error
	}
	return nil
error:
	self.version = ""
	self.program = nil
	self.src = ""
	return err
}
