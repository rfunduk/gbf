package main

import (
	"bytes"
	"fmt"
	"os"
)

const COMMANDS = "+-<>[].,"
const REPEATABLE_COMMANDS = "+-<>."

type BfOp struct {
	Index       uint32
	Command     byte
	RepeatCount uint8
	TargetIndex uint32
}

func (op BfOp) String() string {
	return fmt.Sprintf(
		"{BFOP-%05d cmd=%c repeat=%d target=%d}",
		op.Index,
		op.Command,
		op.RepeatCount,
		op.TargetIndex,
	)
}

func usage() {
	fmt.Println("Usage: gbf input.bf")
}

func main() {
	args := os.Args[1:]
	if len(args) != 1 {
		usage()
		os.Exit(1)
	}

	source, err := os.ReadFile(args[0])
	if err != nil {
		usage()
		fmt.Println(fmt.Errorf("ERROR: unable to read source file: %s", err))
		os.Exit(1)
	}

	source = preprocess(source)
	// fmt.Printf("Preprocessed Source: %s\n", string(source))

	ops := translate(source)
	fmt.Printf("Intermediate Representation:\n")
	for _, op := range ops {
		fmt.Printf("\t%s\n", op)
	}

	// Run the program
	fmt.Printf("\nOutput:\n")
	execute(ops)
	fmt.Println()
}

func execute(ops []BfOp) {
	const MAX_MEM uint16 = 3000
	var mem [MAX_MEM]uint8

	var ip uint16 = 0 // instruction pointer -- where we are in the program
	var ptr int32 = 0 // memory pointer -- where the cursor is in memory

	for ip < uint16(len(ops)) {
		op := ops[ip]
		//fmt.Printf("OP: %v, IP: %d, PTR: %d, MEM: %#x\n", op, ip, ptr, mem[ptr])

		switch op.Command {
		case '+':
			mem[ptr] += op.RepeatCount
		case '-':
			mem[ptr] -= op.RepeatCount
		case '>':
			ptr = (ptr + int32(op.RepeatCount)) % int32(MAX_MEM)
		case '<':
			// you cant use mod to wrap with negative numbers, thanks obama
			ptr = (ptr - int32(op.RepeatCount) + int32(MAX_MEM)) % int32(MAX_MEM)
		case '.':
			fmt.Printf(string(bytes.Repeat([]byte{mem[ptr]}, int(op.RepeatCount))))
		case ',':
			var b []byte = make([]byte, 1)
			os.Stdin.Read(b)
			mem[ptr] = b[0]
		case '[':
			if mem[ptr] == 0 {
				ip = uint16(op.TargetIndex)
			}
		case ']':
			if mem[ptr] != 0 {
				ip = uint16(op.TargetIndex)
			}

		// internal optimised commands
		case 'z':
			mem[ptr] = 0

		}

		ip++
	}
}

// convert source code from a list of command characters
// to an array of operations with extra metadata
func translate(source []byte) []BfOp {
	repeatableBytes := []byte(REPEATABLE_COMMANDS)
	var sourceLength = uint32(len(source))
	var ops = []BfOp{}
	jumpsStack := []uint32{}
	var opIndex uint32 = 0

	for i := uint32(0); i < sourceLength; i++ {
		op := BfOp{
			Index:       opIndex,
			Command:     source[i],
			RepeatCount: 1,
		}

		// consume more commands and add to RepeatCount
		// if there are multiple in a row.
		// only certain commands are repeatable
		for {
			if !bytes.Contains(repeatableBytes, []byte{op.Command}) ||
				i+1 >= sourceLength ||
				source[i+1] != op.Command {
				break
			}
			op.RepeatCount++
			i++
		}

		// for jumpstart [
		// add the index to the jumpsStack
		if source[i] == '[' {
			jumpsStack = append(jumpsStack, opIndex)
		}

		// for jumpend ]
		// pop the last [ and set the jump targets appropriately
		if source[i] == ']' {
			// golang has no stack type :/
			var start uint32
			stackSize := len(jumpsStack)
			jumpsStack, start = jumpsStack[:stackSize-1], jumpsStack[stackSize-1]
			op.TargetIndex = start
			ops[start].TargetIndex = opIndex
		}

		ops = append(ops, op)
		opIndex++

		// optimise [-] to just zero out the current memory location
		// to do this we can look at the last 3 commands and
		// see if they match, then remove them and replace with
		// a custom internal command 'z' which does this
		totalOps := len(ops)
		if ops[totalOps-1].Command == ']' &&
			ops[totalOps-2].Command == '-' &&
			ops[totalOps-3].Command == '[' {
			// remove existing '[-]'
			ops = ops[:totalOps-3]
			opIndex -= 3

			// add 'z'
			optimisedOp := BfOp{
				Index:       opIndex,
				Command:     'z',
				RepeatCount: 1,
			}
			ops = append(ops, optimisedOp)
			opIndex++
		}

		// could do other optimisations here
		// https://www.nayuki.io/page/optimizing-brainfuck-compiler

	}

	return ops
}

// remove comments and other non-command characters
func preprocess(source []byte) []byte {
	commandBytes := []byte(COMMANDS)
	var commands = []byte{}

	for _, char := range source {
		if bytes.Contains(commandBytes, []byte{char}) {
			commands = append(commands, char)
		}
	}

	return commands
}
