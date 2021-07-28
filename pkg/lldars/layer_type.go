package lldars

const (
	DiscoverBroadcast        LLDARSLayerType = iota // サーバ探索
	ServicePortNotify                               // BCへのAck
	GetObjectRequest                                // オブジェクト取得要求
	DeliveryObject                                  // オブジェクト配送
	EndOfDelivery                                   // オブジェクト配送終了
	StartBackupObjectRequest                        // backup要求 (あってもいいけど定期実行したいかも)
	BackupObjectRequest                             // backupのための受信要求
	AcceptBackupObject                              // backupのための受信要求へのAck
	SyncObjectRequest                               // 復元のための送信要求
	AcceptSyncObject                                // 復元のための送信要求へのAck
)

func (t LLDARSLayerType) String() string {
	return toLayerTypeString[t]
}

var toLayerTypeString = map[LLDARSLayerType]string{
	DiscoverBroadcast:        "DiscoverBroadcast",
	ServicePortNotify:        "ServicePortNotify",
	GetObjectRequest:         "GetObjectRequest",
	DeliveryObject:           "DeliveryObject",
	EndOfDelivery:            "EndOfDelivery",
	StartBackupObjectRequest: "StartBackupObjectRequest",
	BackupObjectRequest:      "BackupObjectRequest",
	AcceptBackupObject:       "AcceptSyncObject",
	SyncObjectRequest:        "SyncObjectRequest",
	AcceptSyncObject:         "AcceptSyncObject",
}
