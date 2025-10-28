package download

import "patreon-crawler/patreon"

type ReportItem interface {
	ReportItem()
}

type reportItem struct{}

func (m *reportItem) ReportItem() {}

type ReportSuccessItem struct {
	reportItem
	Media patreon.Media
}

type ReportSkippedItem struct {
	reportItem
	Media  patreon.Media
	Reason string
}

type ReportErrorItem struct {
	reportItem
	Media patreon.Media
	Err   error
}

func NewSuccessItem(media patreon.Media) ReportItem {
	return &ReportSuccessItem{
		Media: media,
	}
}

func NewErrorItem(media patreon.Media, err error) ReportItem {
	return &ReportErrorItem{
		Media: media,
		Err:   err,
	}
}

func NewSkippedItem(media patreon.Media, message string) ReportItem {
	return &ReportSkippedItem{
		Media:  media,
		Reason: message,
	}
}

