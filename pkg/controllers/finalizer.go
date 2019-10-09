package controllers

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func hasFinalizer(meta *metav1.ObjectMeta, finalizer string) bool {
	for _, f := range meta.Finalizers {
		if f == finalizer {
			return true
		}
	}

	return false
}

func addFinalizer(meta *metav1.ObjectMeta, finalizer string) {
	if hasFinalizer(meta, finalizer) {
		return
	}

	meta.Finalizers = append(meta.Finalizers, finalizer)
}

func removeFinalizer(meta *metav1.ObjectMeta, finalizer string) {
	finalizers := []string{}

	for _, f := range meta.Finalizers {
		if f == finalizer {
			continue
		}

		finalizers = append(finalizers, finalizer)
	}

	meta.Finalizers = finalizers
}
