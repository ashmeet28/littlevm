package main

import (
	"fmt"
	"log"
	"os"
	"time"
)

type VMContext struct {
	MM  []byte // Main Memory
	ECM []byte // Environment Call Memory

	BM []byte // Bytecode Memory
	SM []byte // Stack Memory

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

		BM: make([]byte, 0x1000),
		SM: make([]byte, 0x1000),

		PC: 0,
		FP: 0,
		SP: 0,

		Status: VMS_RUNNING,
	}

	if len(bytecode) > len(vm.BM) {
		PrintErrorAndExit("Bytecode size exceeds the permitted limits!")
	}

	copy(vm.BM, bytecode)

	return vm
}

func VMValR(b []byte, s byte) uint64 {
	if (s != 1) && (s != 2) && (s != 4) && (s != 8) {
		PrintErrorAndExit("Invalid instruction!")
	}

	if len(b) < int(s) {
		PrintErrorAndExit("Invalid memory read!")
	}

	var v uint64
	var i byte

	for i = 0; i < s; i++ {
		v = v | (uint64(b[i]) << (8 * i))
	}

	return v
}

func VMValW(b []byte, s byte, v uint64) []byte {
	if (s != 1) && (s != 2) && (s != 4) && (s != 8) {
		PrintErrorAndExit("Invalid instruction!")
	}

	if len(b) < int(s) {
		PrintErrorAndExit("Invalid memory write!")
	}

	var i byte

	for i = 0; i < s; i++ {
		b[i] = byte((v >> (8 * i)) & 0xff)
	}

	return b
}

func VMValInfoIsValid(b byte) bool {
	if (b & 0b11000000) != 0 {
		return false
	}

	s := (b & 0b1111)

	if (s != 1) && (s != 2) && (s != 4) && (s != 8) {
		return false
	}

	return true
}

func VMValInfoSize(b byte) byte {
	if !VMValInfoIsValid(b) {
		PrintErrorAndExit("Invalid instruction!")
	}

	return (b & 0b1111)
}

func VMValInfoIsSigned(b byte) bool {
	if !VMValInfoIsValid(b) {
		PrintErrorAndExit("Invalid instruction!")
	}

	return ((b & 0b10000) == 0b10000)
}

func VMValInfoIsIndirect(b byte) bool {
	if !VMValInfoIsValid(b) {
		PrintErrorAndExit("Invalid instruction!")
	}

	return ((b & 0b100000) == 0b100000)
}

func VMPopVal(vm VMContext, valInfo byte) (VMContext, uint64) {
	var v uint64

	if VMValInfoIsIndirect(valInfo) {
		va := VMValR(vm.SM[vm.SP-8:], 8)
		vm.SP -= 8

		v = VMValR(vm.SM[vm.FP+va:], VMValInfoSize(valInfo))
	} else {
		v = VMValR(vm.SM[vm.SP-uint64(VMValInfoSize(valInfo)):], VMValInfoSize(valInfo))

		vm.SP = vm.SP - uint64(VMValInfoSize(valInfo))
	}

	return vm, v
}

