package domain

import (
	"errors"
	"reflect"
	"testing"
	"time"
)

var kyiv = func() *time.Location {
	l, err := time.LoadLocation("Europe/Kyiv")
	if err != nil {
		panic(err)
	}
	return l
}()

func calDay(y int, m time.Month, d int) time.Time { return time.Date(y, m, d, 0, 0, 0, 0, time.UTC) }

func validFiling(kind FilingKind) Filing {
	return Filing{
		Kind:          kind,
		Edrpou:        "12345678",
		Name:          "ОСББ «Тестовий Будинок»",
		Address:       "м. Київ, вул. Тестова, 1",
		ApplicantName: "Тестовий Голова Правління",
		Position:      PositionHead,
		TermStart:     calDay(2026, 1, 1),
		Extract:       &ExtractFile{Name: "vytyah.pdf", Content: padded([]byte("%PDF-1.7\n"), 2048)},
	}
}

// filingDate is late in the evening in Kyiv: the calendar day in Kyiv is 10 November.
var filingDate = time.Date(2026, 11, 10, 23, 30, 0, 0, kyiv)

func TestFilingValidateNamesEveryBrokenField(t *testing.T) {
	long := func(n int) string {
		r := make([]rune, n)
		for i := range r {
			r[i] = 'а'
		}
		return string(r)
	}
	later := calDay(2026, 11, 11)
	tests := []struct {
		name   string
		mutate func(f *Filing)
		want   []string
	}{
		{"valid", func(f *Filing) {}, nil},
		{"edrpou of 7 digits", func(f *Filing) { f.Edrpou = "1234567" }, []string{"/edrpou"}},
		{"edrpou of 9 digits", func(f *Filing) { f.Edrpou = "123456789" }, []string{"/edrpou"}},
		{"edrpou with a letter", func(f *Filing) { f.Edrpou = "1234567a" }, []string{"/edrpou"}},
		{"edrpou empty", func(f *Filing) { f.Edrpou = "" }, []string{"/edrpou"}},
		{"name empty", func(f *Filing) { f.Name = "" }, []string{"/association_name"}},
		{"name over 200", func(f *Filing) { f.Name = long(201) }, []string{"/association_name"}},
		{"name of exactly 200", func(f *Filing) { f.Name = long(200) }, nil},
		{"address empty", func(f *Filing) { f.Address = "" }, []string{"/association_address"}},
		{"address over 500", func(f *Filing) { f.Address = long(501) }, []string{"/association_address"}},
		{"full name empty", func(f *Filing) { f.ApplicantName = "" }, []string{"/applicant_full_name"}},
		{"full name over 200", func(f *Filing) { f.ApplicantName = long(201) }, []string{"/applicant_full_name"}},
		{"position member", func(f *Filing) { f.Position = "member" }, []string{"/position"}},
		{"position empty", func(f *Filing) { f.Position = "" }, []string{"/position"}},
		{"term start missing", func(f *Filing) { f.TermStart = time.Time{} }, []string{"/term_start"}},
		{"term start after filing date", func(f *Filing) { f.TermStart = later }, []string{"/term_start"}},
		{"term start on filing date", func(f *Filing) { f.TermStart = calDay(2026, 11, 10) }, nil},
		{"no extract", func(f *Filing) { f.Extract = nil }, []string{"/extract"}},
		{"extract not a pdf", func(f *Filing) { f.Extract = &ExtractFile{Name: "a.txt", Content: []byte("text")} }, []string{"/extract"}},
		{"extract over 5 MB", func(f *Filing) {
			f.Extract = &ExtractFile{Name: "a.pdf", Content: padded([]byte("%PDF-1.7\n"), MaxExtractBytes+1)}
		}, []string{"/extract"}},
		{"everything broken at once", func(f *Filing) {
			f.Edrpou, f.Name, f.Address, f.ApplicantName = "12", "", "", ""
			f.Position = "member"
			f.TermStart = time.Time{}
			f.Extract = nil
		}, []string{"/edrpou", "/association_name", "/association_address", "/applicant_full_name", "/position", "/term_start", "/extract"}},
		{"two broken", func(f *Filing) { f.Edrpou = "1"; f.TermStart = later }, []string{"/edrpou", "/term_start"}},
	}
	for _, kind := range []FilingKind{FilingConnection, FilingTransfer} {
		for _, tc := range tests {
			t.Run(string(kind)+"/"+tc.name, func(t *testing.T) {
				f := validFiling(kind)
				tc.mutate(&f)
				err := f.Validate(filingDate)
				if tc.want == nil {
					if err != nil {
						t.Fatalf("Validate() = %v, want nil", err)
					}
					return
				}
				var ce *CodedFieldsError
				if !errors.As(err, &ce) {
					t.Fatalf("Validate() = %v, want *CodedFieldsError naming %v", err, tc.want)
				}
				if ce.Code != CodeApplicationInvalid {
					t.Errorf("code = %q, want %q", ce.Code, CodeApplicationInvalid)
				}
				if got := ce.Pointers(); !reflect.DeepEqual(got, tc.want) {
					t.Errorf("pointers = %v, want %v", got, tc.want)
				}
			})
		}
	}
}

func TestFilingValidateTermEnd(t *testing.T) {
	before, onDay, after := calDay(2026, 11, 9), calDay(2026, 11, 10), calDay(2027, 5, 1)
	tests := []struct {
		name  string
		end   *time.Time
		ended bool
	}{
		{"ended the day before the filing date", &before, true},
		{"ends on the filing date", &onDay, false},
		{"ends later", &after, false},
		{"no term end", nil, false},
	}
	for _, kind := range []FilingKind{FilingConnection, FilingTransfer} {
		for _, tc := range tests {
			t.Run(string(kind)+"/"+tc.name, func(t *testing.T) {
				f := validFiling(kind)
				f.TermEnd = tc.end
				err := f.Validate(filingDate)
				if !tc.ended {
					if err != nil {
						t.Fatalf("Validate() = %v, want nil", err)
					}
					return
				}
				var ce *CodedFieldsError
				if !errors.As(err, &ce) || ce.Code != CodeTermEnded {
					t.Fatalf("Validate() = %v, want code %q", err, CodeTermEnded)
				}
			})
		}
	}
}
