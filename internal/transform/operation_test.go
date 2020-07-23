package transform

import (
	"encoding/base64"
	"fmt"
	"testing"

	ingestio "github.com/stellar/go/exp/ingest/io"
	"github.com/stellar/go/xdr"
	"github.com/stretchr/testify/assert"
)

func TestTransformOperation(t *testing.T) {
	type operationInput struct {
		operation   xdr.Operation
		index       int32
		transaction ingestio.LedgerTransaction
	}
	type transformTest struct {
		input      operationInput
		wantOutput OperationOutput
		wantErr    error
	}
	genericInput := operationInput{
		operation:   genericOperation,
		index:       1,
		transaction: genericLedgerTransaction,
	}

	negativeOpTypeInput := genericInput
	negativeOpTypeEnvelope := genericEnvelope
	negativeOpTypeEnvelope.Tx.Operations[0].Body.Type = xdr.OperationType(-1)
	negativeOpTypeInput.operation.Body.Type = xdr.OperationType(-1)
	negativeOpTypeInput.transaction.Envelope.V1 = &negativeOpTypeEnvelope

	unknownOpTypeInput := genericInput
	unknownOpTypeEnvelope := genericEnvelope
	unknownOpTypeEnvelope.Tx.Operations[0].Body.Type = xdr.OperationType(20)
	unknownOpTypeInput.operation.Body.Type = xdr.OperationType(20)
	unknownOpTypeInput.transaction.Envelope.V1 = &unknownOpTypeEnvelope

	tests := []transformTest{
		{
			negativeOpTypeInput,
			OperationOutput{},
			fmt.Errorf("The operation type (-1) is negative for  operation 1"),
		},
		{
			unknownOpTypeInput,
			OperationOutput{},
			fmt.Errorf("Unknown operation type: "),
		},
	}
	hardCodedInputTransaction, err := prepareHardcodedOperationTestInput()
	assert.NoError(t, err)
	hardCodedOutputArray := prepareHardcodedOperationTestOutputs()

	for i, op := range hardCodedInputTransaction.Envelope.Operations() {
		tests = append(tests, transformTest{
			input:      operationInput{op, int32(i), hardCodedInputTransaction},
			wantOutput: hardCodedOutputArray[i],
			wantErr:    nil,
		})
	}

	for _, test := range tests {
		actualOutput, actualError := TransformOperation(test.input.operation, test.input.index, test.input.transaction)
		assert.Equal(t, test.wantErr, actualError)
		assert.Equal(t, test.wantOutput, actualOutput)
	}
}

