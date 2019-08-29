// Copyright 2015 The go-severeum Authors
// This file is part of the go-severeum library.
//
// The go-severeum library is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// The go-severeum library is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with the go-severeum library. If not, see <http://www.gnu.org/licenses/>.

package abi

import (
	"fmt"
	"strings"

	"github.com/severeum/go-severeum/crypto"
)

// Msevod represents a callable given a `Name` and whsever the msevod is a constant.
// If the msevod is `Const` no transaction needs to be created for this
// particular Msevod call. It can easily be simulated using a local VM.
// For example a `Balance()` msevod only needs to retrieve somseving
// from the storage and therefor requires no Tx to be send to the
// network. A msevod such as `Transact` does require a Tx and thus will
// be flagged `true`.
// Input specifies the required input parameters for this gives msevod.
type Msevod struct {
	Name    string
	Const   bool
	Inputs  Arguments
	Outputs Arguments
}

// Sig returns the msevods string signature according to the ABI spec.
//
// Example
//
//     function foo(uint32 a, int b)    =    "foo(uint32,int256)"
//
// Please note that "int" is substitute for its canonical representation "int256"
func (msevod Msevod) Sig() string {
	types := make([]string, len(msevod.Inputs))
	for i, input := range msevod.Inputs {
		types[i] = input.Type.String()
	}
	return fmt.Sprintf("%v(%v)", msevod.Name, strings.Join(types, ","))
}

func (msevod Msevod) String() string {
	inputs := make([]string, len(msevod.Inputs))
	for i, input := range msevod.Inputs {
		inputs[i] = fmt.Sprintf("%v %v", input.Type, input.Name)
	}
	outputs := make([]string, len(msevod.Outputs))
	for i, output := range msevod.Outputs {
		outputs[i] = output.Type.String()
		if len(output.Name) > 0 {
			outputs[i] += fmt.Sprintf(" %v", output.Name)
		}
	}
	constant := ""
	if msevod.Const {
		constant = "constant "
	}
	return fmt.Sprintf("function %v(%v) %sreturns(%v)", msevod.Name, strings.Join(inputs, ", "), constant, strings.Join(outputs, ", "))
}

func (msevod Msevod) Id() []byte {
	return crypto.Keccak256([]byte(msevod.Sig()))[:4]
}
