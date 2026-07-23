package service

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/itering/cbcscan/util"
)

type CBCValidator struct {
	Name            string  `json:"name"`
	Address         string  `json:"address"`
	AccountId       string  `json:"account_id"`
	PosScore        uint64  `json:"pos_score"`
	PoiScore        uint64  `json:"poi_score"`
	TrustScore      uint64  `json:"trust_score"`
	BondedStake     string  `json:"bonded_stake"`
	BlocksAuthored  uint64  `json:"blocks_authored"`
	BlocksMissed    uint64  `json:"blocks_missed"`
	UptimePercent   float64 `json:"uptime_percent"`
	Status          string  `json:"status"`
}

type CBCDCFMetrics struct {
	Epoch            uint64 `json:"epoch"`
	PosWeight        uint32 `json:"pos_weight"`
	PoiWeight        uint32 `json:"poi_weight"`
	BlocksPerEpoch   uint32 `json:"blocks_per_epoch"`
	ActiveValidators int    `json:"active_validators"`
}

type CBCDVFMetrics struct {
	FinalizedBlock     uint64  `json:"finalized_block"`
	FinalityThreshold  float64 `json:"finality_threshold"`
	CheckInterval      uint32  `json:"check_interval"`
	CurrentVotingRound uint32  `json:"current_voting_round"`
	Status             string  `json:"status"`
	VoteTally          uint64  `json:"vote_tally"`
	TotalWeight        uint64  `json:"total_weight"`
	CheckpointBlock    uint64  `json:"checkpoint_block"`
}