//creates a single transaction that contains one of every operation type
func prepareHardcodedOperationTestInput() (inputTransaction ingestio.LedgerTransaction, err error) {
	inputTransaction = genericLedgerTransaction
	inputEnvelope := genericEnvelope

	inputEnvelope.Tx.SourceAccount = hardCodedAccountThree
	hardCodedInflationDest := hardCodedAccountFourID

	hardCodedTrustAsset, err := hardCodedUDSTAsset.ToAllowTrustOpAsset("USDT")
	if err != nil {
		return
	}

	hardCodedClearFlags := xdr.Uint32(3)
	hardCodedSetFlags := xdr.Uint32(4)
	hardCodedMasterWeight := xdr.Uint32(3)
	hardCodedLowThresh := xdr.Uint32(1)
	hardCodedMedThresh := xdr.Uint32(3)
	hardCodedHighThresh := xdr.Uint32(5)
	hardCodedHomeDomain := xdr.String32("2019=DRA;n-test")
	hardCodedSignerKey, err := xdr.NewSignerKey(xdr.SignerKeyTypeSignerKeyTypeEd25519, xdr.Uint256([32]byte{}))
	if err != nil {
		return
	}

	hardCodedSigner := xdr.Signer{
		Key:    hardCodedSignerKey,
		Weight: xdr.Uint32(1),
	}

	hardCodedDataValue := xdr.DataValue([]byte{0x76, 0x61, 0x6c, 0x75, 0x65})
	hardCodedSequenceNumber := xdr.SequenceNumber(100)
	inputOperations := []xdr.Operation{
		xdr.Operation{
			SourceAccount: nil,
			Body: xdr.OperationBody{
				Type: xdr.OperationTypeCreateAccount,
				CreateAccountOp: &xdr.CreateAccountOp{
					StartingBalance: 25000000,
					Destination:     hardCodedAccountFourID,
				},
			},
		},
		xdr.Operation{
			SourceAccount: nil,
			Body: xdr.OperationBody{
				Type: xdr.OperationTypePayment,
				PaymentOp: &xdr.PaymentOp{
					Destination: hardCodedAccountFour,
					Asset:       hardCodedUDSTAsset,
					Amount:      350000000,
				},
			},
		},
		xdr.Operation{
			SourceAccount: nil,
			Body: xdr.OperationBody{
				Type: xdr.OperationTypePayment,
				PaymentOp: &xdr.PaymentOp{
					Destination: hardCodedAccountFour,
					Asset:       hardCodedNativeAsset,
					Amount:      350000000,
				},
			},
		},
		xdr.Operation{
			SourceAccount: &hardCodedAccountThree,
			Body: xdr.OperationBody{
				Type: xdr.OperationTypePathPaymentStrictReceive,
				PathPaymentStrictReceiveOp: &xdr.PathPaymentStrictReceiveOp{
					SendAsset:   hardCodedNativeAsset,
					SendMax:     8951495900,
					Destination: hardCodedAccountFour,
					DestAsset:   hardCodedNativeAsset,
					DestAmount:  8951495900,
					Path:        []xdr.Asset{hardCodedUDSTAsset},
				},
			},
		},
		xdr.Operation{
			SourceAccount: nil,
			Body: xdr.OperationBody{
				Type: xdr.OperationTypeManageSellOffer,
				ManageSellOfferOp: &xdr.ManageSellOfferOp{
					Selling: hardCodedUDSTAsset,
					Buying:  hardCodedNativeAsset,
					Amount:  765860000,
					Price: xdr.Price{
						N: 128523,
						D: 250000,
					},
					OfferId: 0,
				},
			},
		},
		xdr.Operation{
			SourceAccount: nil,
			Body: xdr.OperationBody{
				Type: xdr.OperationTypeCreatePassiveSellOffer,
				CreatePassiveSellOfferOp: &xdr.CreatePassiveSellOfferOp{
					Selling: hardCodedNativeAsset,
					Buying:  hardCodedUDSTAsset,
					Amount:  631595000,
					Price: xdr.Price{
						N: 99583200,
						D: 1257990000,
					},
				},
			},
		},
		xdr.Operation{
			SourceAccount: nil,
			Body: xdr.OperationBody{
				Type: xdr.OperationTypeSetOptions,
				SetOptionsOp: &xdr.SetOptionsOp{
					InflationDest: &hardCodedInflationDest,
					ClearFlags:    &hardCodedClearFlags,
					SetFlags:      &hardCodedSetFlags,
					MasterWeight:  &hardCodedMasterWeight,
					LowThreshold:  &hardCodedLowThresh,
					MedThreshold:  &hardCodedMedThresh,
					HighThreshold: &hardCodedHighThresh,
					HomeDomain:    &hardCodedHomeDomain,
					Signer:        &hardCodedSigner,
				},
			},
		},
		xdr.Operation{
			SourceAccount: nil,
			Body: xdr.OperationBody{
				Type: xdr.OperationTypeChangeTrust,
				ChangeTrustOp: &xdr.ChangeTrustOp{
					Line:  hardCodedUDSTAsset,
					Limit: xdr.Int64(500000000000000000),
				},
			},
		},
		xdr.Operation{
			SourceAccount: nil,
			Body: xdr.OperationBody{
				Type: xdr.OperationTypeAllowTrust,
				AllowTrustOp: &xdr.AllowTrustOp{
					Trustor:   hardCodedAccountFourID,
					Asset:     hardCodedTrustAsset,
					Authorize: xdr.Uint32(1),
				},
			},
		},
		xdr.Operation{
			SourceAccount: nil,
			Body: xdr.OperationBody{
				Type:        xdr.OperationTypeAccountMerge,
				Destination: &hardCodedAccountFour,
			},
		},
		xdr.Operation{
			SourceAccount: nil,
			Body: xdr.OperationBody{
				Type: xdr.OperationTypeInflation,
			},
		},
		xdr.Operation{
			SourceAccount: nil,
			Body: xdr.OperationBody{
				Type: xdr.OperationTypeManageData,
				ManageDataOp: &xdr.ManageDataOp{
					DataName:  "test",
					DataValue: &hardCodedDataValue,
				},
			},
		},
		xdr.Operation{
			SourceAccount: nil,
			Body: xdr.OperationBody{
				Type: xdr.OperationTypeBumpSequence,
				BumpSequenceOp: &xdr.BumpSequenceOp{
					BumpTo: hardCodedSequenceNumber,
				},
			},
		},
		xdr.Operation{
			SourceAccount: nil,
			Body: xdr.OperationBody{
				Type: xdr.OperationTypeManageBuyOffer,
				ManageBuyOfferOp: &xdr.ManageBuyOfferOp{
					Selling:   hardCodedUDSTAsset,
					Buying:    hardCodedNativeAsset,
					BuyAmount: 7654501001,
					Price: xdr.Price{
						N: 635863285,
						D: 1818402817,
					},
					OfferId: 100,
				},
			},
		},
		xdr.Operation{
			SourceAccount: nil,
			Body: xdr.OperationBody{
				Type: xdr.OperationTypePathPaymentStrictSend,
				PathPaymentStrictSendOp: &xdr.PathPaymentStrictSendOp{
					SendAsset:   hardCodedNativeAsset,
					SendAmount:  1598182,
					Destination: hardCodedAccountFour,
					DestAsset:   hardCodedNativeAsset,
					DestMin:     4280460538,
					Path:        []xdr.Asset{hardCodedUDSTAsset},
				},
			},
		},
		xdr.Operation{
			SourceAccount: nil,
			Body: xdr.OperationBody{
				Type: xdr.OperationTypePathPaymentStrictSend,
				PathPaymentStrictSendOp: &xdr.PathPaymentStrictSendOp{
					SendAsset:   hardCodedNativeAsset,
					SendAmount:  1598182,
					Destination: hardCodedAccountFour,
					DestAsset:   hardCodedNativeAsset,
					DestMin:     4280460538,
					Path:        nil,
				},
			},
		},
	}
	inputEnvelope.Tx.Operations = inputOperations
	results := []xdr.OperationResult{
		xdr.OperationResult{},
		xdr.OperationResult{},
		xdr.OperationResult{},
		//need a true result for path payment receive
		xdr.OperationResult{
			Code: xdr.OperationResultCodeOpInner,
			Tr: &xdr.OperationResultTr{
				Type: xdr.OperationTypePathPaymentStrictReceive,
				PathPaymentStrictReceiveResult: &xdr.PathPaymentStrictReceiveResult{
					Code: xdr.PathPaymentStrictReceiveResultCodePathPaymentStrictReceiveSuccess,
					Success: &xdr.PathPaymentStrictReceiveResultSuccess{
						Last: xdr.SimplePaymentResult{Amount: 8946764349},
					},
				},
			},
		},
		xdr.OperationResult{},
		xdr.OperationResult{},
		xdr.OperationResult{},
		xdr.OperationResult{},
		xdr.OperationResult{},
		xdr.OperationResult{},
		xdr.OperationResult{},
		xdr.OperationResult{},
		xdr.OperationResult{},
		xdr.OperationResult{},
		xdr.OperationResult{
			Code: xdr.OperationResultCodeOpInner,
			Tr: &xdr.OperationResultTr{
				Type: xdr.OperationTypePathPaymentStrictSend,
				PathPaymentStrictSendResult: &xdr.PathPaymentStrictSendResult{
					Code: xdr.PathPaymentStrictSendResultCodePathPaymentStrictSendSuccess,
					Success: &xdr.PathPaymentStrictSendResultSuccess{
						Last: xdr.SimplePaymentResult{Amount: 4334043858},
					},
				},
			},
		},
		xdr.OperationResult{
			Code: xdr.OperationResultCodeOpInner,
			Tr: &xdr.OperationResultTr{
				Type: xdr.OperationTypePathPaymentStrictSend,
				PathPaymentStrictSendResult: &xdr.PathPaymentStrictSendResult{
					Code: xdr.PathPaymentStrictSendResultCodePathPaymentStrictSendSuccess,
					Success: &xdr.PathPaymentStrictSendResultSuccess{
						Last: xdr.SimplePaymentResult{Amount: 4280460538},
					},
				},
			},
		},
	}
	inputTransaction.Result.Result.Result.Results = &results
	inputTransaction.Envelope.V1 = &inputEnvelope
	return
}

