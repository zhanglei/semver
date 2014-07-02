package semver

//TODO: Test incorrect version formats

import (
	"testing"
)

func prstr(s string) *PRVersion {
	return &PRVersion{s, 0, false}
}

func prnum(i uint64) *PRVersion {
	return &PRVersion{"", i, true}
}

type formatTest struct {
	v      Version
	result string
}

var formatTests = []formatTest{
	{Version{1, 2, 3, nil, nil}, "1.2.3"},
	{Version{0, 0, 1, nil, nil}, "0.0.1"},
	{Version{0, 0, 1, []*PRVersion{prstr("alpha"), prstr("preview")}, []string{"123", "456"}}, "0.0.1-alpha.preview+123.456"},
	{Version{1, 2, 3, []*PRVersion{prstr("alpha"), prnum(1)}, []string{"123", "456"}}, "1.2.3-alpha.1+123.456"},
	{Version{1, 2, 3, []*PRVersion{prstr("alpha"), prnum(1)}, nil}, "1.2.3-alpha.1"},
	{Version{1, 2, 3, nil, []string{"123", "456"}}, "1.2.3+123.456"},
	// Prereleases and build metadata hyphens
	{Version{1, 2, 3, []*PRVersion{prstr("alpha"), prstr("b-eta")}, []string{"123", "b-uild"}}, "1.2.3-alpha.b-eta+123.b-uild"},
	{Version{1, 2, 3, nil, []string{"123", "b-uild"}}, "1.2.3+123.b-uild"},
	{Version{1, 2, 3, []*PRVersion{prstr("alpha"), prstr("b-eta")}, nil}, "1.2.3-alpha.b-eta"},
}

func TestStringer(t *testing.T) {
	for _, test := range formatTests {
		if res := test.v.String(); res != test.result {
			t.Errorf("Stringer, expected %q but got %q", test.result, res)
		}
	}
}

func TestParse(t *testing.T) {
	for _, test := range formatTests {
		if v, err := Parse(test.result); err != nil {
			t.Errorf("Error parsing %q: %q", test.result, err)
		} else if comp := v.Compare(&test.v); comp != 0 {
			t.Errorf("Parsing, expected %q but got %q, comp: %d ", test.v, v, comp)
		} else if err := v.Validate(); err != nil {
			t.Errorf("Error validating parsed version %q: %q", test.v, err)
		}
	}
}

func TestValidate(t *testing.T) {
	for _, test := range formatTests {
		if err := test.v.Validate(); err != nil {
			t.Errorf("Error validating %q: %q", test.v, err)
		}
	}
}

type compareTest struct {
	v1     Version
	v2     Version
	result int
}

var compareTests = []compareTest{
	{Version{1, 0, 0, nil, nil}, Version{1, 0, 0, nil, nil}, 0},
	{Version{2, 0, 0, nil, nil}, Version{1, 0, 0, nil, nil}, 1},
	{Version{0, 1, 0, nil, nil}, Version{0, 1, 0, nil, nil}, 0},
	{Version{0, 2, 0, nil, nil}, Version{0, 1, 0, nil, nil}, 1},
	{Version{0, 0, 1, nil, nil}, Version{0, 0, 1, nil, nil}, 0},
	{Version{0, 0, 2, nil, nil}, Version{0, 0, 1, nil, nil}, 1},
	{Version{1, 2, 3, nil, nil}, Version{1, 2, 3, nil, nil}, 0},
	{Version{2, 2, 4, nil, nil}, Version{1, 2, 4, nil, nil}, 1},
	{Version{1, 3, 3, nil, nil}, Version{1, 2, 3, nil, nil}, 1},
	{Version{1, 2, 4, nil, nil}, Version{1, 2, 3, nil, nil}, 1},

	// Spec Examples #11
	{Version{1, 0, 0, nil, nil}, Version{2, 0, 0, nil, nil}, -1},
	{Version{2, 0, 0, nil, nil}, Version{2, 1, 0, nil, nil}, -1},
	{Version{2, 1, 0, nil, nil}, Version{2, 1, 1, nil, nil}, -1},

	{Version{1, 0, 0, []*PRVersion{prstr("alpha")}, nil}, Version{1, 0, 0, []*PRVersion{prstr("alpha"), prnum(1)}, nil}, -1},
	{Version{1, 0, 0, []*PRVersion{prstr("alpha"), prnum(1)}, nil}, Version{1, 0, 0, []*PRVersion{prstr("alpha"), prstr("beta")}, nil}, -1},
	{Version{1, 0, 0, []*PRVersion{prstr("alpha"), prstr("beta")}, nil}, Version{1, 0, 0, []*PRVersion{prstr("beta")}, nil}, -1},
	{Version{1, 0, 0, []*PRVersion{prstr("beta")}, nil}, Version{1, 0, 0, []*PRVersion{prstr("beta"), prnum(2)}, nil}, -1},
	{Version{1, 0, 0, []*PRVersion{prstr("beta"), prnum(2)}, nil}, Version{1, 0, 0, []*PRVersion{prstr("beta"), prnum(11)}, nil}, -1},
	{Version{1, 0, 0, []*PRVersion{prstr("beta"), prnum(11)}, nil}, Version{1, 0, 0, []*PRVersion{prstr("rc"), prnum(1)}, nil}, -1},
	{Version{1, 0, 0, []*PRVersion{prstr("rc"), prnum(1)}, nil}, Version{1, 0, 0, nil, nil}, -1},

	// Ignore Build metadata
	{Version{1, 0, 0, nil, []string{"1", "2", "3"}}, Version{1, 0, 0, nil, nil}, 0},
}

