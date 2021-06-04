package viapi

import (
	"reflect"
	"regexp"
	"testing"
)

// testMatchError performs regex matching for error cases.
func testMatchError(t *testing.T, err error, r *regexp.Regexp) {
	switch {
	case err == nil:
		t.Fatal("expected error, got none")
	case !r.MatchString(err.Error()):
		t.Fatalf("expected error %q to match regexp %q", err.Error(), r)
	}
}

type testParseVersion struct {
	Name string

	product     string
	version     string
	build       string
	expected    VSphereVersion
	expectedErr *regexp.Regexp
}

func (tc *testParseVersion) Test(t *testing.T) {
	actual, err := parseVersion(tc.product, tc.version, tc.build)
	if err != nil && tc.expectedErr == nil {
		t.Fatalf("bad: %s", err)
	}
	if tc.expectedErr != nil {
		testMatchError(t, err, tc.expectedErr)
		return
	}
	if !reflect.DeepEqual(tc.expected, actual) {
		t.Fatalf("expected %#v, got %#v", tc.expected, actual)
	}
}

var testParseVersionExpected = VSphereVersion{
	Product: "VMware vCenter Server",
	Major:   6,
	Minor:   2,
	Patch:   1,
	Build:   1000000,
}

func TestParseVersion(t *testing.T) {
	cases := []testParseVersion{
		{
			Name:     "basic",
			product:  "VMware vCenter Server",
			version:  "6.2.1",
			build:    "1000000",
			expected: testParseVersionExpected,
		},
		{
			Name:        "bad major",
			product:     "VMware vCenter Server",
			version:     "6a.2.1",
			build:       "1000000",
			expectedErr: regexp.MustCompile("could not parse major version"),
		},
		{
			Name:        "bad minor",
			product:     "VMware vCenter Server",
			version:     "6.2a.1",
			build:       "1000000",
			expectedErr: regexp.MustCompile("could not parse minor version"),
		},
		{
			Name:        "bad patch",
			product:     "VMware vCenter Server",
			version:     "6.2.1a",
			build:       "1000000",
			expectedErr: regexp.MustCompile("could not parse patch version"),
		},
		{
			Name:        "bad build",
			product:     "VMware vCenter Server",
			version:     "6.2.1",
			build:       "1000000a",
			expectedErr: regexp.MustCompile("could not parse build version"),
		},
	}
	for _, tc := range cases {
		t.Run(tc.Name, tc.Test)
	}
}

type testCompareVersionExpectedResult string

const (
	testCompareVersionNewer   = testCompareVersionExpectedResult("newer")
	testCompareVersionOlder   = testCompareVersionExpectedResult("older")
	testCompareVersionEqual   = testCompareVersionExpectedResult("equal")
	testCompareVersionUnknown = testCompareVersionExpectedResult("unknown")
)

type testCompareVersion struct {
	Name string

	productA string
	versionA string
	buildA   string
	productB string
	versionB string
	buildB   string

	expected testCompareVersionExpectedResult
}

func (tc *testCompareVersion) Test(t *testing.T) {
	verA, err := parseVersion(tc.productA, tc.versionA, tc.buildA)
	if err != nil {
		t.Fatalf("bad: %s", err)
	}
	verB, err := parseVersion(tc.productB, tc.versionB, tc.buildB)
	if err != nil {
		t.Fatalf("bad: %s", err)
	}

	var actual []testCompareVersionExpectedResult
	if verA.Older(verB) {
		actual = append(actual, testCompareVersionOlder)
	}
	if verA.Newer(verB) {
		actual = append(actual, testCompareVersionNewer)
	}
	if verA.Equal(verB) {
		actual = append(actual, testCompareVersionEqual)
	}

	switch {
	case len(actual) == 0:
		if tc.expected == testCompareVersionUnknown {
			return
		}
		t.Fatalf("expected %s but result was unknown", tc.expected)
	case len(actual) > 1:
		t.Fatalf("expected only one result, got %s", actual)
	case tc.expected != actual[0]:
		t.Fatalf("expected %s, got %s", tc.expected, actual)
	}
}

