package interpreter

import "sort"

const bytecodeCallOperandBalanceSiteLimit = 256

type bytecodeCallOperandBalance struct {
	Violations      uint64
	OperandValues   uint64
	ExpectedResults uint64
	ActualResults   uint64
	Excess          uint64
	MaxExcess       uint64
}

// BytecodeCallOperandBalanceSnapshot reports calls that left more values than
// their receiver/arguments were expected to produce. The operand and result
// counts are aggregated only across those imbalanced completions.
type BytecodeCallOperandBalanceSnapshot struct {
	Op              int    `json:"op"`
	IP              int    `json:"ip"`
	Name            string `json:"name"`
	Origin          string `json:"origin,omitempty"`
	Line            int    `json:"line,omitempty"`
	Column          int    `json:"column,omitempty"`
	Violations      uint64 `json:"violations"`
	OperandValues   uint64 `json:"operand_values"`
	ExpectedResults uint64 `json:"expected_results"`
	ActualResults   uint64 `json:"actual_results"`
	Excess          uint64 `json:"excess"`
	MaxExcess       uint64 `json:"max_excess"`
}

func (i *Interpreter) recordBytecodeCallOperandBalance(site bytecodeStackPeakSiteKey, operandValues int, expectedResults int, actualResults int) {
	if i == nil || !i.bytecodeStatsEnabled || operandValues < 0 || expectedResults < 0 || actualResults <= expectedResults {
		return
	}
	excess := uint64(actualResults - expectedResults)
	balance := bytecodeCallOperandBalance{
		Violations:      1,
		OperandValues:   uint64(operandValues),
		ExpectedResults: uint64(expectedResults),
		ActualResults:   uint64(actualResults),
		Excess:          excess,
		MaxExcess:       excess,
	}
	i.bytecodeCallOperandBalancesMu.Lock()
	defer i.bytecodeCallOperandBalancesMu.Unlock()
	if i.bytecodeCallOperandBalances == nil {
		i.bytecodeCallOperandBalances = make(map[bytecodeStackPeakSiteKey]bytecodeCallOperandBalance, 16)
	}
	if current, ok := i.bytecodeCallOperandBalances[site]; ok {
		current.Violations += balance.Violations
		current.OperandValues += balance.OperandValues
		current.ExpectedResults += balance.ExpectedResults
		current.ActualResults += balance.ActualResults
		current.Excess += balance.Excess
		if balance.MaxExcess > current.MaxExcess {
			current.MaxExcess = balance.MaxExcess
		}
		i.bytecodeCallOperandBalances[site] = current
		return
	}
	if len(i.bytecodeCallOperandBalances) < bytecodeCallOperandBalanceSiteLimit {
		i.bytecodeCallOperandBalances[site] = balance
		return
	}
	var (
		leastSite    bytecodeStackPeakSiteKey
		leastBalance bytecodeCallOperandBalance
		haveLeast    bool
	)
	for candidate, candidateBalance := range i.bytecodeCallOperandBalances {
		if !haveLeast || candidateBalance.Excess < leastBalance.Excess {
			leastSite = candidate
			leastBalance = candidateBalance
			haveLeast = true
		}
	}
	if haveLeast && balance.Excess >= leastBalance.Excess {
		delete(i.bytecodeCallOperandBalances, leastSite)
		i.bytecodeCallOperandBalances[site] = balance
	}
}

func (i *Interpreter) bytecodeCallOperandBalanceSnapshot() []BytecodeCallOperandBalanceSnapshot {
	if i == nil {
		return nil
	}
	i.bytecodeCallOperandBalancesMu.Lock()
	defer i.bytecodeCallOperandBalancesMu.Unlock()
	if len(i.bytecodeCallOperandBalances) == 0 {
		return nil
	}
	balances := make([]BytecodeCallOperandBalanceSnapshot, 0, len(i.bytecodeCallOperandBalances))
	for site, balance := range i.bytecodeCallOperandBalances {
		balances = append(balances, BytecodeCallOperandBalanceSnapshot{
			Op:              site.Op,
			IP:              site.IP,
			Name:            site.Name,
			Origin:          site.Origin,
			Line:            site.Line,
			Column:          site.Column,
			Violations:      balance.Violations,
			OperandValues:   balance.OperandValues,
			ExpectedResults: balance.ExpectedResults,
			ActualResults:   balance.ActualResults,
			Excess:          balance.Excess,
			MaxExcess:       balance.MaxExcess,
		})
	}
	sort.Slice(balances, func(left, right int) bool {
		if balances[left].Excess != balances[right].Excess {
			return balances[left].Excess > balances[right].Excess
		}
		if balances[left].MaxExcess != balances[right].MaxExcess {
			return balances[left].MaxExcess > balances[right].MaxExcess
		}
		if balances[left].Origin != balances[right].Origin {
			return balances[left].Origin < balances[right].Origin
		}
		if balances[left].Line != balances[right].Line {
			return balances[left].Line < balances[right].Line
		}
		if balances[left].Column != balances[right].Column {
			return balances[left].Column < balances[right].Column
		}
		if balances[left].Op != balances[right].Op {
			return balances[left].Op < balances[right].Op
		}
		return balances[left].IP < balances[right].IP
	})
	return balances
}
