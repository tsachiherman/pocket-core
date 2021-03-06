package keeper

import (
	"github.com/pokt-network/pocket-core/x/pocketcore/types"
	sdk "github.com/pokt-network/posmint/types"
)

// "HandleDispatch" - Handles a client request for their session information
func (k Keeper) HandleDispatch(ctx sdk.Ctx, header types.SessionHeader) (*types.DispatchResponse, sdk.Error) {
	// retrieve the latest session block height
	latestSessionBlockHeight := k.GetLatestSessionBlockHeight(ctx)
	// set the session block height
	header.SessionBlockHeight = latestSessionBlockHeight
	// validate the header
	err := header.ValidateHeader()
	if err != nil {
		return nil, err
	}
	// get the session context
	sessionCtx, er := ctx.PrevCtx(latestSessionBlockHeight)
	if er != nil {
		return nil, sdk.ErrInternal(er.Error())
	}
	// check cache
	session, found := types.GetSession(header)
	// if not found generate the session
	if !found {
		var err sdk.Error
		session, err = types.NewSession(sessionCtx, ctx, k.posKeeper, header, types.BlockHash(sessionCtx), int(k.SessionNodeCount(sessionCtx)))
		if err != nil {
			return nil, err
		}
		// add to cache
		types.SetSession(session)
	}
	return &types.DispatchResponse{Session: session, BlockHeight: ctx.BlockHeight()}, nil
}

// "IsSessionBlock" - Returns true if current block, is a session block (beginning of a session)
func (k Keeper) IsSessionBlock(ctx sdk.Ctx) bool {
	return ctx.BlockHeight()%k.posKeeper.BlocksPerSession(ctx) == 1
}

// "GetLatestSessionBlockHeight" - Returns the latest session block height (first block of the session, (see blocksPerSession))
func (k Keeper) GetLatestSessionBlockHeight(ctx sdk.Ctx) (sessionBlockHeight int64) {
	// get the latest block height
	blockHeight := ctx.BlockHeight()
	// get the blocks per session
	blocksPerSession := k.posKeeper.BlocksPerSession(ctx)
	// if block height / blocks per session remainder is zero, just subtract blocks per session and add 1
	if blockHeight%blocksPerSession == 0 {
		sessionBlockHeight = blockHeight - k.posKeeper.BlocksPerSession(ctx) + 1
	} else {
		// calculate the latest session block height by diving the current block height by the blocksPerSession
		sessionBlockHeight = (blockHeight/blocksPerSession)*blocksPerSession + 1
	}
	return
}

// "IsPocketSupportedBlockchain" - Returns true if network identifier param is supported by pocket
func (k Keeper) IsPocketSupportedBlockchain(ctx sdk.Ctx, chain string) bool {
	// loop through supported blockchains (network identifiers)
	for _, c := range k.SupportedBlockchains(ctx) {
		// if contains chain return true
		if c == chain {
			return true
		}
	}
	// else return false
	return false
}

func (Keeper) ClearSessionCache() {
	types.ClearSessionCache()
}
