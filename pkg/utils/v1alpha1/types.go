package utils

type Volume interface {
	ProvisionVolume() error
	GetVolume() error
	DeleteVolume() error
	IsBound() error
	PatchNodeID() error
}
