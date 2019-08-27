package types

const (
	ModuleName   = "commerciodocs"
	StoreKey     = ModuleName
	QuerierRoute = ModuleName

	MsgTypeShareDocument   = "shareDocument"
	MsgTypeDocumentReceipt = "documentReceipt"

	QuerySentDocuments     = "sent"
	QueryReceivedDocuments = "received"
	QueryReceipts          = "receipts"
	QueryUuidReceipt       = "receipt"

	//KVStore prefix
	SentDocumentsPrefix     = StoreKey + "sentBy:"
	ReceivedDocumentsPrefix = StoreKey + "received:"
	DocumentReceiptPrefix   = StoreKey + "receiptOf:"
	ReceiptsCounter         = StoreKey + "receiptsNumber:"
)