func TestCompare(t *testing.T) {
	for _, test := range compareTests {
		if res := test.v1.Compare(&test.v2); res != test.result {
			t.Errorf("Comparing %q : %q, expected %d but got %d", test.v1, test.v2, test.result, res)
		}
		//Test counterpart
		if res := test.v2.Compare(&test.v1); res != -test.result {
			t.Errorf("Comparing %q : %q, expected %d but got %d", test.v2, test.v1, -test.result, res)
		}
	}
}

type wrongformatTest struct {
	v   *Version
	str string
}

var wrongformatTests = []wrongformatTest{
	{nil, ""},
	{nil, "."},
	{nil, "1."},
	{nil, ".1"},
	{nil, "a.b.c"},
	{nil, "1.a.b"},
	{nil, "1.1.a"},
	{nil, "1.a.1"},
	{nil, "a.1.1"},
	{nil, ".."},
	{nil, "1.."},
	{nil, "1.1."},
	{nil, "1..1"},
	{nil, "1.1.+123"},
	{nil, "1.1.-beta"},
	{nil, "-1.1.1"},
	{nil, "1.-1.1"},
	{nil, "1.1.-1"},
	{&Version{0, 0, 0, []*PRVersion{prstr("!")}, nil}, "0.0.0-!"},
	{&Version{0, 0, 0, nil, []string{"!"}}, "0.0.0+!"},
	// empty prversion
	{&Version{0, 0, 0, []*PRVersion{prstr(""), prstr("alpha")}, nil}, "0.0.0-.alpha"},
	// empty build meta data
	{&Version{0, 0, 0, []*PRVersion{prstr("alpha")}, []string{""}}, "0.0.0-alpha+"},
	{&Version{0, 0, 0, []*PRVersion{prstr("alpha")}, []string{"test", ""}}, "0.0.0-alpha+test."},
}

func TestWrongFormat(t *testing.T) {
	for _, test := range wrongformatTests {

		if res, err := Parse(test.str); err == nil {
			t.Errorf("Parsing wrong format version %q, expected error but got %q", test.str, res)
		}

		if test.v != nil {
			if err := test.v.Validate(); err == nil {
				t.Errorf("Validating wrong format version %q (%q), expected error", test.v, test.str)
			}
		}
	}
}

func TestCompareHelper(t *testing.T) {
	v := &Version{1, 0, 0, []*PRVersion{prstr("alpha")}, nil}
	v1 := &Version{1, 0, 0, nil, nil}
	if !v.GTE(v) {
		t.Errorf("%q should be greater than or equal to %q", v, v)
	}
	if !v.LTE(v) {
		t.Errorf("%q should be greater than or equal to %q", v, v)
	}
	if !v.LT(v1) {
		t.Errorf("%q should be less than %q", v, v1)
	}
	if !v.LTE(v1) {
		t.Errorf("%q should be less than or equal %q", v, v1)
	}
	if !v1.GT(v) {
		t.Errorf("%q should be less than %q", v1, v)
	}
	if !v1.GTE(v) {
		t.Errorf("%q should be less than or equal %q", v1, v)
	}
}

func TestPreReleaseVersions(t *testing.T) {
	p1, err := NewPRVersion("123")
	if !p1.IsNumeric() {
		t.Errorf("Expected numeric prversion, got %q", p1)
	}
	if p1.VersionNum != 123 {
		t.Error("Wrong prversion number")
	}
	if err != nil {
		t.Errorf("Not expected error %q", err)
	}
	p2, err := NewPRVersion("alpha")
	if p2.IsNumeric() {
		t.Errorf("Expected non-numeric prversion, got %q", p2)
	}
	if p2.VersionStr != "alpha" {
		t.Error("Wrong prversion string")
	}
	if err != nil {
		t.Errorf("Not expected error %q", err)
	}
}
