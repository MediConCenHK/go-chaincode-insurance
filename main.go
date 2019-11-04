package main

import (
	. "github.com/MediConCenHK/go-chaincode-common"
	. "github.com/davidkhala/fabric-common-chaincode-golang"
	"github.com/davidkhala/fabric-common-chaincode-golang/cid"
	. "github.com/davidkhala/goutils"
	"github.com/hyperledger/fabric/core/chaincode/shim"
	"github.com/hyperledger/fabric/protos/peer"
)

const (
	collectionMember = "member"
	keyAppUserCert   = "appUserCert" // cert of app user allowed to invoke chaincode

)

type InsuranceCC struct {
	InsuranceChaincode
}

// validate creator's identity, only insurance user is allowed
func (t InsuranceCC) verifyCreatorIdentity(expectedCert []byte) {
	creatorCert := cid.NewClientIdentity(t.CCAPI).CertificatePem

	if string(creatorCert) == string(expectedCert) {
		t.Logger.Error("creator", string(creatorCert))
		t.Logger.Error("expectedCert", string(expectedCert))
		PanicString("tx creator's identity is not as expected")
	}

}

func (t InsuranceCC) tokenCandidate(tokenType TokenType) (token string) {
	var transient = t.GetTransient()
	var tokens = getTokenCandidates(transient)

	switch tokenType {
	case TokenTypePay:
		token = tokens.TokenPay
	case TokenTypeVerify:
		token = tokens.TokenVerify
	}

	var exist = GetToken(*t.CommonChaincode, token)
	if exist != nil {
		PanicString("assertion fail: token candidate[" + token + "] exists already")
	}
	return
}

func (t InsuranceCC) getExpiryTime() TimeLong {
	var transient = t.GetTransient()
	var duration = getDuration(transient)
	if duration == 0 {
		return 0
	} else {
		var txTime TimeLong
		txTime = txTime.FromTimeStamp(t.GetTxTimestamp())
		return txTime + duration
	}
}
func (t InsuranceCC) createAndSaveToken(memberID string, tokenType TokenType) (string, TimeLong) {

	var expiryTime = t.getExpiryTime()

	var token = t.tokenCandidate(tokenType)

	var tokenCreateRequest = TokenCreateRequest{
		Owner:      memberID,
		TokenType:  tokenType,
		ExpiryDate: expiryTime,
	}

	CreateToken(*t.CommonChaincode, token, tokenCreateRequest)
	return token, expiryTime

}

func (t InsuranceCC) renewToken(token string) TimeLong {
	var newExpiryTime = t.getExpiryTime()

	RenewToken(*t.CommonChaincode, token, newExpiryTime)
	return newExpiryTime

}
func (t InsuranceCC) updateTokenByCase(memberID string, tokenType TokenType, currentToken string) (string, TimeLong) {

	var currentTokenData = GetToken(*t.CommonChaincode, currentToken)
	var txTime TimeLong
	txTime = txTime.FromTimeStamp(t.GetTxTimestamp())
	if currentTokenData == nil || (currentTokenData.TokenType == TokenTypePay && currentTokenData.TransferDate > 0) {
		return t.createAndSaveToken(memberID, tokenType)
	} else {

		var timeOut = func(currentTime, expiryTime TimeLong) bool {
			return currentTime > expiryTime && expiryTime > 0
		}
		if timeOut(txTime, currentTokenData.ExpiryDate) {
			return currentToken, t.renewToken(currentToken)
		}

	}

	return currentToken, currentTokenData.ExpiryDate
}
func (t InsuranceCC) genTokens(memberID string) []byte {

	var data memberData
	exist := t.GetPrivateObj(collectionMember, memberID, &data)

	var changed = false
	var tokenVerifyExpiryTime TimeLong
	var tokenPayExpiryTime TimeLong
	if exist {
		dataBefore := data
		data.TokenVerify, tokenVerifyExpiryTime = t.updateTokenByCase(memberID, TokenTypeVerify, data.TokenVerify)
		data.TokenPay, tokenPayExpiryTime = t.updateTokenByCase(memberID, TokenTypePay, data.TokenPay)
		changed = (data.TokenVerify != dataBefore.TokenVerify) || (data.TokenPay != dataBefore.TokenPay)
	} else {
		data = memberData{}
		data.TokenVerify, tokenVerifyExpiryTime = t.createAndSaveToken(memberID, TokenTypeVerify)
		data.TokenPay, tokenPayExpiryTime = t.createAndSaveToken(memberID, TokenTypePay)
		changed = true
	}

	if changed {
		t.PutPrivateObj(collectionMember, memberID, data)
	}

	type responseData struct {
		memberData
		TokenVerifyExpiryTime TimeLong
		TokenPayExpiryTime    TimeLong
	}
	return ToJson(responseData{data, tokenVerifyExpiryTime, tokenPayExpiryTime})

}
func (t InsuranceCC) createMember(memberID string, personalInfo []byte) {
	var data memberData

	var exist = t.GetPrivateObj(collectionMember, memberID, &data)
	if !exist {
		data.PersonalInfo = personalInfo
		t.PutPrivateObj(collectionMember, memberID, data)
	} else {
		PanicString("[" + memberID + "] member data exist")
	}
}

func (t InsuranceCC) getMemberData(memberID string) []byte {
	var data memberData

	var exist = t.GetPrivateObj(collectionMember, memberID, &data)
	if ! exist {
		return nil
	} else {
		return ToJson(data)
	}
}

func (t InsuranceCC) Init(stub shim.ChaincodeStubInterface) (response peer.Response) {
	defer Deferred(DeferHandlerPeerResponse, &response)
	t.Prepare(stub)
	t.Logger.Info("Init")

	var _, args = t.GetFunctionAndArgs()

	if len(args) > 0 {
		var appUserCertPem = args[0]
		t.PutState(keyAppUserCert, appUserCertPem)
	}

	return shim.Success(nil)
}
func (t InsuranceCC) Invoke(stub shim.ChaincodeStubInterface) (response peer.Response) {
	defer Deferred(DeferHandlerPeerResponse, &response)
	t.Prepare(stub)
	var fcn, params = stub.GetFunctionAndParameters()

	t.verifyCreatorIdentity(t.GetState(keyAppUserCert))
	t.Logger.Info("Invoke", fcn)
	t.Logger.Debug("Invoke", fcn, params)
	var responseBytes []byte

	var transient = t.GetTransient()
	var memberID = getMemberID(transient)
	switch fcn {
	case "underwrite":
		var personalBytes = getPersonalInfo(transient)
		t.createMember(memberID, personalBytes)
	case "genTokens":
		responseBytes = t.genTokens(memberID)
	case "getMemberData":
		responseBytes = t.getMemberData(memberID)
	default:
		PanicString("unknown fcn:" + fcn)
	}
	return shim.Success(responseBytes)

}

func main() {
	var cc = InsuranceCC{InsuranceChaincode: NewInsuranceChaincode("")}
	ChaincodeStart(cc)
}