func TestCompareVersion(t *testing.T) {
	cases := []testCompareVersion{
		{
			Name:     "equal",
			productA: "VMware vCenter Server",
			versionA: "6.2.1",
			buildA:   "1000000",
			productB: "VMware vCenter Server",
			versionB: "6.2.1",
			buildB:   "1000000",
			expected: testCompareVersionEqual,
		},
		{
			Name:     "unknown (different products)",
			productA: "VMware vCenter Server",
			versionA: "6.2.1",
			buildA:   "1000000",
			productB: "VMware ESXi",
			versionB: "6.2.1",
			buildB:   "1000000",
			expected: testCompareVersionUnknown,
		},
		{
			Name:     "newer (major)",
			productA: "VMware vCenter Server",
			versionA: "6.2.1",
			buildA:   "1000000",
			productB: "VMware vCenter Server",
			versionB: "5.2.1",
			buildB:   "1000000",
			expected: testCompareVersionNewer,
		}, {
			Name:     "newer (minor)",
			productA: "VMware vCenter Server",
			versionA: "6.2.1",
			buildA:   "1000000",
			productB: "VMware vCenter Server",
			versionB: "6.1.1",
			buildB:   "1000000",
			expected: testCompareVersionNewer,
		},
		{
			Name:     "newer (patch)",
			productA: "VMware vCenter Server",
			versionA: "6.2.1",
			buildA:   "1000000",
			productB: "VMware vCenter Server",
			versionB: "6.2.0",
			buildB:   "1000000",
			expected: testCompareVersionNewer,
		},
		{
			Name:     "newer (build)",
			productA: "VMware vCenter Server",
			versionA: "6.2.1",
			buildA:   "1000001",
			productB: "VMware vCenter Server",
			versionB: "6.2.1",
			buildB:   "1000000",
			expected: testCompareVersionNewer,
		},
		{
			Name:     "newer (higher build number but version number wins)",
			productA: "VMware vCenter Server",
			versionA: "6.2.2",
			buildA:   "1000000",
			productB: "VMware vCenter Server",
			versionB: "6.2.1",
			buildB:   "1000001",
			expected: testCompareVersionNewer,
		},
		{
			Name:     "newer (zero-value build number and minor)",
			productA: "VMware vCenter Server",
			versionA: "6.0.0",
			buildA:   "0",
			productB: "VMware vCenter Server",
			versionB: "5.5.0",
			buildB:   "1234567",
			expected: testCompareVersionNewer,
		},
		{
			Name:     "older (zero-value build number and minor)",
			productA: "VMware vCenter Server",
			versionA: "5.5.0",
			buildA:   "1234567",
			productB: "VMware vCenter Server",
			versionB: "6.0.0",
			buildB:   "0",
			expected: testCompareVersionOlder,
		},
		{
			Name:     "older (major)",
			productA: "VMware vCenter Server",
			versionA: "5.2.1",
			buildA:   "1000000",
			productB: "VMware vCenter Server",
			versionB: "6.2.1",
			buildB:   "1000000",
			expected: testCompareVersionOlder,
		}, {
			Name:     "older (minor)",
			productA: "VMware vCenter Server",
			versionA: "6.1.1",
			buildA:   "1000000",
			productB: "VMware vCenter Server",
			versionB: "6.2.1",
			buildB:   "1000000",
			expected: testCompareVersionOlder,
		},
		{
			Name:     "older (patch)",
			productA: "VMware vCenter Server",
			versionA: "6.2.0",
			buildA:   "1000000",
			productB: "VMware vCenter Server",
			versionB: "6.2.1",
			buildB:   "1000000",
			expected: testCompareVersionOlder,
		},
		{
			Name:     "older (build)",
			productA: "VMware vCenter Server",
			versionA: "6.2.1",
			buildA:   "1000000",
			productB: "VMware vCenter Server",
			versionB: "6.2.1",
			buildB:   "1000001",
			expected: testCompareVersionOlder,
		},
		{
			Name:     "older (higher build number but version number wins)",
			productA: "VMware vCenter Server",
			versionA: "6.2.1",
			buildA:   "1000001",
			productB: "VMware vCenter Server",
			versionB: "6.2.2",
			buildB:   "1000000",
			expected: testCompareVersionOlder,
		},
	}
	for _, tc := range cases {
		t.Run(tc.Name, tc.Test)
	}
}
