package service

import (
	config "github.com/openebs/csi/pkg/config/v1alpha1"
	jiva "github.com/openebs/csi/pkg/service/jiva/v1alpha1"
	cstor "github.com/openebs/csi/pkg/service/v1alpha1"
)

func New(config *config.Config) Interface {
	switch config.CASEngine {
	case "jiva":
		return jiva.New(config)
	default:
		return cstor.New(config)
	}
	return nil
}
