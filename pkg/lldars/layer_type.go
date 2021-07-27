package lldars

const (
	DiscoverBroadcast LLDARSLayerType = iota
	ServicePortNotify
	GetObjectRequest
	DeliveryObject
	EndOfDelivery
	BackupObjectRequest // backup要求 (あってもいいけど定期実行したいかも)
	SyncObjectRequest   // backupのための同期要求
	AcceptSyncingObject // 同期要求へのack
	EndOfSync
)

func (t LLDARSLayerType) String() string {
	return toLayerTypeString[t]
}

var toLayerTypeString = map[LLDARSLayerType]string{
	DiscoverBroadcast:   "DiscoverBroadcast",
	ServicePortNotify:   "ServicePortNotify",
	GetObjectRequest:    "GetObjectRequest",
	DeliveryObject:      "DeliveryObject",
	EndOfDelivery:       "EndOfDelivery",
	BackupObjectRequest: "BackupObjectRequest",
	SyncObjectRequest:   "SyncObjectRequest",
	AcceptSyncingObject: "AcceptSyncingObject",
	EndOfSync:           "EndOfSync",
}
