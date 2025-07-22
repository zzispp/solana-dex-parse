package solanaswapgo

import (
	"bytes"

	"github.com/gagliardetto/solana-go"
	"github.com/mr-tron/base58"
)

// isTransfer checks if the instruction is a token transfer (Raydium, Orca)
func (p *Parser) isTransfer(instr solana.CompiledInstruction) bool {
	// Add bounds checking for ProgramIDIndex
	if int(instr.ProgramIDIndex) >= len(p.allAccountKeys) {
		p.Log.Warnf("ProgramIDIndex %d is out of range (allAccountKeys length: %d) in isTransfer check", instr.ProgramIDIndex, len(p.allAccountKeys))
		return false
	}
	progID := p.allAccountKeys[instr.ProgramIDIndex]

	if !progID.Equals(solana.TokenProgramID) {
		return false
	}

	if len(instr.Accounts) < 3 || len(instr.Data) < 9 {
		return false
	}

	if instr.Data[0] != 3 {
		return false
	}

	for i := 0; i < 3; i++ {
		if int(instr.Accounts[i]) >= len(p.allAccountKeys) {
			return false
		}
	}

	return true
}

// isTransferCheck checks if the instruction is a token transfer check (Meteora)
func (p *Parser) isTransferCheck(instr solana.CompiledInstruction) bool {
	// Add bounds checking for ProgramIDIndex
	if int(instr.ProgramIDIndex) >= len(p.allAccountKeys) {
		p.Log.Warnf("ProgramIDIndex %d is out of range (allAccountKeys length: %d) in isTransferCheck", instr.ProgramIDIndex, len(p.allAccountKeys))
		return false
	}
	progID := p.allAccountKeys[instr.ProgramIDIndex]

	if !progID.Equals(solana.TokenProgramID) && !progID.Equals(solana.Token2022ProgramID) {
		return false
	}

	if len(instr.Accounts) < 4 || len(instr.Data) < 9 {
		return false
	}

	if instr.Data[0] != 12 {
		return false
	}

	for i := 0; i < 4; i++ {
		if int(instr.Accounts[i]) >= len(p.allAccountKeys) {
			return false
		}
	}

	return true
}

func (p *Parser) isPumpFunTradeEventInstruction(inst solana.CompiledInstruction) bool {
	// Add bounds checking for ProgramIDIndex
	if int(inst.ProgramIDIndex) >= len(p.allAccountKeys) {
		p.Log.Warnf("ProgramIDIndex %d is out of range (allAccountKeys length: %d) in isPumpFunTradeEventInstruction", inst.ProgramIDIndex, len(p.allAccountKeys))
		return false
	}
	if !p.allAccountKeys[inst.ProgramIDIndex].Equals(PUMP_FUN_PROGRAM_ID) || len(inst.Data) < 16 {
		return false
	}
	decodedBytes, err := base58.Decode(inst.Data.String())
	if err != nil {
		return false
	}
	return bytes.Equal(decodedBytes[:16], PumpfunTradeEventDiscriminator[:])
}

func (p *Parser) isJupiterRouteEventInstruction(inst solana.CompiledInstruction) bool {
	// Add bounds checking for ProgramIDIndex
	if int(inst.ProgramIDIndex) >= len(p.allAccountKeys) {
		p.Log.Warnf("ProgramIDIndex %d is out of range (allAccountKeys length: %d) in isJupiterRouteEventInstruction", inst.ProgramIDIndex, len(p.allAccountKeys))
		return false
	}
	if !p.allAccountKeys[inst.ProgramIDIndex].Equals(JUPITER_PROGRAM_ID) || len(inst.Data) < 16 {
		return false
	}
	decodedBytes, err := base58.Decode(inst.Data.String())
	if err != nil {
		return false
	}
	return bytes.Equal(decodedBytes[:16], JupiterRouteEventDiscriminator[:])
}
