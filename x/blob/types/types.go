package types

// Event types and attributes
const (
	EventTypeBlobMetadataReceived    = "blob_metadata_received"
	EventTypeDAAttestationConfirmed  = "da_attestation_confirmed"
	EventTypeDAAttestationRejected   = "da_attestation_rejected"

	AttributeKeyBlobHash      = "hash"
	AttributeKeyBlobSize      = "size"
	AttributeKeyTopic         = "topic"
	AttributeKeyNonce         = "nonce"
	AttributeKeyTimestamp     = "timestamp"
	AttributeKeyTotalPower    = "total_power"
	AttributeKeyAttestedPower = "attested_power"
	AttributeKeyConfirmed     = "confirmed"
)
