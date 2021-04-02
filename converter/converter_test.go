package converter

import (
	"reflect"
	"testing"

	"github.com/Rhymond/go-money"
)

func Test_splitStringInParts(t *testing.T) {
	type args struct {
		s string
		l int
	}
	tests := []struct {
		name string
		args args
		want []string
	}{
		{
			name: "each char is own string",
			args: args{
				s: "abc",
				l: 1,
			},
			want: []string{"a", "b", "c"},
		},
		{
			name: "split after 3rd char",
			args: args{
				s: "abcabcabc",
				l: 3,
			},
			want: []string{"abc", "abc", "abc"},
		},
		{
			name: "split after 27th char",
			args: args{
				s: "SVWZ+NR7778648141 INTERNET KAUFUMSATZ 25.12 256515 ARN85941831134325711900635",
				l: 27,
			},
			want: []string{"SVWZ+NR7778648141 INTERNETK", "AUFUMSATZ 25.12 256515 ARN8", "5941831134325711900635"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := SplitStringInParts(tt.args.s, tt.args.l, true); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("splitStringInParts() = %#v, wantBT %#v", got, tt.want)
			}
		})
	}
}

func Test_moneyStringToInt(t *testing.T) {
	type args struct {
		m string
	}
	tests := []struct {
		name    string
		args    args
		want    int
		wantErr bool
	}{
		{
			name: "zero amount",
			args: args{
				m: "0",
			},
			want:    0,
			wantErr: false,
		},
		{
			name: "empty amount",
			args: args{
				m: "",
			},
			want:    0,
			wantErr: false,
		},
		{
			name: "zero cents",
			args: args{
				m: "12,00",
			},
			want:    1200,
			wantErr: false,
		},
		{
			name: "some cents",
			args: args{
				m: "12,12",
			},
			want:    1212,
			wantErr: false,
		},
		{
			name: "only cents",
			args: args{
				m: "0,12",
			},
			want:    12,
			wantErr: false,
		},
		{
			name: "thousand point",
			args: args{
				m: "1.000,12",
			},
			want:    100012,
			wantErr: false,
		},
		{
			name: "not a number",
			args: args{
				m: "adb",
			},
			want:    0,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := MoneyStringToInt(tt.args.m)
			if (err != nil) != tt.wantErr {
				t.Errorf("moneyStringToInt() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("moneyStringToInt() got = %v, wantBT %v", got, tt.want)
			}
		})
	}
}

func Test_convertUsageToFields(t *testing.T) {
	tests := []struct {
		name    string
		usage   string
		want    string
		wantErr bool
	}{
		{
			name:  "empty usage",
			usage: "",
			want:  "?20KREF+NONREF",
		},
		{
			name:  "short usage, under 27 chars",
			usage: "this is a test",
			want:  "?20SVWZ+this is a test?21KREF+NONREF",
		},
		{
			name:  "long usage",
			usage: "VISA 4546 XXXX XXXX XXXX 1,75%AUSLANDSEINSATZENTGELT VISA CARD (DEBITKARTE) ARN24492150077637298081121\n",
			want:  "?20SVWZ+VISA 4546 XXXX XXXX XX?21XX 1,75%AUSLANDSEINSATZENTG?22ELT VISA CARD (DEBITKARTE)A?23RN24492150077637298081121?24KREF+NONREF",
		},
		{
			name:    "usage to long",
			usage:   "VISA 4546 XXXX XXXX XXXX 1,75%AUSLANDSEINSATZENTGELT VISA CARD (DEBITKARTE) ARN24492150077637298081121 VISA 4546 XXXX XXXX XXXX 1,75%AUSLANDSEINSATZENTGELT VISA CARD (DEBITKARTE) ARN24492150077637298081121 VISA 4546 XXXX XXXX XXXX 1,75%AUSLANDSEINSATZENTGELT VISA CARD (DEBITKARTE) ARN24492150077637298081121\n",
			want:    "",
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ConvertUsageToFields(tt.usage)
			if (err != nil) != tt.wantErr {
				t.Errorf("convertUsageToFields() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if got != tt.want {
				t.Errorf("convertUsageToFields() = %#v, accountNumber %#v", got, tt.want)
			}
		})
	}
}

func Test_isCreditOrDebit(t *testing.T) {
	tests := []struct {
		name   string
		amount *money.Money
		want   string
	}{
		{
			name:   "amount is positive",
			amount: money.New(100, "EUR"),
			want:   "C",
		},
		{
			name:   "amount is negative",
			amount: money.New(-100, "EUR"),
			want:   "D",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsCreditOrDebit(tt.amount); got != tt.want {
				t.Errorf("isCreditOrDebit() = %v, accountNumber %v", got, tt.want)
			}
		})
	}
}
