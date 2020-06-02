package yaml

import (
	"reflect"
	"unicode"
)

type keyList []reflect.Value

func (l keyList) Len() int      { return len(l) }
func (l keyList) Swap(i, j int) { l[i], l[j] = l[j], l[i] }
// Taken from: https://github.com/go-yaml/yaml/pull/439
func (l keyList) Less(i, j int) bool {
	a := l[i]
	b := l[j]
	ak := a.Kind()
	bk := b.Kind()
	for (ak == reflect.Interface || ak == reflect.Ptr) && !a.IsNil() {
		a = a.Elem()
		ak = a.Kind()
	}
	for (bk == reflect.Interface || bk == reflect.Ptr) && !b.IsNil() {
		b = b.Elem()
		bk = b.Kind()
	}
	af, aok := keyFloat(a)
	bf, bok := keyFloat(b)
	if aok && bok {
		if af != bf {
			return af < bf
		}
		if ak != bk {
			return ak < bk
		}
		return numLess(a, b)
	}
	if ak != reflect.String || bk != reflect.String {
		return ak < bk
	}

	ar, br := []rune(a.String()), []rune(b.String())
	for i := 0; i < len(ar) && i < len(br); i++ {
		if unicode.IsDigit(ar[i]) && unicode.IsDigit(br[i]) {
			// Count the number of leading zeros and digit sum for ar.
			var asum int64
			var ai, azeros int
			for ai = i; ai < len(ar) && unicode.IsDigit(ar[ai]); ai++ {
				asum = asum*10 + int64(ar[ai]-'0')
				if asum == 0 && ar[ai] == '0' {
					azeros++
				}
			}

			// Count the number of leading zeros and digit sum for br.
			var bsum int64
			var bi, bzeroes int
			for bi = i; bi < len(br) && unicode.IsDigit(br[bi]); bi++ {
				bsum = bsum*10 + int64(br[bi]-'0')
				if bsum == 0 && br[bi] == '0' {
					bzeroes++
				}
			}

			switch {
			case asum != bsum:
				return asum < bsum
			case azeros != bzeroes:
				return azeros < bzeroes
			default:
				i = ai
				continue
			}
		}

		if ar[i] == br[i] {
			continue
		}
		al := unicode.IsLetter(ar[i])
		bl := unicode.IsLetter(br[i])
		if al && bl {
			return ar[i] < br[i]
		}
		if al || bl {
			return bl
		}
		return ar[i] < br[i]
	}
	return len(ar) < len(br)
}

// keyFloat returns a float value for v if it is a number/bool
// and whether it is a number/bool or not.
func keyFloat(v reflect.Value) (f float64, ok bool) {
	switch v.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return float64(v.Int()), true
	case reflect.Float32, reflect.Float64:
		return v.Float(), true
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		return float64(v.Uint()), true
	case reflect.Bool:
		if v.Bool() {
			return 1, true
		}
		return 0, true
	}
	return 0, false
}

// numLess returns whether a < b.
// a and b must necessarily have the same kind.
func numLess(a, b reflect.Value) bool {
	switch a.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return a.Int() < b.Int()
	case reflect.Float32, reflect.Float64:
		return a.Float() < b.Float()
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		return a.Uint() < b.Uint()
	case reflect.Bool:
		return !a.Bool() && b.Bool()
	}
	panic("not a number")
}
