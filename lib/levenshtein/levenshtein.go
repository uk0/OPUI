package levenshtein

import "unicode"

func minimum(a, b, c int) int {
	if a < b && a < c {
		return a
	}
	if b < c && b < a {
		return b
	}
	return c
}

func runes(s string) (a []rune) {
	for _, ch := range s {
		a = append(a, ch)
	}
	return a
}

// DistanceFunc calc the strings distance between sstr and tstr,
// use function to determine if characters is equal.
func DistanceFunc(sstr, tstr string, equal func(a, b rune) bool) int {
	if sstr == tstr {
		return 0
	}

	// code from wikipedia
	s, t := runes(sstr), runes(tstr)
	sn := len(s)
	tn := len(t)

	if sn == 0 {
		return tn
	}

	if tn == 0 {
		return sn
	}

	sz := tn + 1
	v0 := make([]int, sz)
	v1 := make([]int, sz)

	for i := 0; i < sz; i++ {
		v0[i] = i
	}

	for i := 0; i < sn; i++ {
		v1[0] = i + 1

		for j := 0; j < tn; j++ {
			var cost int
			if !equal(s[i], t[j]) {
				cost = 1
			}
			v1[j+1] = minimum(v1[j]+1, v0[j+1]+1, v0[j]+cost)
		}

		for j := 0; j < sz; j++ {
			v0[j] = v1[j]
		}
	}

	return v1[tn]
}

// Distance calc the strings distance between sstr and tstr
func Distance(sstr, tstr string) int {
	return DistanceFunc(sstr, tstr, func(a, b rune) bool {
		return a == b
	})
}

// DistanceCI calc the strings distance between sstr and tstr, case insensitive.
func DistanceCI(sstr, tstr string) int {
	return DistanceFunc(sstr, tstr, func(a, b rune) bool {
		return unicode.ToLower(a) == unicode.ToLower(b)
	})
}
