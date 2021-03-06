package keeper

import (
	"fmt"
	"github.com/pokt-network/pocket-core/x/nodes/types"
	sdk "github.com/pokt-network/posmint/types"
	"github.com/pokt-network/posmint/x/auth"
	"github.com/pokt-network/posmint/x/auth/exported"
)

// GetStakedTokens - Retrieve total staking tokens supply which is staked
func (k Keeper) GetStakedTokens(ctx sdk.Ctx) sdk.Int {
	stakedPool := k.GetStakedPool(ctx)
	return stakedPool.GetCoins().AmountOf(k.StakeDenom(ctx))
}

// TotalTokens - Retrieve staking tokens from the total supply
func (k Keeper) TotalTokens(ctx sdk.Ctx) sdk.Int {
	return k.AccountKeeper.GetSupply(ctx).GetTotal().AmountOf(k.StakeDenom(ctx))
}

// GetStakedPool - Retrieve the staked tokens pool's module account
func (k Keeper) GetStakedPool(ctx sdk.Ctx) (stakedPool exported.ModuleAccountI) {
	return k.AccountKeeper.GetModuleAccount(ctx, types.StakedPoolName)
}

// coinsFromStakedToUnstaked - Transfer coins from the module account to the validator -> used in unstaking
func (k Keeper) coinsFromStakedToUnstaked(ctx sdk.Ctx, validator types.Validator) error {
	coins := sdk.NewCoins(sdk.NewCoin(k.StakeDenom(ctx), validator.StakedTokens))
	err := k.AccountKeeper.SendCoinsFromModuleToAccount(ctx, types.StakedPoolName, validator.Address, coins)
	if err != nil {
		return fmt.Errorf("unable to send coins from staked to unstaked for address: %s", validator.Address)
	}
	return nil
}

// coinsFromUnstakedToStaked - Transfer coins from the module account to validator -> used in staking
func (k Keeper) coinsFromUnstakedToStaked(ctx sdk.Ctx, validator types.Validator, amount sdk.Int) sdk.Error {
	if amount.LT(sdk.ZeroInt()) {
		return sdk.ErrInternal("cannot send a negative")
	}
	coins := sdk.NewCoins(sdk.NewCoin(k.StakeDenom(ctx), amount))
	err := k.AccountKeeper.SendCoinsFromAccountToModule(ctx, validator.Address, types.StakedPoolName, coins)
	return err
}

// burnStakedTokens - Removes coins from the staked pool module account
func (k Keeper) burnStakedTokens(ctx sdk.Ctx, amt sdk.Int) sdk.Error {
	if !amt.IsPositive() {
		return nil
	}
	coins := sdk.NewCoins(sdk.NewCoin(k.StakeDenom(ctx), amt))
	return k.AccountKeeper.BurnCoins(ctx, types.StakedPoolName, coins)
}

// getFeePool - Retrieve fee pool
func (k Keeper) getFeePool(ctx sdk.Ctx) (feePool exported.ModuleAccountI) {
	return k.AccountKeeper.GetModuleAccount(ctx, auth.FeeCollectorName)
}
