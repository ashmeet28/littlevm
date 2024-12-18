package main

import (
	"fmt"
	"log"
	"os"
)

type VMContext struct {
	MM  []byte // Main Memory
	ECM []byte // Environment Call Memory

	BCM []byte // Bytecode Memory
	SM  []byte // Stack Memory

	PC uint64 // Program Counter
	FP uint64 // Frame Pointer
	SP uint64 // Stack Pointer

	Status int
}

const (
	VMS_ILLEGAL int = iota

	VMS_HALT
	VMS_RUNNING
	VMS_ECALL
)

var (
	OP_HALT  byte = 0x01
	OP_ECALL byte = 0x02

	OP_CALL   byte = 0x04
	OP_RETURN byte = 0x05

	OP_JUMP   byte = 0x08
	OP_BRANCH byte = 0x09

	OP_PUSH   byte = 0x0c
	OP_POP    byte = 0x0d
	OP_ASSIGN byte = 0x0e

	OP_ADD byte = 0x40
	OP_SUB byte = 0x41

	OP_AND byte = 0x44
	OP_OR  byte = 0x45
	OP_XOR byte = 0x46

	OP_SHL byte = 0x48
	OP_SHR byte = 0x49

	OP_MUL byte = 0x4c
	OP_QUO byte = 0x4d
	OP_REM byte = 0x4e

	OP_EQL byte = 0x50
	OP_NEQ byte = 0x51
	OP_LSS byte = 0x52
	OP_GTR byte = 0x53
	OP_LEQ byte = 0x54
	OP_GEQ byte = 0x55

	OP_CONVERT byte = 0x58

	OP_LOAD  byte = 0x20
	OP_STORE byte = 0x21

	OP_STORE_STRING byte = 0x22
)

func VMInit(bytecode []byte) VMContext {
	vm := VMContext{
		MM:  make([]byte, 0x1000),
		ECM: make([]byte, 0x1000),

		BCM: make([]byte, 0x1000),
		SM:  make([]byte, 0x1000),

		PC: 0,
		FP: 0,
		SP: 0,

		Status: VMS_RUNNING,
	}

	if len(bytecode) > len(vm.BCM) {
		PrintErrorAndExit("Bytecode size exceeds the permitted limits!")
	}

	copy(bytecode, vm.BCM)

	return vm
}

func VMTick(vm VMContext) {
	op := vm.BCM[vm.PC]

	switch op {
	case OP_HALT:
		vm.Status = VMS_HALT
		vm.PC++
	case OP_ECALL:
		vm.Status = VMS_ECALL
		vm.PC++

	case OP_CALL:
	case OP_RETURN:

	case OP_JUMP:
	case OP_BRANCH:

	case OP_PUSH:
	case OP_POP:
	case OP_ASSIGN:

	case OP_ADD:
	case OP_SUB:

	case OP_AND:
	case OP_OR:
	case OP_XOR:

	case OP_SHL:
	case OP_SHR:

	case OP_MUL:
	case OP_QUO:
	case OP_REM:

	case OP_EQL:
	case OP_NEQ:
	case OP_LSS:
	case OP_GTR:
	case OP_LEQ:
	case OP_GEQ:

	case OP_CONVERT:

	case OP_LOAD:
	case OP_STORE:

	case OP_STORE_STRING:
	}

}

func PrintErrorAndExit(s string) {
	fmt.Println("Error: " + s)
	os.Exit(1)
}

func main() {
	bytecodeFilePath := os.Args[1]

	data, err := os.ReadFile(bytecodeFilePath)

	if err != nil {
		log.Fatal(err)
	}

	VMInit(data)
}
