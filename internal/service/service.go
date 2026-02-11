package service

import (
	"context"
	"fmt"
	"github.com/itering/scale.go/source"
	"github.com/itering/scale.go/types"
	"os"
	"strings"

	"github.com/itering/subscan/internal/cbc"
	"github.com/itering/subscan/internal/dao"
	"github.com/itering/subscan/util"
	"github.com/itering/substrate-api-rpc"
	"github.com/itering/substrate-api-rpc/metadata"
	"github.com/itering/substrate-api-rpc/websocket"
)

// Service
type Service struct {
	dao       dao.IDao
	dbStorage *dao.DbStorage
}

// New  a service and return.
func New() (s *Service) {
	websocket.SetEndpoint(util.WSEndPoint)
	d, dbStorage, pool := dao.New()
	s = &Service{dao: d, dbStorage: dbStorage}
	
	// CBC Chain specific initialization MUST run BEFORE initSubRuntimeLatest
	// because CBC metadata is too large for WebSocket and needs HTTP fetching
	if err := s.initCBCChain(); err != nil {
		util.Logger().Warning(fmt.Sprintf("CBC initialization warning: %v", err))
		// If CBC init fails, try standard initialization
		s.initSubRuntimeLatest()
	} else {
		// CBC init succeeded, skip standard metadata fetching
		// Register custom types but don't fetch metadata again
		s.initSubRuntimeLatestWithoutFetch()
	}
	
	s.unknownToken()
	pluginRegister(dbStorage, pool)
	return s
}

func (s *Service) GetDao() dao.IDao {
	return s.dao
}

func (s *Service) GetDbStorage() *dao.DbStorage {
	return s.dbStorage
}

// Close close the resource.
func (s *Service) Close() {
	s.dao.Close()
}

// initCBCChain performs CBC Chain specific initialization
func (s *Service) initCBCChain() error {
	util.Logger().Info("Initializing CBC Chain specific components...")
	
	// Check if this is CBC Chain by looking at network node config
	// Accept both "cbc" and "cbc-chain" as valid network names
	networkName := strings.ToLower(util.NetworkNode)
	if networkName != "cbc" && networkName != "cbc-chain" && !strings.Contains(networkName, "cbc") {
		util.Logger().Info(fmt.Sprintf("Not a CBC Chain network (network: %s), skipping CBC initialization", util.NetworkNode))
		return fmt.Errorf("not a CBC network")
	}
	
	util.Logger().Info(fmt.Sprintf("Detected CBC Chain network: %s", util.NetworkNode))
	
	// Create CBC initializer
	cbcInit := cbc.NewCBCInitializer(s.dao, util.WSEndPoint)
	
	// Initialize CBC runtime (bootstrap if needed)
	ctx := context.Background()
	if err := cbcInit.Initialize(ctx); err != nil {
		return fmt.Errorf("CBC initialization failed: %w", err)
	}
	
	// Verify DCF finality integration
	if err := cbcInit.VerifyDCFFinality(); err != nil {
		util.Logger().Warning(fmt.Sprintf("DCF finality verification warning: %v", err))
		// Don't fail on finality verification - it's informational
	}
	
	util.Logger().Info("CBC Chain initialization completed successfully")
	return nil
}

func (s *Service) initSubRuntimeLatest() {
	// reg network custom type
	defer func() {
		if data, err := readTypeRegistry(); err == nil {
			substrate.RegCustomTypes(data)
		}
		types.RegCustomTypes(map[string]source.TypeStruct{
			"WeightV2":              {Type: "struct", TypeMapping: [][]string{{"ref_time", "Compact<u64>"}, {"proofSize", "Compact<u64>"}}},
			"RuntimeDispatchInfo":   {Type: "struct", TypeMapping: [][]string{{"weight", "WeightV2"}, {"class", "DispatchClass"}, {"partialFee", "Balance"}}},
			"RuntimeDispatchInfoV1": {Type: "struct", TypeMapping: [][]string{{"weight", "Weight"}, {"class", "DispatchClass"}, {"partialFee", "Balance"}}},
		})
	}()

	// find db
	if recent := s.dao.RuntimeVersionRecent(); recent != nil && strings.HasPrefix(recent.RawData, "0x") {
		metadata.Latest(&metadata.RuntimeRaw{Spec: recent.SpecVersion, Raw: recent.RawData})
		return
	}
	// find metadata for blockChain
	var raw string
	for i := 0; i < 3; i++ {
		if raw = s.regCodecMetadata(); strings.HasPrefix(raw, "0x") {
			metadata.Latest(&metadata.RuntimeRaw{Spec: 1, Raw: raw})
			return
		}
		util.Logger().Warning(fmt.Sprintf("Attempt %d: regCodecMetadata failed, retrying...", i+1))
	}
	util.Logger().Error(fmt.Errorf("DEBUG: regCodecMetadata failed after 3 attempts. Raw length: %d", len(raw)))
	panic("Can not find chain metadata, please check network")
}

// initSubRuntimeLatestWithoutFetch registers custom types but doesn't fetch metadata
// Used when CBC initialization has already populated the database with metadata
func (s *Service) initSubRuntimeLatestWithoutFetch() {
	// reg network custom type
	defer func() {
		if data, err := readTypeRegistry(); err == nil {
			substrate.RegCustomTypes(data)
		}
		types.RegCustomTypes(map[string]source.TypeStruct{
			"WeightV2":              {Type: "struct", TypeMapping: [][]string{{"ref_time", "Compact<u64>"}, {"proofSize", "Compact<u64>"}}},
			"RuntimeDispatchInfo":   {Type: "struct", TypeMapping: [][]string{{"weight", "WeightV2"}, {"class", "DispatchClass"}, {"partialFee", "Balance"}}},
			"RuntimeDispatchInfoV1": {Type: "struct", TypeMapping: [][]string{{"weight", "Weight"}, {"class", "DispatchClass"}, {"partialFee", "Balance"}}},
		})
	}()

	// CBC has already populated the database, just load it
	if recent := s.dao.RuntimeVersionRecent(); recent != nil && strings.HasPrefix(recent.RawData, "0x") {
		util.Logger().Info(fmt.Sprintf("Loading CBC metadata from database (spec: %d, size: %d bytes)", recent.SpecVersion, len(recent.RawData)))
		metadata.Latest(&metadata.RuntimeRaw{Spec: recent.SpecVersion, Raw: recent.RawData})
		return
	}
	
	util.Logger().Warning("CBC initialization succeeded but no metadata found in database, falling back to standard initialization")
	s.initSubRuntimeLatest()
}

// read custom registry from local or remote
func readTypeRegistry() ([]byte, error) {
	return os.ReadFile(fmt.Sprintf(util.ConfDir+"/source/%s.json", util.NetworkNode))
}
