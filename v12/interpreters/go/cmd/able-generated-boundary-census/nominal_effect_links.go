package main

import (
	"encoding/json"
	"fmt"
	"os"
	"sort"
	"strings"
)

type nominalEffectInput struct {
	Callables []nominalEffectCallableInput `json:"callables"`
}

type nominalEffectCallableInput struct {
	Callable        string                        `json:"callable"`
	Package         string                        `json:"package"`
	Kind            string                        `json:"kind"`
	GeneratedGoName string                        `json:"generated_go_name"`
	Parameters      []nominalEffectParameterInput `json:"nominal_parameters"`
}

type nominalEffectParameterInput struct {
	Index               int      `json:"index"`
	Name                string   `json:"name"`
	Nominal             string   `json:"nominal"`
	ReadOnlyNonEscaping bool     `json:"read_only_non_escaping"`
	Blockers            []string `json:"blockers"`
}

type nominalEffectLink struct {
	Site       nominalUnknownCallSite       `json:"generated_call_site"`
	Resolution string                       `json:"resolution"`
	Candidates []nominalEffectLinkCandidate `json:"typed_callable_candidates"`
}

type nominalEffectLinkCandidate struct {
	Callable            string   `json:"callable"`
	Package             string   `json:"package"`
	Kind                string   `json:"kind"`
	GeneratedGoName     string   `json:"generated_go_name,omitempty"`
	ParameterIndex      int      `json:"parameter_index"`
	ParameterName       string   `json:"parameter_name"`
	ReadOnlyNonEscaping bool     `json:"read_only_non_escaping"`
	Blockers            []string `json:"blockers,omitempty"`
}

func joinNominalEffects(
	path string,
	proofs map[string]*nominalProof,
) (map[string][]nominalEffectLink, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var input nominalEffectInput
	if err := json.Unmarshal(data, &input); err != nil {
		return nil, fmt.Errorf("decode nominal effects: %w", err)
	}
	result := make(map[string][]nominalEffectLink)
	for nominal, proof := range proofs {
		for _, site := range proof.UnknownCallSites {
			link := nominalEffectLink{
				Site:       site,
				Resolution: "unresolved",
			}
			for _, callable := range input.Callables {
				for _, parameter := range callable.Parameters {
					if parameter.Nominal != nominal {
						continue
					}
					resolution, match := nominalEffectSiteMatch(
						site,
						callable,
						parameter,
					)
					if !match {
						continue
					}
					if nominalEffectResolutionRank(resolution) >
						nominalEffectResolutionRank(link.Resolution) {
						link.Resolution = resolution
						link.Candidates = nil
					}
					if resolution != link.Resolution {
						continue
					}
					link.Candidates = append(
						link.Candidates,
						nominalEffectLinkCandidate{
							Callable:            callable.Callable,
							Package:             callable.Package,
							Kind:                callable.Kind,
							GeneratedGoName:     callable.GeneratedGoName,
							ParameterIndex:      parameter.Index,
							ParameterName:       parameter.Name,
							ReadOnlyNonEscaping: parameter.ReadOnlyNonEscaping,
							Blockers:            parameter.Blockers,
						},
					)
				}
			}
			sort.Slice(link.Candidates, func(i, j int) bool {
				return link.Candidates[i].Callable < link.Candidates[j].Callable
			})
			result[nominal] = append(result[nominal], link)
		}
	}
	return result, nil
}

func nominalEffectSiteMatch(
	site nominalUnknownCallSite,
	callable nominalEffectCallableInput,
	parameter nominalEffectParameterInput,
) (string, bool) {
	if generatedCallTargetsName(site.Callee, callable.GeneratedGoName) &&
		containsInt(site.ArgumentIndexes, parameter.Index) {
		return "exact-generated-target", true
	}
	method := generatedSelectorMethod(site.Callee)
	if method != "" &&
		strings.HasSuffix(callable.Callable, "."+method) &&
		containsInt(site.ArgumentIndexes, parameter.Index-1) {
		return "interface-method-candidate-set", true
	}
	if callable.Kind == "lambda" &&
		containsInt(site.ArgumentIndexes, parameter.Index) {
		return "indirect-callable-candidate-set", true
	}
	return "", false
}

func generatedCallTargetsName(callee, generatedName string) bool {
	if callee == "" || generatedName == "" {
		return false
	}
	for _, suffix := range []string{"", "_ctx"} {
		if callee == "__able_compiled_"+generatedName+suffix {
			return true
		}
	}
	return false
}

func generatedSelectorMethod(callee string) string {
	dot := strings.LastIndexByte(callee, '.')
	if dot < 0 {
		return ""
	}
	method := callee[dot+1:]
	return strings.TrimPrefix(method, "__able_ctx_")
}

func nominalEffectResolutionRank(resolution string) int {
	switch resolution {
	case "exact-generated-target":
		return 3
	case "interface-method-candidate-set":
		return 2
	case "indirect-callable-candidate-set":
		return 1
	default:
		return 0
	}
}

func containsInt(values []int, wanted int) bool {
	for _, value := range values {
		if value == wanted {
			return true
		}
	}
	return false
}
