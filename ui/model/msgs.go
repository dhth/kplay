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
	storeKey    string
	record      *kgo.Record
	msgMetadata string
}

type KMsgValueReadyMsg struct {
	storeKey string
	record   *kgo.Record
	msgValue string
	err      error
}
