package cmux

import (
	"encoding/json"
	"fmt"
	"net"
	"sync/atomic"
	"time"
)

const (
	dialTimeout = 2 * time.Second
	rwTimeout   = 5 * time.Second
)

var rpcID atomic.Int64

type rpcRequest struct {
	Jsonrpc string      `json:"jsonrpc"`
	ID      int64       `json:"id"`
	Method  string      `json:"method"`
	Params  interface{} `json:"params,omitempty"`
}

type rpcResponse struct {
	Jsonrpc string          `json:"jsonrpc"`
	ID      int64           `json:"id"`
	Result  json.RawMessage `json:"result,omitempty"`
	Error   *rpcError       `json:"error,omitempty"`
}

type rpcError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

func (e *rpcError) Error() string { return fmt.Sprintf("cmux rpc %d: %s", e.Code, e.Message) }

// Call sends a JSON-RPC 2.0 request to the cmux socket and decodes the result.
// Pass result=nil for fire-and-forget calls (the response is still consumed).
func Call(method string, params, result interface{}) error {
	if !Detect() {
		return ErrNotDetected
	}
	conn, err := net.DialTimeout("unix", SocketPath(), dialTimeout)
	if err != nil {
		return fmt.Errorf("cmux dial: %w", err)
	}
	defer conn.Close()
	_ = conn.SetDeadline(time.Now().Add(rwTimeout))

	req := rpcRequest{
		Jsonrpc: "2.0",
		ID:      rpcID.Add(1),
		Method:  method,
		Params:  params,
	}
	if err := json.NewEncoder(conn).Encode(&req); err != nil {
		return fmt.Errorf("cmux write %s: %w", method, err)
	}

	var resp rpcResponse
	if err := json.NewDecoder(conn).Decode(&resp); err != nil {
		return fmt.Errorf("cmux read %s: %w", method, err)
	}
	if resp.Error != nil {
		return resp.Error
	}
	if result != nil && len(resp.Result) > 0 {
		if err := json.Unmarshal(resp.Result, result); err != nil {
			return fmt.Errorf("cmux decode %s result: %w", method, err)
		}
	}
	return nil
}
