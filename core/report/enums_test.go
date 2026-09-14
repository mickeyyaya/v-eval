package report

import "testing"

func TestEnumValidators(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name string
		ok   bool
		fn   func() bool
	}{
		{"PASS", true, Result("PASS").IsValid},
		{"pass", false, Result("pass").IsValid},
		{"INCOMPLETE", true, Overall("INCOMPLETE").IsValid},
		{"ADVISORY", true, Overall("ADVISORY").IsValid},
		{"remote_sandbox", true, Isolation("remote_sandbox").IsValid},
		{"docker", false, Isolation("docker").IsValid},
		{"candidate_supplied", true, Origin("candidate_supplied").IsValid},
		{"rubric_judgment", true, Method("rubric_judgment").IsValid},
		{"not_checkable", true, ClaimStatus("not_checkable").IsValid},
		{"project_rubric", true, Authority("project_rubric").IsValid},
		{"detector", true, ObservationOrigin("detector").IsValid},
		{"", false, Kind("").IsValid},
		{"approved", true, ContractStatus("approved").IsValid},
		{"pending", false, ContractStatus("pending").IsValid},
		{"suspicious", true, Severity("suspicious").IsValid},
		{"critical", false, Severity("critical").IsValid},
		{"explained", true, Disposition("explained").IsValid},
		{"closed", false, Disposition("closed").IsValid},
	}
	for _, c := range cases {
		if c.fn() != c.ok {
			t.Errorf("%s: IsValid = %v, want %v", c.name, !c.ok, c.ok)
		}
	}
}
