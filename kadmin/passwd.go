package kadmin

import (
	"gopkg.in/jcmturner/gokrb5.v4/crypto"
	"gopkg.in/jcmturner/gokrb5.v4/krberror"
	"gopkg.in/jcmturner/gokrb5.v4/messages"
	"gopkg.in/jcmturner/gokrb5.v4/types"
)

func ChangePasswdMsg(cname types.PrincipalName, realm, password string, tkt messages.Ticket, sessionKey types.EncryptionKey) (r Request, k types.EncryptionKey, err error) {
	// Create change password data struct and marshal to bytes
	chgpasswd := ChangePasswdData{
		NewPasswd: []byte(password),
		TargName:  cname,
		TargRealm: realm,
	}
	chpwdb, err := chgpasswd.Marshal()
	if err != nil {
		err = krberror.Errorf(err, krberror.KRBMsgError, "error marshaling change passwd data")
		return
	}

	// Generate authenticator
	auth, err := types.NewAuthenticator(realm, cname)
	if err != nil {
		err = krberror.Errorf(err, krberror.KRBMsgError, "error generating new authenticator")
		return
	}
	etype, err := crypto.GetEtype(sessionKey.KeyType)
	if err != nil {
		err = krberror.Errorf(err, krberror.KRBMsgError, "error generating subkey etype")
		return
	}
	err = auth.GenerateSeqNumberAndSubKey(sessionKey.KeyType, etype.GetKeyByteSize())
	if err != nil {
		err = krberror.Errorf(err, krberror.KRBMsgError, "error generating subkey")
		return
	}
	k = auth.SubKey

	// Generate AP_REQ
	APreq, err := messages.NewAPReq(tkt, sessionKey, auth)
	if err != nil {
		return
	}

	// Form the KRBPriv encpart data
	//TODO set the SAddress field???
	kp := messages.EncKrbPrivPart{
		UserData:       chpwdb,
		Timestamp:      auth.CTime,
		Usec:           auth.Cusec,
		SequenceNumber: auth.SeqNumber,
	}
	kpriv := messages.NewKRBPriv(kp)
	err = kpriv.EncryptEncPart(k)
	//_, kp.UserData, err = etype.EncryptData(auth.SubKey.KeyValue, chpwdb)
	if err != nil {
		err = krberror.Errorf(err, krberror.EncryptingError, "error encrypting change passwd data")
		return
	}

	r = Request{
		APREQ:   APreq,
		KRBPriv: kpriv,
	}
	return
}