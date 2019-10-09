package controller

import (
	apierrs "k8s.io/apimachinery/pkg/api/errors"
)

func IgnoreNotFound(err error) error {
	if apierrs.IsNotFound(err) {
		return nil
	}

	return err
}
