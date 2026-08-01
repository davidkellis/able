package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestJoinNominalEffectsLinksInterfaceAndIndirectSites(t *testing.T) {
	path := filepath.Join(t.TempDir(), "effects.json")
	data := []byte(`{
  "callables": [
    {
      "callable": "sample.impl Visitor for Reader.inspect",
      "package": "sample",
      "kind": "method",
      "generated_go_name": "impl_Visitor_inspect_0",
      "nominal_parameters": [{
        "index": 2,
        "name": "state",
        "nominal": "State",
        "read_only_non_escaping": true
      }]
    },
    {
      "callable": "sample.main::lambda@1",
      "package": "sample",
      "kind": "lambda",
      "nominal_parameters": [{
        "index": 0,
        "name": "record",
        "nominal": "Record",
        "read_only_non_escaping": true
      }]
    }
  ]
}`)
	if err := os.WriteFile(path, data, 0o600); err != nil {
		t.Fatal(err)
	}
	proofs := map[string]*nominalProof{
		"State": {
			UnknownCallSites: []nominalUnknownCallSite{{
				Caller:          "__able_compiled_fn_walk_ctx",
				Callee:          "visitor.__able_ctx_inspect",
				ArgumentIndexes: []int{1},
			}},
		},
		"Record": {
			UnknownCallSites: []nominalUnknownCallSite{{
				Caller:          "__able_compiled_method_Result_map_spec",
				Callee:          "__able_tmp_5",
				ArgumentIndexes: []int{0},
			}},
		},
	}
	links, err := joinNominalEffects(path, proofs)
	if err != nil {
		t.Fatal(err)
	}
	if got := links["State"][0]; got.Resolution != "interface-method-candidate-set" ||
		len(got.Candidates) != 1 ||
		!got.Candidates[0].ReadOnlyNonEscaping {
		t.Fatalf("State link = %#v", got)
	}
	if got := links["Record"][0]; got.Resolution != "indirect-callable-candidate-set" ||
		len(got.Candidates) != 1 ||
		!got.Candidates[0].ReadOnlyNonEscaping {
		t.Fatalf("Record link = %#v", got)
	}
}
