package domain

import shareddomain "github.com/mohsinazam/banking/internal/shared/domain"

// BalanceFromEntries derives an account balance from its append-only entry history.
// This is the core banking rule: balance is computed, never stored as source of truth.
func BalanceFromEntries(account Account, entries []Entry) (shareddomain.Money, error) {
	balance := shareddomain.ZeroSAR()

	for _, entry := range entries {
		if entry.AccountID() != account.ID() {
			continue
		}

		var err error
		switch {
		case account.IsDebitNormal() && entry.Side() == SideDebit:
			balance, err = balance.Add(entry.Amount())
		case account.IsDebitNormal() && entry.Side() == SideCredit:
			balance, err = balance.Sub(entry.Amount())
		case !account.IsDebitNormal() && entry.Side() == SideCredit:
			balance, err = balance.Add(entry.Amount())
		case !account.IsDebitNormal() && entry.Side() == SideDebit:
			balance, err = balance.Sub(entry.Amount())
		}
		if err != nil {
			return shareddomain.Money{}, err
		}
	}

	return balance, nil
}

// EntriesForAccount filters journal lines belonging to a single account.
func EntriesForAccount(accountID AccountID, entries []Entry) []Entry {
	var out []Entry
	for _, e := range entries {
		if e.AccountID() == accountID {
			out = append(out, e)
		}
	}
	return out
}

// VerifyEntriesBalanced checks that total debits equal total credits across a ledger slice.
// Used by reconciliation to prove the books still tie after many transactions.
func VerifyEntriesBalanced(entries []Entry) error {
	return validateBalanced(entries)
}
