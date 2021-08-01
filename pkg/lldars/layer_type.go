package lldars

const (
	DiscoverBroadcast   LLDARSLayerType = iota // サーバ探索
	ServicePortNotify                          // BCへのAck
	GetObjectRequest                           // オブジェクト取得要求
	DeliveryObject                             // オブジェクト配送
	EndOfDelivery                              // オブジェクト配送終了
	ReceivedObjects                            // オブジェクト受け取り完了
	BackupObjectRequest                        // backupのための受信要求
	AcceptBackupObject                         // backupのための受信要求へのAck
	RejectBackupObject                         // backupのための受信要求を却下
	SyncObjectRequest                          // 復元のための送信要求
	AcceptSyncObject                           // 復元のための送信要求へのAck
	RejectSyncObject                           // 復元のための送信要求を却下
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
	AcceptBackupObject:  "AcceptSyncObject",
	SyncObjectRequest:   "SyncObjectRequest",
	AcceptSyncObject:    "AcceptSyncObject",
}
