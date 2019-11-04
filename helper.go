package main

import (
	. "github.com/MediConCenHK/go-chaincode-common"
	. "github.com/davidkhala/goutils"
)

const keyDuration = "duration"
const keyMemberID = "memberID"
const keyPersonal = "personal"
const keyTokenCandidates = "tokenCandidates"

func getDuration(transient map[string][]byte) TimeLong {
	var durationBytes = transient[keyDuration]
	if durationBytes == nil {
		return 0
	}
	return TimeLong(Atoi(string(durationBytes)))
}
func getMemberID(transient map[string][]byte) string {
	var memberIDBytes = EnsureTransientMap(transient, keyMemberID)
	return string(memberIDBytes)
}

func getPersonalInfo(transient map[string][]byte) []byte {
	return transient[keyPersonal]
}

type tokenCandidates struct {
	TokenVerify string
	TokenPay    string
}

func getTokenCandidates(transient map[string][]byte) tokenCandidates {
	var tokens tokenCandidates
	FromJson(EnsureTransientMap(transient, keyTokenCandidates), &tokens)
	return tokens
}
