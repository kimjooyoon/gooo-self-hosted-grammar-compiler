package model

// ImprovementIdentity is the only admissible identity for before/after
// improvement evidence. Every field is exact; there is no aggregate score or
// estimated percentage fallback.
type ImprovementIdentity struct {
	Scenario        string `json:"scenario"`
	SourceDigest    string `json:"source_digest"`
	ContractDigest  string `json:"contract_digest"`
	ToolchainDigest string `json:"toolchain_digest"`
	RunnerDigest    string `json:"runner_digest"`
}

func ExactImprovementPair(before, after ImprovementIdentity) bool {
	return before.Scenario != "" && before == after && before.SourceDigest != "" && before.ContractDigest != "" && before.ToolchainDigest != "" && before.RunnerDigest != ""
}
