package model

import (
	"github.com/twmb/franz-go/pkg/kgo"
)

//	type KMsg struct{
//	    Offset int64
//	    Key []byte
//	    Value []byte
//	    Header []byte
//
// }
type KMsgFetchedMsg struct {
	records []*kgo.Record
	err     error
}

type KMsgChosenMsg struct {
	key string
}

type RecordSavedToDiskMsg struct {
	path string
	err  error
}

type KMsgMetadataReadyMsg struct {
	key         string
	msgMetadata string
}

type KMsgDataReadyMsg struct {
	storeKey    string
	record      *kgo.Record
	msgMetadata string
	msgValue    string
	err         error
}
