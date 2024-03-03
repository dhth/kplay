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
	item KMsgItem
}
