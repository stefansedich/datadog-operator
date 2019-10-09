package controllers

import (
	apierrs "k8s.io/apimachinery/pkg/api/errors"
)

func ignoreNotFound(err error) error {
	if apierrs.IsNotFound(err) {
		return nil
	}

	return err
}
