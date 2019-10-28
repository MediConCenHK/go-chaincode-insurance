package main

import (
	. "github.com/MediConCenHK/go-chaincode-common"
	. "github.com/davidkhala/fabric-common-chaincode-golang"
	"github.com/davidkhala/fabric-common-chaincode-golang/cid"
	. "github.com/davidkhala/goutils"
	"github.com/hyperledger/fabric/core/chaincode/shim"
	"github.com/hyperledger/fabric/protos/peer"
	"strings"
)

const (
	collectionMember   = "member"
	keyAppUserCert     = "AppUserCert" // cert of app user allowed to invoke chaincode
	keyTokenCandidates = "tokenCandidates"
)

type InsuranceCC struct {
	InsuranceChaincode
	Payer
}

func (t InsuranceCC) verifyCreatorIdentity(expectedCert []byte) {
	creatorCert := cid.NewClientIdentity(t.CCAPI).CertificatePem

	if strings.Compare(string(creatorCert), string(expectedCert)) != 0 {
		t.Logger.Error("creator", string(creatorCert))
		t.Logger.Error("expectedCert", string(expectedCert))
		PanicString("tx creator's identity is not as expected")
	}

}

func (t InsuranceCC) GenTokens(params []string) []byte {
	// TODO WIP stub
	return nil
}

func (t InsuranceCC) getMemberData(params []string) []byte {

	// TODO WIP stub
	return nil
}

func (t InsuranceCC) Init(stub shim.ChaincodeStubInterface) (response peer.Response) {
	defer Deferred(DeferHandlerPeerResponse, &response)
	t.Prepare(stub)
	t.Logger.Info("Init")

	// AppUserCert is used to validate the tx creator's identity
	// when init, pass in the BC blockchain application's user certificate
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
	t.Logger.Info("Invoke", fcn)
	t.Logger.Debug("Invoke", fcn, params)
	var responseBytes []byte

	// validate creator's identity, only BC blockchain application's user is allowed
	t.verifyCreatorIdentity(t.GetState(keyAppUserCert))

	switch fcn {
	case "underwrite":

	case "genTokens":
		responseBytes = t.GenTokens(params)
	case "getMemberData":
		responseBytes = t.getMemberData(params)
	default:
		PanicString("unknown fcn:" + fcn)
	}
	return shim.Success(responseBytes)

}

func main() {
	var cc = InsuranceCC{InsuranceChaincode: NewInsuranceChaincode("")}
	ChaincodeStart(cc)
}