func VMTick(vm VMContext) VMContext {

	switch vm.BM[vm.PC] {

	case OP_HALT:

		vm.Status = VMS_HALT
		vm.PC += 1

	case OP_ECALL:

		vm.Status = VMS_ECALL
		vm.PC += 1

	case OP_CALL:

		va := VMValR(vm.SM[vm.SP-8:], 8)
		vm.SP -= 8

		VMValW(vm.SM[vm.SP:], 8, vm.FP)
		vm.SP += 8
		VMValW(vm.SM[vm.SP:], 8, vm.PC+1)
		vm.SP += 8

		vm.FP = vm.SP
		vm.PC = vm.PC + va

	case OP_RETURN:

		va := VMValR(vm.SM[vm.SP-8:], 8)
		vm.SP -= 8

		vx := VMValR(vm.SM[vm.FP-8:], 8)
		vy := VMValR(vm.SM[vm.FP-16:], 8)

		b1 := vm.BM[vm.PC+1]

		if b1 == 0 {

			vm.SP = va + vm.FP

		} else {

			var vj uint64
			vm, vj = VMPopVal(vm, b1)

			vm.SP = va + vm.FP

			VMValW(vm.SM[vm.SP:], VMValInfoSize(b1), vj)
			vm.SP = vm.SP + uint64(VMValInfoSize(b1))

		}

		vm.PC = vx
		vm.FP = vy

	case OP_JUMP:
	case OP_BRANCH:

	case OP_PUSH:

		b1 := vm.BM[vm.PC+1]

		VMValW(vm.SM[vm.SP:], VMValInfoSize(b1), VMValR(vm.BM[vm.PC+2:], VMValInfoSize(b1)))

		vm.SP = vm.SP + uint64(VMValInfoSize(b1))
		vm.PC = vm.PC + 2 + uint64(VMValInfoSize(b1))

	case OP_POP:

		b1 := vm.BM[vm.PC+1]

		vm.SP = vm.SP - uint64(VMValInfoSize(b1))
		vm.PC += 2

	case OP_ASSIGN:

	case OP_ADD:

		b1 := vm.BM[vm.PC+1]
		b2 := vm.BM[vm.PC+2]

		if (b1 & 0b11111) != (b2 & 0b11111) {
			PrintErrorAndExit("Invalid instruction!")
		}

		var vj uint64
		var vk uint64

		vm, vk = VMPopVal(vm, b2)
		vm, vj = VMPopVal(vm, b1)

		VMValW(vm.SM[vm.SP:], VMValInfoSize(b1), vj+vk)
		vm.SP = vm.SP + uint64(VMValInfoSize(b1))

		vm.PC += 3

	case OP_SUB:

		b1 := vm.BM[vm.PC+1]
		b2 := vm.BM[vm.PC+2]

		if (b1 & 0b11111) != (b2 & 0b11111) {
			PrintErrorAndExit("Invalid instruction!")
		}

		var vj uint64
		var vk uint64

		vm, vk = VMPopVal(vm, b2)
		vm, vj = VMPopVal(vm, b1)

		VMValW(vm.SM[vm.SP:], VMValInfoSize(b1), vj+((^vk)+1))
		vm.SP = vm.SP + uint64(VMValInfoSize(b1))

		vm.PC += 3

	case OP_AND:

		b1 := vm.BM[vm.PC+1]
		b2 := vm.BM[vm.PC+2]

		if (b1 & 0b11111) != (b2 & 0b11111) {
			PrintErrorAndExit("Invalid instruction!")
		}

		var vj uint64
		var vk uint64

		vm, vk = VMPopVal(vm, b2)
		vm, vj = VMPopVal(vm, b1)

		VMValW(vm.SM[vm.SP:], VMValInfoSize(b1), vj&vk)
		vm.SP = vm.SP + uint64(VMValInfoSize(b1))

		vm.PC += 3

	case OP_OR:

		b1 := vm.BM[vm.PC+1]
		b2 := vm.BM[vm.PC+2]

		if (b1 & 0b11111) != (b2 & 0b11111) {
			PrintErrorAndExit("Invalid instruction!")
		}

		var vj uint64
		var vk uint64

		vm, vk = VMPopVal(vm, b2)
		vm, vj = VMPopVal(vm, b1)

		VMValW(vm.SM[vm.SP:], VMValInfoSize(b1), vj|vk)
		vm.SP = vm.SP + uint64(VMValInfoSize(b1))

		vm.PC += 3

	case OP_XOR:

		b1 := vm.BM[vm.PC+1]
		b2 := vm.BM[vm.PC+2]

		if (b1 & 0b11111) != (b2 & 0b11111) {
			PrintErrorAndExit("Invalid instruction!")
		}

		var vj uint64
		var vk uint64

		vm, vk = VMPopVal(vm, b2)
		vm, vj = VMPopVal(vm, b1)

		VMValW(vm.SM[vm.SP:], VMValInfoSize(b1), vj^vk)
		vm.SP = vm.SP + uint64(VMValInfoSize(b1))

		vm.PC += 3

	case OP_SHL:

		b1 := vm.BM[vm.PC+1]
		b2 := vm.BM[vm.PC+2]

		var vj uint64
		var vk uint64

		vm, vk = VMPopVal(vm, b2)
		vm, vj = VMPopVal(vm, b1)

		if VMValInfoIsSigned(b2) &&
			((vk & (uint64(1) << (uint64(VMValInfoSize(b2)*8) - 1))) ==
				(uint64(1) << (uint64(VMValInfoSize(b2)*8) - 1))) {

			PrintErrorAndExit("Negative shift count!")

		}

		VMValW(vm.SM[vm.SP:], VMValInfoSize(b1), vj<<vk)
		vm.SP = vm.SP + uint64(VMValInfoSize(b1))

		vm.PC += 3

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

	default:
		PrintErrorAndExit("Invalid instruction!")
	}

	return vm

}

func VMRun(vm VMContext) {
	for vm.Status == VMS_RUNNING {
		vm = VMTick(vm)
		VMPrint(vm)
		time.Sleep(time.Millisecond * 250)
	}

	if vm.Status == VMS_HALT {
		VMPrint(vm)
	}
}

func VMPrint(vm VMContext) {
	fmt.Println(vm.PC)
	fmt.Println(vm.FP)
	fmt.Println(vm.SP)
	fmt.Println(vm.SM[:16])
	fmt.Println(vm.SM[16 : 16*2])
	fmt.Println(vm.SM[16*2 : 16*3])
	fmt.Println(vm.SM[16*3 : 16*4])
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

	VMRun(VMInit(data))
}
