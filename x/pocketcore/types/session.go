package types

import (
	"encoding/hex"
	"encoding/json"
	appexported "github.com/pokt-network/pocket-core/x/apps/exported"
	nodeexported "github.com/pokt-network/pocket-core/x/nodes/exported"
	"github.com/pokt-network/pocket-core/x/nodes/types"
	"github.com/pokt-network/posmint/codec"
	sdk "github.com/pokt-network/posmint/types"
	"sort"
)

// a session is the relationship between an application and the pocket network
type Session struct {
	SessionHeader `json:"header"`
	SessionKey    `json:"key"`
	SessionNodes  `json:"nodes"`
}

// create a new session from seed data
func NewSession(appPubKey string, nonNativeChain string, blockHash string, blockHeight int64, allActiveNodes []nodeexported.ValidatorI, sessionNodesCount int) (*Session, sdk.Error) {
	// first generate session key
	sessionKey, err := NewSessionKey(appPubKey, nonNativeChain, blockHash)
	if err != nil {
		return nil, err
	}
	// then generate the service nodes for that session
	sessionNodes, err := NewSessionNodes(nonNativeChain, sessionKey, allActiveNodes, sessionNodesCount)
	if err != nil {
		return nil, err
	}
	// then populate the structure and return
	return &Session{
		SessionKey: sessionKey,
		SessionHeader: SessionHeader{
			ApplicationPubKey:  appPubKey,
			Chain:              nonNativeChain,
			SessionBlockHeight: blockHeight,
		},
		SessionNodes: sessionNodes,
	}, nil
}

func (s Session) Validate(ctx sdk.Context, node nodeexported.ValidatorI, app appexported.ApplicationI, sessionNodeCount int) sdk.Error {
	// validate chain
	if len(s.Chain) == 0 {
		return NewEmptyNonNativeChainError(ModuleName)
	}
	// validate sessionBlockHeight
	if s.SessionBlockHeight <= 0 {
		return NewInvalidBlockHeightError(ModuleName)
	}
	// validate the app public key
	if err := PubKeyVerification(s.ApplicationPubKey); err != nil {
		return err
	}
	// validate app corresponds to appPubKey
	if hex.EncodeToString(app.GetConsPubKey().Bytes()) != s.ApplicationPubKey {
		return NewInvalidAppPubKeyError(ModuleName)
	}
	// validate app chains
	chains := app.GetChains()
	found := false
	for _, c := range chains {
		if c == s.Chain {
			found = true
			break
		}
	}
	if !found {
		return NewUnsupportedBlockchainAppError(ModuleName)
	}
	// validate sessionNodes
	err := s.SessionNodes.Validate(sessionNodeCount)
	if err != nil {
		return err
	}
	// validate node is of the session
	if !s.SessionNodes.Contains(node) {
		return NewInvalidSessionError(ModuleName)
	}
	return nil
}

// service nodes in a session
type SessionNodes []nodeexported.ValidatorI

func (sn SessionNodes) Validate(sessionNodesCount int) sdk.Error {
	if len(sn) < sessionNodesCount {
		return NewInsufficientNodesError(ModuleName)
	}
	return nil
}

func (sn SessionNodes) Contains(nodeVerify nodeexported.ValidatorI) bool { // todo use a map instead of a slice to save time
	if nodeVerify == nil {
		return false
	}
	for _, node := range sn {
		if node.GetConsPubKey().Equals(nodeVerify.GetConsPubKey()) {
			return true
		}
	}
	return false
}

func (sn *SessionNodes) MarshalJSON() ([]byte, error) {
	var marshable types.Validators
	marshable = make(types.Validators, len(*sn))
	for i, expNode := range *sn {
		marshable[i] = types.Validator{ // todo depends on nodes.Types()
			Address:      expNode.GetAddress(),
			ConsPubKey:   expNode.GetConsPubKey(),
			Jailed:       expNode.IsJailed(),
			Status:       expNode.GetStatus(),
			Chains:       expNode.GetChains(),
			ServiceURL:   expNode.GetServiceURL(),
			StakedTokens: expNode.GetTokens()}
	}
	return codec.Cdc.MarshalJSON(marshable)
}

// UnmarshalJSON unmarshals the validator from JSON // todo test if function needed
func (sn *SessionNodes) UnmarshalJSON(data []byte) error {
	v := &types.Validators{} // todo depends on nodes.Types()
	if err := codec.Cdc.UnmarshalJSON(data, v); err != nil {
		return err
	}
	for _, nonExNode := range *v {
		*sn = append(*sn, nonExNode)
	}
	return nil
}

// generates nodes for the session
func NewSessionNodes(chain string, sessionKey SessionKey, allNodes []nodeexported.ValidatorI, sessionNodesCount int) (SessionNodes, sdk.Error) {
	// validate chain
	if len(chain) == 0 {
		return nil, NewEmptyNonNativeChainError(ModuleName)
	}
	// validate sessionKey
	if err := sessionKey.Validate(); err != nil {
		return nil, NewInvalidSessionKeyError(ModuleName, err)
	}
	// validate allNodes
	if len(allNodes) < sessionNodesCount {
		return nil, NewInsufficientNodesError(ModuleName)
	}
	// filter `allNodes` by the HASH(chain)
	sortedNodeDistances, err := filter(allNodes, chain, sessionNodesCount)
	if err != nil {
		return nil, NewFilterNodesError(ModuleName, err)
	}
	// xor each node's public key and session key
	nodeDistances, err := xor(sortedNodeDistances, sessionKey)
	if err != nil {
		return nil, NewXORError(ModuleName, err)
	}
	// sort the nodes based off of distance
	sortedNodeDistances = revSort(nodeDistances)
	// return the top 5 nodes
	return sortedNodeDistances[0:sessionNodesCount], nil
}