func prepareHardcodedOperationTestOutputs() (transformedOperations []OperationOutput) {
	hardCodedSourceAccountAddress := hardCodedAccountThreeAddress
	hardCodedDestAccountAddress := hardCodedAccountFourAddress
	transformedOperations = []OperationOutput{
		OperationOutput{
			SourceAccount:    hardCodedSourceAccountAddress,
			Type:             0,
			ApplicationOrder: 1,
			OperationDetails: Details{
				Account:         hardCodedDestAccountAddress,
				Funder:          hardCodedSourceAccountAddress,
				StartingBalance: 2.5,
			},
		},
		OperationOutput{
			Type:             1,
			ApplicationOrder: 2,
			SourceAccount:    hardCodedSourceAccountAddress,

			OperationDetails: Details{
				From:        hardCodedSourceAccountAddress,
				To:          hardCodedDestAccountAddress,
				Amount:      35,
				AssetCode:   "USDT",
				AssetType:   "credit_alphanum4",
				AssetIssuer: hardCodedDestAccountAddress,
			},
		},
		OperationOutput{
			Type:             1,
			ApplicationOrder: 3,
			SourceAccount:    hardCodedSourceAccountAddress,

			OperationDetails: Details{
				From:      hardCodedSourceAccountAddress,
				To:        hardCodedDestAccountAddress,
				Amount:    35,
				AssetType: "native",
			},
		},
		OperationOutput{
			Type:             2,
			ApplicationOrder: 4,
			SourceAccount:    hardCodedSourceAccountAddress,

			OperationDetails: Details{
				From:            hardCodedSourceAccountAddress,
				To:              hardCodedDestAccountAddress,
				SourceAmount:    894.6764349,
				SourceMax:       895.14959,
				Amount:          895.14959,
				SourceAssetType: "native",
				AssetType:       "native",
				Path:            []AssetOutput{hardCodedUSDTAssetOutput},
			},
		},
		OperationOutput{
			Type:             3,
			ApplicationOrder: 5,
			SourceAccount:    hardCodedSourceAccountAddress,
			OperationDetails: Details{
				Price:  0.514092,
				Amount: 76.586,
				PriceR: Price{
					Numerator:   128523,
					Denominator: 250000,
				},
				SellingAssetCode:   "USDT",
				SellingAssetType:   "credit_alphanum4",
				SellingAssetIssuer: hardCodedDestAccountAddress,
				BuyingAssetType:    "native",
			},
		},
		OperationOutput{
			Type:             4,
			ApplicationOrder: 6,
			SourceAccount:    hardCodedSourceAccountAddress,
			OperationDetails: Details{
				Amount: 63.1595,
				Price:  0.0791606,
				PriceR: Price{
					Numerator:   99583200,
					Denominator: 1257990000,
				},
				BuyingAssetCode:   "USDT",
				BuyingAssetType:   "credit_alphanum4",
				BuyingAssetIssuer: hardCodedDestAccountAddress,
				SellingAssetType:  "native",
			},
		},
		OperationOutput{
			Type:             5,
			ApplicationOrder: 7,
			SourceAccount:    hardCodedSourceAccountAddress,
			OperationDetails: Details{
				InflationDest:    hardCodedDestAccountAddress,
				ClearFlags:       []int32{1, 2},
				ClearFlagsString: []string{"auth_required", "auth_revocable"},
				SetFlags:         []int32{4},
				SetFlagsString:   []string{"auth_immutable"},
				MasterKeyWeight:  3,
				LowThreshold:     1,
				MedThreshold:     3,
				HighThreshold:    5,
				HomeDomain:       "2019=DRA;n-test",
				SignerKey:        "GAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAWHF",
				SignerWeight:     1,
			},
		},
		OperationOutput{
			Type:             6,
			ApplicationOrder: 8,
			SourceAccount:    hardCodedSourceAccountAddress,
			OperationDetails: Details{
				Trustor:     hardCodedSourceAccountAddress,
				Trustee:     hardCodedDestAccountAddress,
				Limit:       50000000000,
				AssetCode:   "USDT",
				AssetType:   "credit_alphanum4",
				AssetIssuer: hardCodedDestAccountAddress,
			},
		},
		OperationOutput{
			Type:             7,
			ApplicationOrder: 9,
			SourceAccount:    hardCodedSourceAccountAddress,
			OperationDetails: Details{
				Trustee:     hardCodedSourceAccountAddress,
				Trustor:     hardCodedDestAccountAddress,
				Authorize:   true,
				AssetCode:   "USDT",
				AssetType:   "credit_alphanum4",
				AssetIssuer: hardCodedSourceAccountAddress,
			},
		},
		OperationOutput{
			Type:             8,
			ApplicationOrder: 10,
			SourceAccount:    hardCodedSourceAccountAddress,
			OperationDetails: Details{
				Account: hardCodedSourceAccountAddress,
				Into:    hardCodedDestAccountAddress,
			},
		},
		OperationOutput{
			Type:             9,
			ApplicationOrder: 11,
			SourceAccount:    hardCodedSourceAccountAddress,
			OperationDetails: Details{},
		},
		OperationOutput{
			Type:             10,
			ApplicationOrder: 12,
			SourceAccount:    hardCodedSourceAccountAddress,
			OperationDetails: Details{
				Name:  "test",
				Value: base64.StdEncoding.EncodeToString([]byte{0x76, 0x61, 0x6c, 0x75, 0x65}),
			},
		},
		OperationOutput{
			Type:             11,
			ApplicationOrder: 13,
			SourceAccount:    hardCodedSourceAccountAddress,
			OperationDetails: Details{
				BumpTo: "100",
			},
		},
		OperationOutput{
			Type:             12,
			ApplicationOrder: 14,
			SourceAccount:    hardCodedSourceAccountAddress,
			OperationDetails: Details{
				Price:  0.3496823,
				Amount: 765.4501001,
				PriceR: Price{
					Numerator:   635863285,
					Denominator: 1818402817,
				},
				SellingAssetCode:   "USDT",
				SellingAssetType:   "credit_alphanum4",
				SellingAssetIssuer: hardCodedDestAccountAddress,
				BuyingAssetType:    "native",
				OfferID:            100,
			},
		},
		OperationOutput{
			Type:             13,
			ApplicationOrder: 15,
			SourceAccount:    hardCodedSourceAccountAddress,
			OperationDetails: Details{
				From:            hardCodedSourceAccountAddress,
				To:              hardCodedDestAccountAddress,
				SourceAmount:    0.1598182,
				DestinationMin:  "428.0460538",
				Amount:          433.4043858,
				Path:            []AssetOutput{hardCodedUSDTAssetOutput},
				SourceAssetType: "native",
				AssetType:       "native",
			},
		},
		OperationOutput{
			Type:             13,
			ApplicationOrder: 16,
			SourceAccount:    hardCodedSourceAccountAddress,
			OperationDetails: Details{
				From:            hardCodedSourceAccountAddress,
				To:              hardCodedDestAccountAddress,
				SourceAmount:    0.1598182,
				DestinationMin:  "428.0460538",
				Amount:          428.0460538,
				Path:            nil,
				SourceAssetType: "native",
				AssetType:       "native",
			},
		},
	}
	return
}
