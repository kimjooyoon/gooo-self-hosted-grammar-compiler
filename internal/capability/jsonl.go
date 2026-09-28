package capability

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"strings"
)

type CapabilityJSONLRequest struct {
	ID      json.RawMessage `json:"id"`
	Grammar string          `json:"grammar"`
	Query   string          `json:"query"`
}

type CapabilityJSONLResponse struct {
	ID       json.RawMessage `json:"id"`
	Report   Report         `json:"report"`
	Evidence Evidence       `json:"evidence"`
	ReadOnly bool           `json:"read_only"`
}

func ServeCapabilityJSONL(input io.Reader, output io.Writer) error {
	scanner := bufio.NewScanner(input)
	scanner.Buffer(make([]byte, 1024), 1024*1024)
	encoder := json.NewEncoder(output)
	lineNumber := 0
	for scanner.Scan() {
		lineNumber++
		if strings.TrimSpace(scanner.Text()) == "" {
			continue
		}
		var request CapabilityJSONLRequest
		if err := json.Unmarshal(scanner.Bytes(), &request); err != nil {
			return fmt.Errorf("capability request line %d: %w", lineNumber, err)
		}
		report := Discover([]byte(request.Grammar), request.Query)
		if err := report.Validate(); err != nil {
			return fmt.Errorf("capability report line %d: %w", lineNumber, err)
		}
		evidence := report.Evidence()
		if err := evidence.Validate(); err != nil {
			return fmt.Errorf("capability evidence line %d: %w", lineNumber, err)
		}
		response := CapabilityJSONLResponse{
			ID:       request.ID,
			Report:   report,
			Evidence: evidence,
			ReadOnly: true,
		}
		if err := encoder.Encode(response); err != nil {
			return fmt.Errorf("capability response line %d: %w", lineNumber, err)
		}
	}
	return scanner.Err()
}