// filter the nodes by non native chain
func filter(allActiveNodes []nodeexported.ValidatorI, nonNativeChainHash string, sessionNodesCount int) (SessionNodes, error) {
	var result SessionNodes
	for _, node := range allActiveNodes {
		chains := node.GetChains()
		contains := false
		// todo get rid of slice and use map (amino doesn't support map encoding so custom struct to encode and decode)
		for _, chain := range chains {
			if chain == nonNativeChainHash {
				contains = true
			}
		}
		if !contains {
			continue
		}
		result = append(result, node)
	}
	if err := result.Validate(sessionNodesCount); err != nil {
		return nil, err
	}
	return result, nil
}

// A node linked to it's computational distance
type nodeDistance struct {
	Node     nodeexported.ValidatorI
	distance []byte
}

type nodeDistances []nodeDistance

// xor the sessionNodes.publicKey against the sessionKey to find the computationally closest nodes
func xor(sessionNodes SessionNodes, sessionkey SessionKey) (nodeDistances, error) {
	var keyLength = len(sessionkey)
	result := make([]nodeDistance, len(sessionNodes))
	// for every node, find the distance between it's pubkey and the sesskey
	for index, node := range sessionNodes {
		pubKey := node.GetConsPubKey()
		pubKeyHash := Hash(pubKey.Bytes()) // currently hashing public key but could easily just take the first n bytes to compare
		if len(pubKeyHash) != keyLength {
			return nil, MismatchedByteArraysError
		}
		result[index].Node = node
		result[index].distance = make([]byte, keyLength)
		for i := 0; i < keyLength; i++ {
			result[index].distance[i] = pubKeyHash[i] ^ sessionkey[i]
		}
	}
	return result, nil
}

// sort the nodes by shortest computational distance
func revSort(sessionNodes nodeDistances) SessionNodes {
	result := make(SessionNodes, len(sessionNodes))
	sort.Sort(sort.Reverse(sessionNodes))
	for _, node := range sessionNodes {
		result = append(result, node.Node)
	}
	return result
}

// returns the length of the node pool -> needed for sort.Sort() interface
func (n nodeDistances) Len() int { return len(n) }

// swaps two elements in the node pool -> needed for sort.Sort() interface
func (n nodeDistances) Swap(i, j int) { n[i], n[j] = n[j], n[i] }

// returns if node i is less than node j by XOR value (big endian encoding)
func (n nodeDistances) Less(i, j int) bool {
	// compare size of byte arrays
	if len(n[i].distance) < len(n[j].distance) {
		return false
	}
	// bitwise comparison
	for a := range n[i].distance {
		if n[i].distance[a] < n[j].distance[a] {
			return true
		}
		if n[i].distance[a] < n[i].distance[a] {
			return false
		}
	}
	return false
}

// the hash identifier of the session
type SessionKey []byte

type sessionKey struct {
	AppPublicKey   string `json:"app_pub_key"`
	NonNativeChain string `json:"chain"`
	BlockHash      string `json:"blockchain"`
}

// generates the session key
func NewSessionKey(appPubKey string, chain string, blockHash string) (SessionKey, sdk.Error) {
	// validate appPubKey
	if err := PubKeyVerification(appPubKey); err != nil {
		return nil, err
	}
	// validate chain
	if err := HashVerification(chain); err != nil {
		return nil, NewEmptyChainError(ModuleName)
	}
	// validate block hash
	if err := HashVerification(blockHash); err != nil {
		return nil, err
	}
	// marshal into json
	seed, err := json.Marshal(sessionKey{
		AppPublicKey:   appPubKey,
		NonNativeChain: chain,
		BlockHash:      blockHash,
	})
	if err != nil {
		return nil, NewJSONMarshalError(ModuleName, err)
	}
	// return the hash of the result
	return Hash(seed), nil
}

func (sk SessionKey) Validate() sdk.Error {
	return HashVerification(hex.EncodeToString(sk))
}

func BlockHashFromBlockHeight(ctx sdk.Context, height int64) string {
	return hex.EncodeToString(ctx.WithBlockHeight(height).BlockHeader().LastBlockId.Hash)
}

// proof of relay header
type SessionHeader struct {
	ApplicationPubKey  string `json:"app_pub_key"`
	Chain              string `json:"chain"`
	SessionBlockHeight int64  `json:"session_height"`
}

func (sh SessionHeader) ValidateHeader() sdk.Error {
	if err := PubKeyVerification(sh.ApplicationPubKey); err != nil {
		return err
	}
	if err := HashVerification(sh.Chain); err != nil {
		return err
	}
	if sh.SessionBlockHeight <= 1 {
		return NewInvalidBlockHeightError(ModuleName)
	}
	return nil
}

// hash the header bytes
func (sh SessionHeader) Hash() []byte {
	res := sh.Bytes()
	return Hash(res)
}

// hex encode the header hash
func (sh SessionHeader) HashString() string {
	return hex.EncodeToString(sh.Hash())
}

// get the bytes of the header structure
func (sh SessionHeader) Bytes() []byte {
	res, err := json.Marshal(sh)
	if err != nil {
		panic(err)
	}
	return res
}