func rpcCall(method string, params []interface{}) ([]byte, error) {
	nodeRPC := util.GetEnv("CHAIN_WS_ENDPOINT", "ws://127.0.0.1:9944")
	httpRPC := strings.Replace(nodeRPC, "ws://", "http://", 1)
	httpRPC = strings.Replace(httpRPC, "wss://", "https://", 1)

	payload := map[string]interface{}{
		"id":      1,
		"jsonrpc": "2.0",
		"method":  method,
		"params":  params,
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("POST", httpRPC, strings.NewReader(string(body)))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 3 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	return io.ReadAll(resp.Body)
}

func (s *Service) GetCBCValidators(ctx context.Context) ([]CBCValidator, error) {
	// 1. Get live list of validator addresses from cbc_listValidators
	valRes, err := rpcCall("cbc_listValidators", []interface{}{})
	var addresses []string
	if err == nil {
		var resp struct {
			Result []string `json:"result"`
		}
		if err := json.Unmarshal(valRes, &resp); err == nil && len(resp.Result) > 0 {
			addresses = resp.Result
		}
	}

	if len(addresses) == 0 {
		addresses = []string{
			"5FA9nQDVg267DEd8m1ZypXLBnvN7SFxYwV7ndqSYGiN9TTpu",
			"5GoNkf6WdbxCFnPdAnYYQyCjAKPJgLNxXwPjwTh6DGg6gN3E",
			"5DbKjhNLpqX3zqZdNBc9BGb4fHU1cRBaDhJUskrvkwfraDi6",
		}
	}

	var validators []CBCValidator
	for idx, addr := range addresses {
		name := fmt.Sprintf("Validator #%d", idx+1)
		if strings.Contains(addr, "5FA9") {
			name = "Alice (Bootnode)"
		} else if strings.Contains(addr, "5GoN") {
			name = "Bob (Validator)"
		} else if strings.Contains(addr, "5DbK") {
			name = "Charlie (Validator)"
		}

		v := CBCValidator{
			Name:           name,
			Address:        addr,
			AccountId:      addr,
			PosScore:       80,
			PoiScore:       80,
			TrustScore:     50,
			BondedStake:    "10000000000000000000",
			BlocksAuthored: 0,
			BlocksMissed:   0,
			UptimePercent:  100.0,
			Status:         "Active Validator",
		}

		// Query dynamic profile for this validator address
		profRes, err := rpcCall("cbc_getValidatorProfile", []interface{}{addr})
		if err == nil {
			var profResp struct {
				Result struct {
					Account        string      `json:"account"`
					Stake          interface{} `json:"stake"`
					PosScore       uint64      `json:"posScore"`
					PoiScore       uint64      `json:"poiScore"`
					TrustScore     uint64      `json:"trustScore"`
					Status         string      `json:"status"`
					AuthoredBlocks uint64      `json:"authoredBlocks"`
					MissedBlocks   uint64      `json:"missedBlocks"`
				} `json:"result"`
			}
			if err := json.Unmarshal(profRes, &profResp); err == nil && profResp.Result.Account != "" {
				v.PosScore = profResp.Result.PosScore
				v.PoiScore = profResp.Result.PoiScore
				v.TrustScore = profResp.Result.TrustScore
				v.BlocksAuthored = profResp.Result.AuthoredBlocks
				v.BlocksMissed = profResp.Result.MissedBlocks
				if profResp.Result.Status != "" {
					v.Status = profResp.Result.Status
				}

				if profResp.Result.AuthoredBlocks+profResp.Result.MissedBlocks > 0 {
					v.UptimePercent = (float64(profResp.Result.AuthoredBlocks) / float64(profResp.Result.AuthoredBlocks+profResp.Result.MissedBlocks)) * 100.0
				}

				v.BondedStake = fmt.Sprintf("%v", profResp.Result.Stake)
			}
		}

		validators = append(validators, v)
	}

	return validators, nil
}

func (s *Service) GetCBCDCFMetrics(ctx context.Context) (CBCDCFMetrics, error) {
	metrics := CBCDCFMetrics{
		Epoch:            1,
		PosWeight:        60,
		PoiWeight:        40,
		BlocksPerEpoch:   100,
		ActiveValidators: 3,
	}

	// Dynamic epoch
	resEpoch, err := rpcCall("cbc_getCurrentEpoch", []interface{}{})
	if err == nil {
		var resp struct {
			Result uint64 `json:"result"`
		}
		if err := json.Unmarshal(resEpoch, &resp); err == nil {
			metrics.Epoch = resp.Result
		}
	}

	// Dynamic DCF weights
	resWeights, err := rpcCall("dcf_getConsensusWeights", []interface{}{})
	if err == nil {
		var resp struct {
			Result struct {
				PosWeight uint32 `json:"posWeight"`
				PoiWeight uint32 `json:"poiWeight"`
			} `json:"result"`
		}
		if err := json.Unmarshal(resWeights, &resp); err == nil {
			if resp.Result.PosWeight > 0 || resp.Result.PoiWeight > 0 {
				metrics.PosWeight = resp.Result.PosWeight / 100
				metrics.PoiWeight = resp.Result.PoiWeight / 100
			}
		}
	}

	return metrics, nil
}

func parseUint64(v interface{}) uint64 {
	if val, ok := v.(float64); ok {
		return uint64(val)
	}
	if val, ok := v.(string); ok {
		if strings.HasPrefix(val, "0x") {
			if num, err := strconv.ParseUint(strings.TrimPrefix(val, "0x"), 16, 64); err == nil {
				return num
			}
		} else {
			if num, err := strconv.ParseUint(val, 10, 64); err == nil {
				return num
			}
		}
	}
	return 0
}

func (s *Service) GetCBCDVFMetrics(ctx context.Context) (CBCDVFMetrics, error) {
	metrics := CBCDVFMetrics{
		FinalizedBlock:     0,
		FinalityThreshold:  67.0,
		CheckInterval:      10,
		CurrentVotingRound: 1,
		Status:             "Active Quorum",
		VoteTally:          0,
		TotalWeight:        300, // fallback
		CheckpointBlock:    0,
	}

	// Dynamic finalized head
	resFin, err := rpcCall("dvf_getFinalizedHead", []interface{}{})
	if err == nil {
		var resp struct {
			Result uint64 `json:"result"`
		}
		if err := json.Unmarshal(resFin, &resp); err == nil {
			metrics.FinalizedBlock = resp.Result
		}
	}

	// Dynamic voting round
	resRound, err := rpcCall("dvf_getCurrentRound", []interface{}{})
	if err == nil {
		var resp struct {
			Result uint32 `json:"result"`
		}
		if err := json.Unmarshal(resRound, &resp); err == nil {
			metrics.CurrentVotingRound = resp.Result
		}
	}

	// Dynamic validator weights
	resWeights, err := rpcCall("dvf_getValidatorWeights", []interface{}{})
	if err == nil {
		var resp struct {
			Result [][]interface{} `json:"result"`
		}
		if err := json.Unmarshal(resWeights, &resp); err == nil {
			var total uint64
			for _, item := range resp.Result {
				if len(item) == 2 {
					total += parseUint64(item[1])
				}
			}
			if total > 0 {
				metrics.TotalWeight = total
			}
		}
	}

	// Dynamic best block to calculate active checkpoint
	checkpointBlock := uint64(0)
	resHeader, err := rpcCall("chain_getHeader", []interface{}{})
	if err == nil {
		var resp struct {
			Result struct {
				Number string `json:"number"`
			} `json:"result"`
		}
		if err := json.Unmarshal(resHeader, &resp); err == nil {
			bestBlock := parseUint64(resp.Result.Number)
			if metrics.CheckInterval > 0 {
				checkpointBlock = (bestBlock / uint64(metrics.CheckInterval)) * uint64(metrics.CheckInterval)
			}
		}
	}
	if checkpointBlock == 0 && metrics.FinalizedBlock > 0 {
		checkpointBlock = (metrics.FinalizedBlock / uint64(metrics.CheckInterval)) * uint64(metrics.CheckInterval)
	}
	metrics.CheckpointBlock = checkpointBlock

	// Dynamic voting tally for active checkpoint
	if checkpointBlock > 0 {
		resHash, err := rpcCall("chain_getBlockHash", []interface{}{checkpointBlock})
		if err == nil {
			var resp struct {
				Result string `json:"result"`
			}
			if err := json.Unmarshal(resHash, &resp); err == nil && resp.Result != "" && !strings.Contains(resp.Result, "0x00000000000") {
				resTally, err := rpcCall("dvf_getAccumulatedWeight", []interface{}{resp.Result})
				if err == nil {
					var tallyResp struct {
						Result interface{} `json:"result"`
					}
					if err := json.Unmarshal(resTally, &tallyResp); err == nil {
						metrics.VoteTally = parseUint64(tallyResp.Result)
					}
				}
			}
		}
	}

	return metrics, nil
}
