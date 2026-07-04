package domain

import shareddomain "github.com/mohsinazam/banking/internal/shared/domain"

// ApplyEntries projects balance after applying additional journal lines to a starting balance.
func ApplyEntries(account Account, starting shareddomain.Money, entries []Entry) (shareddomain.Money, error) {
	balance := starting

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

// NetDebitExposure returns the total debit amount applied to a liability-style account
// (credits for asset-normal accounts) within the supplied entries.
func NetDebitExposure(account Account, entries []Entry) (shareddomain.Money, error) {
	total := shareddomain.ZeroSAR()
	for _, entry := range entries {
		if entry.AccountID() != account.ID() {
			continue
		}
		reducesBalance := (account.IsDebitNormal() && entry.Side() == SideCredit) ||
			(!account.IsDebitNormal() && entry.Side() == SideDebit)
		if !reducesBalance {
			continue
		}
		var err error
		total, err = total.Add(entry.Amount())
		if err != nil {
			return shareddomain.Money{}, err
		}
	}
	return total, nil
}
