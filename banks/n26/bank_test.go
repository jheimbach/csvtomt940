package n26

import "testing"

func Test_extractAccountAndBankNumber(t *testing.T) {
	tests := []struct {
		name          string
		iban          string
		bankNumber    string
		accountNumber string
	}{
		{
			name:          "without spaces",
			iban:          "DE00111111110000000000",
			bankNumber:    "11111111",
			accountNumber: "0000000000",
		},
		{
			name:          "with spaces",
			iban:          "DE22 1111 1111 0000 0000 00",
			bankNumber:    "11111111",
			accountNumber: "0000000000",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bankNumber, accountNumber := extractAccountAndBankNumber(tt.iban)
			if bankNumber != tt.bankNumber {
				t.Errorf("extractAccountAndBankNumber() bankNumber = %v, accountNumber %v", bankNumber, tt.bankNumber)
			}
			if accountNumber != tt.accountNumber {
				t.Errorf("extractAccountAndBankNumber() accountNumber = %v, accountNumber %v", accountNumber, tt.accountNumber)
			}
		})
	}
}
