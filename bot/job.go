package bot

type JobEntry struct {
	ID          string `json:"id,omitempty"`
	State       string `json:"state,omitempty"`
	DownloadURL string `json:"download_url,omitempty"`
	InfoURL     string `json:"info_url,omitempty"`
	Name        string `json:"name,omitempty"`
}

type JobLog struct {
	JobID     string `json:"job_id"`
	Message   string `json:"message"`
	Initiator string `json:"initiator"`
}
