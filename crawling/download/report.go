package download

import "patreon-crawler/patreon"

type ReportStatus string

const (
	ReportStatusSuccess ReportStatus = "success"
	ReportStatusSkipped ReportStatus = "skipped"
	ReportStatusError   ReportStatus = "error"
)

type ReportItem interface {
	Media() patreon.Media
	Status() ReportStatus
	Message() string
	Error() error
}

type mediaReportItem struct {
	media   patreon.Media
	status  ReportStatus
	message string
	err     error
}

func (m *mediaReportItem) Media() patreon.Media {
	return m.media
}

func (m *mediaReportItem) Status() ReportStatus {
	return m.status
}

func (m *mediaReportItem) Message() string {
	return m.message
}

func (m *mediaReportItem) Error() error {
	return m.err
}

func NewSuccessItem(media patreon.Media) ReportItem {
	return &mediaReportItem{
		media:  media,
		status: ReportStatusSuccess,
	}
}

func NewErrorItem(media patreon.Media, err error) ReportItem {
	return &mediaReportItem{
		media:  media,
		status: ReportStatusError,
		err:    err,
	}
}

func NewSkippedItem(media patreon.Media, message string) ReportItem {
	return &mediaReportItem{
		media:   media,
		message: message,
		status:  ReportStatusSkipped,
	}
}

type ReportStream <-chan ReportItem
