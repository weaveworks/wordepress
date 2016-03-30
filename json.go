package wordepress

type commonDocumentJSON struct {
	Parent    int    `json:"parent"`
	MenuOrder int    `json:"menu_order"`
	Slug      string `json:"slug"`
	Product   string `json:"wpcf-product"`
	Version   string `json:"wpcf-version"`
	Status    string `json:"status"`
}

// POST request body
type DocumentRequestJSON struct {
	commonDocumentJSON
	Title   string `json:"title"`
	Content string `json:"content"`
}

type RenderedJSON struct {
	Rendered string `json:"rendered"`
}

// POST and GET response body
type DocumentJSON struct {
	commonDocumentJSON
	Id      int          `json:"id"`
	Title   RenderedJSON `json:"title"`
	Content RenderedJSON `json:"content"`
}

type MediaDetailsJSON struct {
	File string `json:"file"`
}

type MediaJSON struct {
	MediaDetails MediaDetailsJSON `json:"media_details"`
}
