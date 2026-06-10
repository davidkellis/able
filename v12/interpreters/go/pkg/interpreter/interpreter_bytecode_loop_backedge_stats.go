package interpreter

import "sort"

const bytecodeLoopBackedgeBalanceSiteLimit = 512

type bytecodeLoopBackedgeBalanceKey struct {
	Site    bytecodeStackPeakSiteKey
	Program *bytecodeProgram
}

type bytecodeLoopBackedgeBalance struct {
	Backedges uint64
	Baseline  int
	Excess    uint64
	MaxExcess uint64
}

// BytecodeLoopBackedgeBalanceSnapshot reports stack depth retained across a
// repeated unconditional backward jump. Baseline is the first observed target
// depth, while Excess and MaxExcess cover later observations above it.
type BytecodeLoopBackedgeBalanceSnapshot struct {
	Op        int    `json:"op"`
	IP        int    `json:"ip"`
	Name      string `json:"name"`
	Origin    string `json:"origin,omitempty"`
	Line      int    `json:"line,omitempty"`
	Column    int    `json:"column,omitempty"`
	Backedges uint64 `json:"backedges"`
	Baseline  int    `json:"baseline"`
	Excess    uint64 `json:"excess"`
	MaxExcess uint64 `json:"max_excess"`
}

func (i *Interpreter) recordBytecodeLoopBackedgeBalance(site bytecodeStackPeakSiteKey, program *bytecodeProgram, stackDepth int) {
	if i == nil || !i.bytecodeStatsEnabled || stackDepth < 0 {
		return
	}
	key := bytecodeLoopBackedgeBalanceKey{Site: site, Program: program}
	i.bytecodeLoopBackedgeBalancesMu.Lock()
	defer i.bytecodeLoopBackedgeBalancesMu.Unlock()
	if i.bytecodeLoopBackedgeBalances == nil {
		i.bytecodeLoopBackedgeBalances = make(map[bytecodeLoopBackedgeBalanceKey]bytecodeLoopBackedgeBalance, 16)
	}
	balance, ok := i.bytecodeLoopBackedgeBalances[key]
	if !ok {
		if len(i.bytecodeLoopBackedgeBalances) >= bytecodeLoopBackedgeBalanceSiteLimit {
			return
		}
		i.bytecodeLoopBackedgeBalances[key] = bytecodeLoopBackedgeBalance{Backedges: 1, Baseline: stackDepth}
		return
	}
	balance.Backedges++
	if stackDepth > balance.Baseline {
		excess := uint64(stackDepth - balance.Baseline)
		balance.Excess += excess
		if excess > balance.MaxExcess {
			balance.MaxExcess = excess
		}
	}
	i.bytecodeLoopBackedgeBalances[key] = balance
}

func (i *Interpreter) bytecodeLoopBackedgeBalanceSnapshot() []BytecodeLoopBackedgeBalanceSnapshot {
	if i == nil {
		return nil
	}
	i.bytecodeLoopBackedgeBalancesMu.Lock()
	defer i.bytecodeLoopBackedgeBalancesMu.Unlock()
	if len(i.bytecodeLoopBackedgeBalances) == 0 {
		return nil
	}
	balances := make([]BytecodeLoopBackedgeBalanceSnapshot, 0, len(i.bytecodeLoopBackedgeBalances))
	for key, balance := range i.bytecodeLoopBackedgeBalances {
		if balance.Excess == 0 {
			continue
		}
		site := key.Site
		balances = append(balances, BytecodeLoopBackedgeBalanceSnapshot{
			Op:        site.Op,
			IP:        site.IP,
			Name:      site.Name,
			Origin:    site.Origin,
			Line:      site.Line,
			Column:    site.Column,
			Backedges: balance.Backedges,
			Baseline:  balance.Baseline,
			Excess:    balance.Excess,
			MaxExcess: balance.MaxExcess,
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
