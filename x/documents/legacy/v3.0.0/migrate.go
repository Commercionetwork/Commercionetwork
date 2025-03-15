package v3_0_0

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	v220docs "github.com/commercionetwork/commercionetwork/x/documents/legacy/v2.2.0"
	"github.com/commercionetwork/commercionetwork/x/documents/types"
)

// Migrate accepts exported genesis state from v2.2.0 and migrates it to v3.0.0
// genesis state. This migration changes the data that are saved for each document
// removing the metadataSchema and proposers.
// A DocumentReceipt will be included only there is a corresponding Document (that is, with the same UUID field).
func Migrate(oldGenState v220docs.GenesisState) *types.GenesisState {

	//documents
	documents := []*types.Document{}
	var document *types.Document

	documentsDeleted := make(map[string]struct{})
	appears := map[string]struct{}{}

	for _, v220document := range oldGenState.Documents {

		// invalid 2.2.0 documents won't be added
		if err := v220document.Validate(); err == nil {

			if _, found := appears[string(v220document.UUID)]; !found {

				document = migrateDocument(v220document)
				documents = append(documents, document)

				appears[string(v220document.UUID)] = struct{}{}
			}
		} else {
			documentsDeleted[v220document.UUID] = struct{}{}
		}

	}

	//document receipts
	receipts := []*types.DocumentReceipt{}
	var documentReceipt *types.DocumentReceipt
	for _, v220documentReceipt := range oldGenState.Receipts {

		if _, found := appears[v220documentReceipt.DocumentUUID]; found {
			if _, deleted := documentsDeleted[v220documentReceipt.DocumentUUID]; !deleted {
				documentReceipt = migrateReceipt(v220documentReceipt)
				receipts = append(receipts, documentReceipt)
			}
		}
	}

	return &types.GenesisState{
		Documents: documents,
		Receipts:  receipts,
	}
}

//convert a slice of sdk.accAddr to a slice of string
func fromSliceOfAddrToSliceOfString(Addresses []sdk.AccAddress) []string {
	var strings []string

	for _, s := range Addresses {
		strings = append(strings, s.String())
	}

	return strings
}

// migrateDocument migrates a single v2.2.0 document into a 3.0.0 document
func migrateDocument(doc v220docs.Document) *types.Document {
	// Convert the metadata schemes
	var documentMetadataSchema *types.DocumentMetadataSchema
	if doc.Metadata.Schema != nil {
		documentMetadataSchema = &types.DocumentMetadataSchema{
			URI:     doc.Metadata.Schema.URI,
			Version: doc.Metadata.Schema.Version,
		}
	}

	// Convert the encryption data
	var encryptionData *types.DocumentEncryptionData
	if doc.EncryptionData != nil {

		// Convert encryption keys
		keys := make([]*types.DocumentEncryptionKey, len(doc.EncryptionData.Keys))
		for i, key := range doc.EncryptionData.Keys {
			keys[i] = &types.DocumentEncryptionKey{
				Recipient: key.Recipient.String(),
				Value:     key.Value,
			}
		}

		encryptionData = &types.DocumentEncryptionData{
			Keys:          keys,
			EncryptedData: doc.EncryptionData.EncryptedData,
		}
	}

	//convert the Do sign
	var doSign *types.DocumentDoSign
	if doc.DoSign != nil {
		doSign = &types.DocumentDoSign{
			StorageURI:         doc.DoSign.StorageURI,
			SignerInstance:     doc.DoSign.SignerInstance,
			SdnData:            doc.DoSign.SdnData,
			VcrID:              doc.DoSign.VcrID,
			CertificateProfile: doc.DoSign.CertificateProfile,
		}
	}

	var doChecksum *types.DocumentChecksum
	if doc.Checksum != nil {
		doChecksum = &types.DocumentChecksum{
			Value:     doc.Checksum.Value,
			Algorithm: doc.Checksum.Algorithm,
		}
	}

	// Return a new document
	return &types.Document{
		Sender:     doc.Sender.String(),
		Recipients: fromSliceOfAddrToSliceOfString(doc.Recipients),
		UUID:       doc.UUID,
		Metadata: &types.DocumentMetadata{
			ContentURI: doc.Metadata.ContentURI,
			Schema:     documentMetadataSchema,
		},
		ContentURI:     doc.ContentURI,
		Checksum:       doChecksum,
		EncryptionData: encryptionData,
		DoSign:         doSign,
	}
}

// migrateReceipt migrates a v2.2.0 document receipt into a v3.0.0 document receipt
func migrateReceipt(receipt v220docs.DocumentReceipt) *types.DocumentReceipt {
	return &types.DocumentReceipt{
		UUID:         receipt.UUID,
		Sender:       receipt.Sender.String(),
		Recipient:    receipt.Recipient.String(),
		TxHash:       receipt.TxHash,
		DocumentUUID: receipt.DocumentUUID,
		Proof:        receipt.Proof,
	}
}
