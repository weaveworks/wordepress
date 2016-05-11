package wordepress

import (
	"log"
)

type PublishObserver interface {
	SkippingDocument(local *Document)
	UpdatingDocument(local *Document)
	UploadingDocument(local *Document)
	DeletingDocument(remote *Document)
	SkippingImage(image *Image)
	UploadingImage(image *Image)
}

func NewDryRunPublishObserver() PublishObserver {
	return &dryRunObserver{}
}

func NewWetRunPublishObserver(report *log.Logger) PublishObserver {
	return &wetRunObserver{report}
}

type dryRunObserver struct {
}

func (*dryRunObserver) SkippingDocument(local *Document) {
	log.Printf("Would skip document: %s", local.Slug)
}

func (*dryRunObserver) UpdatingDocument(local *Document) {
	log.Printf("Would update document: %s", local.Slug)
}

func (*dryRunObserver) UploadingDocument(local *Document) {
	log.Printf("Would upload document: %s", local.Slug)
}

func (*dryRunObserver) DeletingDocument(remote *Document) {
	log.Printf("Would delete document: %s", remote.Slug)
}

func (*dryRunObserver) SkippingImage(image *Image) {
	log.Printf("Would skip image: %s%s", image.Hash, image.Extension)
}

func (*dryRunObserver) UploadingImage(image *Image) {
	log.Printf("Would upload image: %s%s", image.Hash, image.Extension)
}

type wetRunObserver struct {
	report *log.Logger
}

func (*wetRunObserver) SkippingDocument(local *Document) {
	log.Printf("Skipping document: %s", local.Slug)
}

func (wro *wetRunObserver) UpdatingDocument(local *Document) {
	log.Printf("Updating document: %s", local.Slug)
	if wro.report != nil {
		wro.report.Printf("Updated %s")
	}
}

func (wro *wetRunObserver) UploadingDocument(local *Document) {
	log.Printf("Uploading document: %s", local.Slug)
	if wro.report != nil {
		wro.report.Printf("Added %s")
	}
}

func (wro *wetRunObserver) DeletingDocument(remote *Document) {
	log.Printf("Deleting document: %s", remote.Slug)
	if wro.report != nil {
		wro.report.Printf("Deleted %s")
	}
}

func (*wetRunObserver) SkippingImage(image *Image) {
	log.Printf("Skipping image: %s%s", image.Hash, image.Extension)
}

func (*wetRunObserver) UploadingImage(image *Image) {
	log.Printf("Uploading image: %s%s", image.Hash, image.Extension)
}
