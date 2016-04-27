package wordepress

type Text struct {
	Rendered string `json:"rendered"`
	Raw      string `json:"raw"`
}

type Document struct {
	LocalParent    *Document `json:"-"`
	RemoteDocument *Document `json:"-"`

	ID        int    `json:"id,omitempty"`
	Title     Text   `json:"title"`
	Content   Text   `json:"content"`
	Parent    int    `json:"parent"`
	MenuOrder int    `json:"menu_order"`
	Slug      string `json:"slug"`
	Product   string `json:"wpcf-product"`
	Version   string `json:"wpcf-version"`
	Name      string `json:"wpcf-name"`
	Status    string `json:"status"`
}

type MediaDetails struct {
	File string `json:"file"`
}

type Media struct {
	MediaDetails MediaDetails `json:"media_details"`
}
