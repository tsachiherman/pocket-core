package keeper

import (
	"github.com/pokt-network/pocket-core/x/nodes/types"
	sdk "github.com/pokt-network/posmint/types"
	"github.com/stretchr/testify/assert"
	abci "github.com/tendermint/tendermint/abci/types"
	"reflect"
	"testing"
)

func TestKeeper_BeginUnstakingValidator(t *testing.T) {
	type fields struct {
		keeper Keeper
	}
	type args struct {
		ctx       sdk.Context
		validator types.Validator
	}
	validator := getBondedValidator()
	context, _, keeper := createTestInput(t, true)
	tests := []struct {
		name   string
		fields fields
		args   args
		want   sdk.Error
	}{
		{"Test BeginUnstakingValidator", fields{keeper: keeper}, args{
			ctx:       context,
			validator: validator,
		}, nil},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			k := tt.fields.keeper
			if got := k.BeginUnstakingValidator(tt.args.ctx, tt.args.validator); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("BeginUnstakingValidator() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestKeeper_FinishUnstakingValidator(t *testing.T) {
	type fields struct {
		keeper Keeper
	}

	type args struct {
		ctx       sdk.Context
		validator types.Validator
	}

	validator := getBondedValidator()
	context, _, keeper := createTestInput(t, true)

	tests := []struct {
		name   string
		fields fields
		args   args
		want   sdk.Error
	}{
		{"Test FinishUnstakingValidator", fields{keeper: keeper}, args{
			ctx:       context,
			validator: validator,
		}, nil},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			k := tt.fields.keeper
			if got := k.FinishUnstakingValidator(tt.args.ctx, tt.args.validator); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("FinishUnstakingValidator() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestKeeper_JailValidator(t *testing.T) {
	type fields struct {
		keeper Keeper
	}
	type args struct {
		ctx  sdk.Context
		addr sdk.Address
	}

	validator := getBondedValidator()
	context, _, keeper := createTestInput(t, true)
	keeper.SetValidator(context, validator)
	keeper.SetValidatorByConsAddr(context, validator)

	tests := []struct {
		name   string
		fields fields
		args   args
	}{
		{"Test JailValidator", fields{keeper: keeper}, args{
			ctx:  context,
			addr: validator.ConsAddress(),
		}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			k := tt.fields.keeper
			k.JailValidator(tt.args.ctx, tt.args.addr)
		})
	}
}

func TestKeeper_RegisterValidator(t *testing.T) {
	type fields struct {
		keeper Keeper
	}
	type args struct {
		ctx       sdk.Context
		validator types.Validator
	}

	validator := getBondedValidator()
	context, _, keeper := createTestInput(t, true)

	tests := []struct {
		name   string
		fields fields
		args   args
	}{
		{"Test RegisterValidator", fields{keeper: keeper}, args{
			ctx:       context,
			validator: validator,
		}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			k := tt.fields.keeper
			k.RegisterValidator(tt.args.ctx, tt.args.validator)
		})
	}
}

func TestKeeper_ReleaseWaitingValidators(t *testing.T) {
	type fields struct {
		keeper Keeper
	}
	type args struct {
		ctx sdk.Context
	}

	validator := getUnbondingValidator()
	context, _, keeper := createTestInput(t, true)

	tests := []struct {
		name   string
		fields fields
		args   args
	}{
		{"Test ReleaseWaitingValidators", fields{keeper: keeper}, args{ctx: context}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			k := tt.fields.keeper
			k.SetWaitingValidator(tt.args.ctx, validator)
			k.ReleaseWaitingValidators(tt.args.ctx)
		})
	}
}

func TestKeeper_StakeValidator(t *testing.T) {
	type fields struct {
		keeper Keeper
	}
	type args struct {
		ctx       sdk.Context
		validator types.Validator
		amount    sdk.Int
	}

	validator := getBondedValidator()
	context, _, keeper := createTestInput(t, true)

	tests := []struct {
		name   string
		fields fields
		args   args
		want   sdk.Error
	}{
		{"Test StakeValidator", fields{keeper: keeper}, args{
			ctx:       context,
			validator: validator,
			amount:    sdk.ZeroInt(),
		}, nil},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			k := tt.fields.keeper
			if got := k.StakeValidator(tt.args.ctx, tt.args.validator, tt.args.amount); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("StakeValidator() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestKeeper_UnjailValidator(t *testing.T) {
	type fields struct {
		keeper Keeper
	}
	type args struct {
		ctx  sdk.Context
		addr sdk.Address
	}
	validator := getBondedValidator()
	context, _, keeper := createTestInput(t, true)
	validator.Jailed = true
	keeper.SetValidator(context, validator)
	keeper.SetValidatorByConsAddr(context, validator)

	tests := []struct {
		name   string
		fields fields
		args   args
	}{
		{"Test UnjailValidator", fields{keeper: keeper}, args{
			ctx:  context,
			addr: validator.ConsAddress(),
		}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			k := tt.fields.keeper
			k.UnjailValidator(tt.args.ctx, tt.args.addr)
		})
	}
}

func TestKeeper_UpdateTendermintValidators(t *testing.T) {
	type fields struct {
		keeper Keeper
	}
	type args struct {
		ctx sdk.Context
	}

	//validator := getBondedValidator()
	context, _, keeper := createTestInput(t, true)

	tests := []struct {
		name        string
		fields      fields
		args        args
		wantUpdates []abci.ValidatorUpdate
	}{
		{"Test UpdateTenderMintValidators", fields{keeper: keeper}, args{ctx: context},
			[]abci.ValidatorUpdate{}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			k := tt.fields.keeper
			if gotUpdates := k.UpdateTendermintValidators(tt.args.ctx); !assert.True(t, len(gotUpdates) == len(tt.wantUpdates)) {
				t.Errorf("UpdateTendermintValidators() = %v, want %v", gotUpdates, tt.wantUpdates)
			}
		})
	}
}

func TestKeeper_ValidateValidatorBeginUnstaking(t *testing.T) {
	type fields struct {
		keeper Keeper
	}
	type args struct {
		ctx       sdk.Context
		validator types.Validator
	}

	validator := getBondedValidator()
	context, _, keeper := createTestInput(t, true)

	tests := []struct {
		name   string
		fields fields
		args   args
		want   sdk.Error
	}{
		{"Test ValidateValidatorBeginUnstaking", fields{keeper: keeper}, args{
			ctx:       context,
			validator: validator,
		}, nil},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			k := tt.fields.keeper
			if got := k.ValidateValidatorBeginUnstaking(tt.args.ctx, tt.args.validator); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ValidateValidatorBeginUnstaking() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestKeeper_ValidateValidatorFinishUnstaking(t *testing.T) {
	type fields struct {
		keeper Keeper
	}
	type args struct {
		ctx       sdk.Context
		validator types.Validator
	}

	validator := getUnbondingValidator()
	context, _, keeper := createTestInput(t, true)

	tests := []struct {
		name   string
		fields fields
		args   args
		want   sdk.Error
	}{
		{"Test ValidateValidatorFinishUnstaking", fields{keeper: keeper}, args{
			ctx:       context,
			validator: validator,
		}, nil},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			k := tt.fields.keeper
			if got := k.ValidateValidatorFinishUnstaking(tt.args.ctx, tt.args.validator); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ValidateValidatorFinishUnstaking() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestKeeper_ValidateValidatorStaking(t *testing.T) {
	type fields struct {
		keeper Keeper
	}
	type args struct {
		ctx       sdk.Context
		validator types.Validator
		amount    sdk.Int
	}

	validator := getUnbondedValidator()
	context, _, keeper := createTestInput(t, true)

	tests := []struct {
		name   string
		fields fields
		args   args
		want   sdk.Error
	}{
		{"Test ValidateValidatorStaking", fields{keeper: keeper}, args{
			ctx:       context,
			validator: validator,
			amount:    sdk.OneInt(),
		}, types.ErrNotEnoughCoins(types.ModuleName)},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			k := tt.fields.keeper
			if got := k.ValidateValidatorStaking(tt.args.ctx, tt.args.validator, tt.args.amount); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ValidateValidatorStaking() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestKeeper_WaitToBeginUnstakingValidator(t *testing.T) {
	type fields struct {
		keeper Keeper
	}
	type args struct {
		ctx       sdk.Context
		validator types.Validator
	}

	validator := getBondedValidator()
	context, _, keeper := createTestInput(t, true)

	tests := []struct {
		name   string
		fields fields
		args   args
		want   sdk.Error
	}{
		{"Test WaitToBeginUnstakingValidator", fields{keeper: keeper}, args{
			ctx:       context,
			validator: validator,
		}, nil},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			k := tt.fields.keeper
			if got := k.WaitToBeginUnstakingValidator(tt.args.ctx, tt.args.validator); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("WaitToBeginUnstakingValidator() = %v, want %v", got, tt.want)
			}
		})
	}
}