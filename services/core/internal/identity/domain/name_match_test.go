package domain

import "testing"

func TestSamePerson(t *testing.T) {
	tests := []struct {
		name string
		a, b string
		want bool
	}{
		{"identical", "Петренко Олена Іванівна", "Петренко Олена Іванівна", true},
		{"letter case", "ПЕТРЕНКО олена іванівна", "Петренко Олена Іванівна", true},
		{"extra spaces", "  Петренко   Олена  Іванівна ", "Петренко Олена Іванівна", true},
		{"apostrophe ascii vs right quote", "Мар'яна Дзюба Петрівна", "Мар’яна Дзюба Петрівна", true},
		{"apostrophe modifier letter", "Мар'яна Дзюба Петрівна", "Марʼяна Дзюба Петрівна", true},
		{"word order", "Олена Іванівна Петренко", "Петренко Олена Іванівна", true},
		{"all levelling at once", " ОЛЕНА  Іванівна ПЕТРЕНКО", "Петренко олена іванівна", true},
		{"another surname", "Петренко Олена Іванівна", "Шевченко Олена Іванівна", false},
		{"missing patronymic", "Петренко Олена", "Петренко Олена Іванівна", false},
		{"latin for cyrillic", "Petrenko Olena Ivanivna", "Петренко Олена Іванівна", false},
		{"latin lookalike letters", "Петренкo Олена Іванівна", "Петренко Олена Іванівна", false},
		{"empty vs name", "", "Петренко Олена Іванівна", false},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if got := SamePerson(tc.a, tc.b); got != tc.want {
				t.Errorf("SamePerson(%q, %q) = %v, want %v", tc.a, tc.b, got, tc.want)
			}
			if got := SamePerson(tc.b, tc.a); got != tc.want {
				t.Errorf("SamePerson(%q, %q) = %v, want %v (symmetry)", tc.b, tc.a, got, tc.want)
			}
		})
	}
}
